# Internal Design of Chaos Mesh Workflow

<!-- TOC -->

- [Internal Design of Chaos Mesh Workflow](#internal-design-of-chaos-mesh-workflow)
  - [Overview](#overview)
  - [Core Concepts](#core-concepts)
  - [Workflow Engine](#workflow-engine)
    - [Trigger](#trigger)
    - [Scheduler](#scheduler)
    - [Actor and Playground](#actor-and-playground)
    - [WorkflowManager: rest glue codes](#workflowmanager-rest-glue-codes)
    - [States for Node](#states-for-node)
    - [Node StateMachine](#node-statemachine)
      - [Serial](#serial)
      - [Suspend](#suspend)
      - [Chaos](#chaos)
    - [States for Workflow](#states-for-workflow)
    - [Event](#event)
    - [Others](#others)
  - [Implement for Various Template](#implement-for-various-template)
    - [Suspend](#suspend-1)
    - [Parallel/Serial](#parallelserial)
    - [Chaos](#chaos-1)
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

An instance of Chaos Mesh Workflow is basically like a tree:

- each action is described by a node;
- a root node represents for entry-point;
- nodes with certain types could have children nodes;
- depends on the type of this node, children nodes execute serial or parallel;
- node with certain type could select which children nodes to run(conditional branch);

So Chaos Mesh Workflow is mostly like:

- Argo

It doesn't like:

- BPMN

> Do we really need a tree? How about a pipeline?

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
type Trigger interface {
	TriggerName() string
	Acquire(ctx context.Context) (event Event, canceled bool, err error)
}
```

For easy implementation, we define a `OperableTrigger`, it works like a queue:

```go
type OperableTrigger interface {
	Trigger
	Notify(event Event) error
	NotifyDelay(event Event, delay time.Duration) error
}
```

- `Notify` as `Enqueue`
- `Acquire` as `Dequeue`

`OperableTrigger` is the bridge between workflow engine and other components.

> For implementation with kubernetes, we should make a **Trigger** which is a reconciler in the controller-manager.

### Scheduler

Scheduler should pick out one or more templates that should be executed next. Scheduler should:

- parse the whole Workflow(or other components tell scheduler the hierarchy of workflow, I'm not sure for that)
- pick next template(s) to run
- evaluate expressions for conditional branch(only for `Task` Template, which contains conditional branches)

We should implement at least 4 types of Scheduler:

- `EntryScheduler`, it's the first scheduler for each workflow, it only picks the `entry` template.
- `SerialScheduler`, it should pick the next **one** template from children templates in SerialTemplate.
- `ParallelScheduler`, it should pick the all templates from children templates in ParallelTemplate.
- `TaskScheduler`, it is based on `ParallelScheduler`, but contains logic about evaluation expressions for conditional branches.

```go
type Scheduler interface{
  ScheduleNext(ctx context.Context) (nextTemplates []template.Template, parentNodeName string, err error)
}
```

### Actor and Playground

> - Nothing happens until the Actor works.
> - The status of Workflow does NOT present the result for the Actor. (It means we could not fetch the result of one "Actor" from kubernetes.) Because the "real world" is so complicated to describe, we want the result of "Actor" do not "pollution" our schedule objects.
> - While an "Actor" failed, the whole workflow instance is trapped into failed.

The Actor will do all "dirty" effects, for example: creating Chaos Resource, running user-defined tasks, and so on.

```go
type Actor interface {
  PlayOn(pg Playground) error
}

type Playground interface {
  CreateNetworkChaos(networkChaos chaosmeshv1alph1.NetworkChaos) error
  DeleteNetworkChaos(namespace, name string) error
  CreatePodChaos(podChaos chaosmeshv1alph1.PodChaos) error
  DeletePodChaos(namespace, name string) error

  // Other ad-hoc methods
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

### States for Node

There are 8 available phase of one Node:

- Init
- WaitingForSchedule
- WaitingForChild
- Running
- Evaluating
- Holding
- Succeed
- Failed

**Init** means is the default phase when **Node** just created, means this node is just created, did not effect real world yet.

**WaitingForSchedule** is only available on Serial, Parallel and Task, it means this Node is idle and safe for next scheduling; It is presents:

- For Chaos, Suspend, Task, Parallel:
  - This Node did not make "Effect" yet.
- For Serial:
  - This Node did not create child node yet;
  - Or previous child node just succeed;

**WaitingForChild** is only available on Serial, Parallel and Task, it means at least 1 child node is in **Running** state.

**Running** is available on type Chaos, Suspend, Task; It means an **Actor** is doing dirty work for this node. For Chaos, both of "create Chaos CRD resource" and "delete Chaos CRD resource" are presented as **Running**.

**Evaluating** is only available on Task, it means Task is collecting result of user's pod, then picks templates to execute.

> Question: Should we split it?

**Holding** is available on Chaos, Suspend; It means current node is waiting for next action. For example: a Chaos node which in **Holding** is waiting for the end of this ChaosExperiments, then delete it.

**Succeed** means this node is completed.

**Failed** means this node failed. One **Failed** node will cause the whole Workflow fall into **Failed**.

Examples:

A **NetworkChaos** Node: **Init** -> **Running** -> **Holding** -> **Running** -> **Succeed**

A **Suspend** Node: **Init** -> **Holding** -> **Succeed**

A **Serial** Node(which contains 3 children): **Init** -> **WaitingForSchedule** -> **WaitingForChild** -> **WaitingForSchedule** -> **WaitingForChild** -> **WaitingForSchedule** -> **WaitingForChild** -> **Succeed**

A **Parallel** Node(which contains 3 children): **Init** -> **WaitingForSchedule** -> **WaitingForChild** -> **WaitingForChild** -> **WaitingForChild** -> **Succeed**

A **Task** Node: **Init** -> **Running** -> **WaitingForChild** -> **Evaluating** -> **WaitingForSchedule** -> **WaitingForChild** -> **Succeed**

### Node StateMachine

#### Serial

Available phase: Init, WaitingForSchedule, WaitingForChild, Succeed, Failed.

| Event \ Current Phase   | Init                                                                                          | WaitingForSchedule                                                                                                                | WaitingForChild                                                                                                                                                                                                      | Succeed | Failed |
| :---------------------- | :-------------------------------------------------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :------ | :----- |
| NodeCreated             | Change phase to `WaitingForSchedule`, then notify itself with event `NodePickChildToSchedule` | -                                                                                                                                 | -                                                                                                                                                                                                                    | -       | -      |
| NodePickChildToSchedule | -                                                                                             | Create children nodes, then change phase to `WaitingForChild`, then notify all children nodes one by one with event `NodeCreated` | -                                                                                                                                                                                                                    | -       | -      |
| ChildNodeSucceed        | -                                                                                             | -                                                                                                                                 | Check children nodes / Change to phase `WaitingForSchedule` then notify itself with event `NodePickChildToSchedule` / or Change to phase `Succeed` then notify parent node (if exists) with event `ChildNodeSucceed` | -       | -      |
| ChildNodeFailed         | -                                                                                             | -                                                                                                                                 | Change phase to `Failed` then notify parent node (if exists) with event `ChildNodeFailed`                                                                                                                            | -       | -      |

#### Suspend

Available phase: Init, Holding, Succeed.

| Event \ Current Phase | Init                                                                                             | Holding                                                                                     | Succeed |
| :-------------------- | :----------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------ | :------ |
| NodeCreated           | Change phase to `Holding`, then delay notify itself with event `NodeHoldingAwake` and `duration` | -                                                                                           | -       |
| NodeHoldingAwake      | -                                                                                                | Change to phase `Succeed` then notify parent node (if exists) with event `ChildNodeSucceed` | -       |

#### Chaos

Available phase: Init, Running, Holding, Succeed, Failed.

| Event \ Current Phase | Init                                                                                                    | Running                                                                                          | Holding                                                                                                 | Succeed | Failed |
| :-------------------- | :------------------------------------------------------------------------------------------------------ | :----------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------ | :------ | :----- |
| NodeCreated           | Change phase to `Running`, create CRD about chaos by Actor, notify itself with `NodeChaosInjectSucceed` | -                                                                                                | -                                                                                                       | -       | -      |
| NodeChaosInjected     | -                                                                                                       | Change phase to `Holding`, then delay notify itself with event `NodeHoldingAwake` and `duration` | -                                                                                                       | -       | -      |
| NodeHoldingAwake      | -                                                                                                       | -                                                                                                | Change to phase `Running` ,remove CRD about chaos by Actor, notify itself with event `NodeChaosCleaned` | -       | -      |
| NodeChaosCleaned      | -                                                                                                       | Change phase to `Succeed`, notify parent node(if exists) with `ChildNodeSucceed`                 | -                                                                                                       | -       | -      |

### States for Workflow

There are 4 phase of one Workflow:

- Init
- Running
- Succeed
- Failed

**Init** is the default phase when **Workflow** just created. It means there are no Node is in **Running** or **Holding** state, and it's safe for scheduling next operation.

**Running** means at least 1 node is in **Running**/**Holding**/**WaitingForChild**/**WaitingForSchedule** state.

**Succeed** means all nodes is in **Succeed** node and no Template could schedule anymore.

**Failed** means at least one Node is in **Failed** state.

> Question: Should we interpret other node when Workflow fall into **Failed**?
> I think we should do that.

### Event

> Unimplemented now.

Inspired by kubernetes events, it should emit an event when something happens:

- chaos or other resource created
- make decision about conditional branches
- unexpected errors

### Others

- Persistent Repo or just "repo", fetches and updates status for Workflow and Node.

Detail about instating template:

After calling `Scheduler.ScheduleNext()`, it will return

- which template should be instated
- the parent node of template's new node instant

1. update the state of parent node to `WaitingForChild`
1. instant a new node of this template

Detail about Node

## Implement for Various Template

For each Actor, its operation is "instantly", which means one Chaos which contains "duration" field will be implemented as 2 operations: create, then delete.

### Suspend

Please checkout the statemachine.

### Parallel/Serial

Please checkout the statemachine.

### Chaos

Chaos Experiments is implemented by both **Trigger** and **Actor**, every Chaos has a required field: duration, so the story about Chaos Experiments look like:

- **Actor** runs, create a CRD resource about Chaos Experiments
- calling **Trigger**, requeue this workflow after `duration`
- waiting for `duration`
- another **Actor** runs, delete Chaos CRD

It will also watch on Chaos Object, if Chaos status is not expected, this template will fail.

> Another **Trigger** do the watch things.

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

- Support DAG or not? No. I think it's not enough necessary.
- Assertion SDK / WebAPI for Task
- Event-driven & Webhook. Yes! I have updated the design.
- Split controller-manager and workflow-engine as two binary
- Should `Node` be a standalone CRD? No. Just write the codes at first.
