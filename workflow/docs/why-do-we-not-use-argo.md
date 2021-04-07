# Why do we not use Argo
<!-- TOC -->

- [Why do we not use Argo](#why-do-we-not-use-argo)
  - [What's the advantage if we use argo](#whats-the-advantage-if-we-use-argo)
  - [Argo workflow is not the best way to describe declarative chaos experiments](#argo-workflow-is-not-the-best-way-to-describe-declarative-chaos-experiments)
  - [We could do more things if we have a customized workflow](#we-could-do-more-things-if-we-have-a-customized-workflow)
  - [Why the other chaos platform choose to use Argo, but we do not prefer to use it](#why-the-other-chaos-platform-choose-to-use-argo-but-we-do-not-prefer-to-use-it)
  - [Other misc](#other-misc)

<!-- /TOC -->

It has been a long time that we decide to build up another workflow engine. I believe that we have enough reasons that push us to build a new workflow engine.

But when other contributors ask me, "Why not use Argo"? I could not give them convincing words in time. That's why I decide to write this doc.

## What's the advantage if we use argo

Argo is an amazing workflow engine for common purpose. Argo provides "sequential/parallel execution", "conditional branches", "DAG"; and workloads like "script", "container", "resource".

I think the most advantage is reducing our jobs, we do not need to re-implement another workflow engine. We could build a workflow by writing some scripts with `kubectl`.

## Argo workflow is not the best way to describe declarative chaos experiments

Argo supports "script", "container" and "resource" workloads, it could integrate with Chaos Mesh to apply/recover chaos experiments. But this way to apply chaos experiments is not declarative anymore, it turns to imperative.

Argo is so powerful and it's impossible to ask users always for following the best practices. Users could write complex scripts for determining the situation, then `kubectl apply` or `kubectl patch` something, once a command make unexpected side-effects, it will let the whole system unstable.

We define the workflow that could describe chaos experiments in the declarative, changing chaos experiments is just like "from expected status A turning into expected status B", users should not considering about "remove some chaos CR object" or "create new chaos CR object". We also allow users (only) to execute their scripts in `Task` node, it will not affect the changing of chaos experiment states.

## We could do more things if we have a customized workflow

Based on our own workflow, it's easier to provide more useful features like:

- more detailed status perceive and retry/failover
- chaos experiment report
- better UI

## Why the other chaos platform choose to use Argo, but we do not prefer to use it

Actually, for example, litmus chaos is not strictly declarative chaos experiments, its `ChaosExperiments` contains scripts that how to inject chaos. During executing, a `ChaosEngine` will instantiate it into a `Job`, then running those scripts. If something unexpected happened, such as the pod was killed by OOM killer, the dirty thing is just left here without recovering anymore. In other words, it just a wrapper of `Job`, with more effective selectors and conditions. So it is suitable for Argo, it is imperative commands from start to end.

In other words, we do not use Argo, which does not mean we reject Argo. We are trying to defining something that could schedule chaos, and independent with Argo.

## Other misc

Applications based on kubernetes have an issue: it's nearly not possible to deploy different instances with different versions in one cluster. If Chaos Mesh depends on argo version X, and the user already installed argo version Y, it might be incompatible, there is no way to resolve it.

Chaos Mesh has the plans for more situation not only kubernetes, just like CaaS(Chaos engineering as a Service). If one day we need to migrate workflow into a non-kubernetes cluster, I think our own workflow is easier to do.
