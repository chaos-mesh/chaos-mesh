---
id: overview_of_chaos_dashboard
title: Overview of Chaos Dashboard
---

This document will give you an overview of Chaos Dashboard. We will simplify some of the steps we saw in [Run Chaos Experiment](run_chaos_experiment) and guide you to use the visual interface based on our existing features.

We are going to talk about four parts of the dashboard, as below:

- [Status of Experiments](#status-of-experiments)
- [Experiments](#experiments)
  - [Create a New Experiment](#create-a-new-experiment)
  - [Experiment detail](#experiment-detail)
- [Events](#events)
  - [Search Events](#search-events)
- [Archive](#archive)
- [At Last](#at-last)

The pre-setup of this overview is:

```bash
minikube start # You can also use Kind

kubectl create deployment hello-minikube --image=registry.cn-hangzhou.aliyuncs.com/google_containers/echoserver:1.10

curl -sSL https://mirrors.chaos-mesh.org/latest/install.sh | bash

kubectl port-forward -n chaos-testing svc/chaos-dashboard 2333:2333
```

Let's get started.

## Status of Experiments

After open <http://localhost:2333>, if you haven't created some experiments, you will see a page like this:

![Overview](/img/chaos-dashboard/overview.png)

Currently, we place the status of experiments on this page. The left is the total number of experiments and the right half of the page is our abstraction of the entire status: `RUNNING`, `PAUSED`, `FAILED`, `WAITING`, `FINISHED`.

It is worth noting that the `WAITING` is not as intuitive as other status. So, what does it mean?

It will change from 0 to 1, 2, 3... when you have a scheduled job **temporarily finished**. Since the scheduled job will continue to run in future intervals, so we define a specific status `WAITING` for this situation.

## Experiments

Creating experiments is an important part of the dashboard. Before we initially wanted to develop the dashboard, **conveniently creating experiments is a feature that we attach great importance to**.
In the following content, we will show you how to create a `NetworkChaos` experiment as an example.

### Create a New Experiment

Below is the creating page:

![Create Experiment](/img/chaos-dashboard/create-experiment.png)

The left half is the form of create experiment. You have to walk through four steps to create an experiment.

- Basic `(Basic info of the experiment, same as kubenetes object metadata)`
- Scope `(Limit the scope of the experiment)`
- Target `(Choose the chaos you want)`
- Schedule `(How to run the experiment you defined)`

We will not teach you by hand, most of the creation operations are obvious. The important thing is we hope you can understand some interface design to help you better use the dashboard to create experiments.

First, type `network-delay` in the name field, you can ignore the other fields at this moment. Then click next.

#### Scope

In the second step, we call it scope. In this example, I create a simple `deployment` named `hello-minikube`. So this page is what you will see:

![Create Experiment Scope](/img/chaos-dashboard/create-experiment-scope.png)

If you do not fill in any fields, the scope will be applied to all pods by default.(In the `Affected Pods Preview` at the bottom, It's not just a preview, you can check or uncheck pods to limit experiment scope more precisely)

Also, like the `label selectors`, you can select the existing labels in the k8s system, all selectors will filter out the system objects with the same labels.

Because we only have a single pods at this example, we can click next directly. For more detail about `Scope`, you can view [Define the Scope of Chaos Experiment](experiment_scope).

#### Target

In this overview we will select a `NetworkChaos` with `delay` `10ms latency`, that's very simple. Just select the tab of `Network`, choose action to Delay, then fill `Lantency` to `10ms`.

Below is the result:

![Create Experiment Target](/img/chaos-dashboard/create-experiment-target.png)

Click the next button you can set the experiment schedule by unchecking the `immediate` radio button, fill two fields `cron` and `duration` you want, then all steps are finished. You can submit this experiment right now.

As you can see, no need to write spec by YAML, no need to apply by hand, the chaos dashboard can help you quickly try different chaos.

**But what if you have already defined some experiments by YAML files, and what if you want to use them by dashboard?**

**Or, you have already created or archived some experiments and you want to duplicate or recover the experiment?**

The right half is for you. Currently, we have these three options:

- Load From Existing Experiments
- Load From Existing Archives
- Load From YAML File

You can try these after the overview.

### Experiment detail

Click the detail button of experiment in the `Experiments` page, we can view our previous created:

![Experiments](/img/chaos-dashboard/experiments.png)

![Experiment Detail](/img/chaos-dashboard/experiment-detail.png)

You can exec `Archive`, `Pause (Start)`, `Update` operation to the experiment, the detail page also shows the timeline and events with this experiments. You can explore more after this overview.

## Events

![Events](/img/chaos-dashboard/events.png)

The events page shows all experiments' events, also with the timeline panel.

### Search Events

If you want to quickly locate some events. You can search at the top right of the events table.

We also define some useful patterns to help you search events:

- `namespace:default` will search events with namespace default.
- `kind:NetworkChaos` will search events with kind NetworkChaos, you can also type kind:net, it will also work because the search is fuzzy.
- `pod:xxx` will search events with pod name xxx.
- `uuid:xxx` will search events with uuid xxx.

All of these patterns can be appended with `experiment name`, for example, you can type `kind:PodChaos xxx` to filter events which name starts with `xxx` with kind PodChaos.

## Archive

To introduce this section, I have already archived the experiment we just defined.

![Archive](/img/chaos-dashboard/archive.png)

We define the content of the archive page as `report`. All of your experiments will not be lost, their final form is archive. If you want to stop a experiment, just archive it.

And you can recover(duplicate) the archive anytime.

## At Last

The overview is ended.

We hope this overview can give you a general understanding of the dashboard.

For now, we are still constantly improving the entire dashboard. Through this overview, I think we have conveyed some important concepts of the dashboard to you, even though the entire dashboard interface may have major changes in the future, we think you can also adapt to it quickly.

If you have any questions or suggestions about Chaos Dashboard, please submit [issues](https://github.com/chaos-mesh/chaos-mesh/issues) to our repo. All of our maintainers are glad to help you.~
