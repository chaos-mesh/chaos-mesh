# Internal Design of Chaos Mesh Workflow

<!-- TOC -->

- [Internal Design of Chaos Mesh Workflow](#internal-design-of-chaos-mesh-workflow)
  - [Overview](#overview)
  - [Core Concepts](#core-concepts)
  - [Workflow Engine](#workflow-engine)
    - [Trigger](#trigger)
    - [Scheduler](#scheduler)
    - [Actor](#actor)
    - [WorkflowManager: rest glue codes](#workflowmanager-rest-glue-codes)
    - [Others](#others)
  - [Implement for Various Template](#implement-for-various-template)
    - [Suspend](#suspend)
    - [Parallel/Serial](#parallelserial)
    - [Chaos](#chaos)
    - [Task](#task)
      - [Downward API](#downward-api)
  - [Performance limitation](#performance-limitation)
  - [API](#api)
    - [Kubernetes CRD](#kubernetes-crd)
  - [Schema-less Serialization](#schema-less-serialization)
    - [Raw](#raw)
    - [Unstructured](#unstructured)
    - [v1.Json](#v1json)
    - [mapstructure](#mapstructure)
  - [Observability](#observability)
    - [Logs](#logs)
    - [Metrics](#metrics)
    - [Tracing](#tracing)
  - [Compare with v1.0.2](#compare-with-v102)
    - [Breaking Changes](#breaking-changes)
  - [Unresolved Problems](#unresolved-problems)

<!-- /TOC -->

> WIP: This document is still under development, everything is not stable, might change frequently.

## Overview

As the RFC of Chaos Mesh Workflow is stable, we create this document as the reference for implementation.

We will not define the certain struct here, but we present pseudo-code for interface and main logic.

An instance of  Chaos Mesh Workflow is basically like a tree:

- each action is described by a node;
- a root node represents for entry-point;
- nodes with certain types could have children nodes;
- depends on the type of this node, children nodes execute serial or parallel;
- node with certain type could select which children nodes to run(conditional branch);

So Chaos Mesh Workflow is mostly like:

- Argo

It doesn't like:

- BPMN

## Core Concepts

**Workflow**: a resource that defines the orchestration of chaos experiments.

**Template**: the node in **Workflow** "tree", represents for operation. There are various operations: like "create chaos experiments", "wait for 5 minutes", "create a serial job with children templates", etc.

**Node**: a running/completed step within a workflow, it represents the status; We could think of it as an instance rendered from a **Template**.

**CronWorkflow**: a way to execute workflow with a schedule. The relationship between **CronWorkflow** and **Workflow** is like what between [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) and [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/).

## Workflow Engine

For driving workflow, here must a "controller" thing, we call it "Workflow Engine". Workflow Engine is the core of Chaos Mesh Workflow.

### Trigger

Trigger should:

- watch on the "real world", send events to other components.
- accept request for "waiting for a duration, then trigger event"
- apply rate-limiting on triggered events

```go
type Trigger interface{
  AcquireNext(ctx context.Context) (WorkflowKey, error)
  ReQueue(key WorkflowKey, duration time.Duration) error
}
```

**Trigger** is most like a multiplexer, it could react with various type of events, like:

- resources changes
- delay trigger
- ...

> For implementation with kubernetes, **Trigger** is controller/reconciler.

### Scheduler

Scheduler should:

- parse the whole Workflow
- pick next templates to run
- evaluate expressions for conditional branch

Every workflow instance has its Scheduler, and Scheduler must be rebuild from Workflow status.

```go
type Scheduler interface{
  ScheduleNext(ctx context.Context) ([]Template, error)
}
```

### Actor

> - Nothing happens until the Actor works.
> - The status of Workflow does NOT present the result for the Actor. (It means we could not fetch the result of one "Actor" from kubernetes.) Because the "real world" is so complicated to describe, we want the result of "Actor" do not "pollution" our schedule objects.
> - While an "Actor" failed, the whole workflow instance is trapped into failed.

The Actor will do all "dirty" effects, for example: creating Chaos Resource, running user-defined tasks, and so on.

```go
type Actor interface {
  Apply(ctx context.Context) error
}
```

### WorkflowManager: rest glue codes

We need a manager who makes these components work together, and controls the main-loop, and update the status of the workflow, and max-concurrency, and so-on.

The main loop looks like:

```go
func loop(){
  for {
    workflowKey, err := trigger.AcquireNext()
    scheduler, err := fetchSchedulerFor(workflowKey)

    go func(){
      templates, err := scheduler.scheduleNext()
      for item := range templates {
        go func (){
          // TODO: update node status
          actor := fetchActor(item)
          err := actor.Apply()
          // TODO: update node status
          if item.NeedReQueue() {
            // TODO: manually ask for next trigger
          }
        }()
      }
    }
  }
}
```

### Others

- Persistent Repo or just "repo", fetches and updates status for Workflow and Node.

## Implement for Various Template

For each Actor, its operation is "instantly", which means one Chaos which contains "duration" field will be implemented as 2 operations: create, then delete.

### Suspend

As **Trigger** supporting `Requeue()`,  **Suspend** is quite simple, just write the time to wake up into status then requeue could implement operation based on time.

It's also the basement for implementing other types of templates.

### Parallel/Serial

Parallel and Serial is implemented by composite pattern, it's implemented in **Scheduler**; When someone asks for `ScheduleNext()`, it will return all templates to execute: it returns only one template with **Serial**, returns all children templates with **Parallel**, evaluates expressions and returns matched templates with **Task**.

Notice here is a field called `deadline` for **Parallel**/**Serial** template; If `deadline` is set, this template will fail if the sub-template is not finished when `deadline` exceeds.

### Chaos

Chaos Experiments is implemented by both **Trigger** and **Actor**, every Chaos has a required field: duration, so the story about Chaos Experiments look like:

- **Actor** runs, create a CRD resource about Chaos Experiments
- calling **Trigger**, requeue this workflow after `duration`
- waiting for `duration`
- another **Actor** runs, delete Chaos CRD

It will also watch on Chaos Object, if Chaos status is not expected, this template will fail.

### Task

Task could treat as three parts:

- run customized job in a Pod
- evaluate expressions
- same thing as Parallel

We could reuse `PodTemplateSpec` to claim how to create the pod. We should decorate the pod for implementing "Downward API", we will talk about it later.

After pod created, a specific **Trigger** watches on the pod, when the pod stopped(succeed or failed), trigger **Scheduler** for next operation;

#### Downward API

For supporting users could write their codes for judgment based on current workflow status, we provide a mechanism also called "Downward API". Likes Kubernetes Downward API, it also injects status as a file into the pod. The feature is enabled by default, there are no configurations for it currently.

We will create a ConfigMap Resource, the content of it is Workflow CRD object, contains spec and status as JSON. When creating the Pod, it will be mounted on `/var/run/chaos-mesh/`, so the user could access this file like `cat /var/run/chaos-mesh/workflow.json`.

> The content of Downward API is a static **snapshot** about workflow object, it will **NOT** change when Workflow object in kubernetes updates.

## Performance limitation

We want Chaos Mesh Workflow could drive these things at the same time:

- <= 100 workflows
- each workflow contains <= 100 templates
- "scheduling next operation" should be completed in 1 seconde

> This is only about workflow, it doesn't contains performance about Chaos Experiments

## API

### Kubernetes CRD

## Schema-less Serialization

### Raw

### Unstructured

### v1.Json

### mapstructure

## Observability

### Logs

### Metrics

### Tracing

## Compare with v1.0.2

### Breaking Changes

## Unresolved Problems

- Support DAG or not
- Assertion SDK / WebAPI for Task
- Event-driven & Webhook
- Split controller-manager and workflow-engine as two binary
