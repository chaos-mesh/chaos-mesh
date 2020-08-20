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
- [Archives](#archives)

Let's get started.

## Status of Experiments

After open <http://localhost:2333>, if you haven't created some experiments, you will see a page like this:

![Overview](/img/chaos-dashboard/overview.png)

Currently, we place the status of experiments on this page. The left is the total number of experiments and the right half of the page is our abstraction of the entire status: `RUNNING`, `PAUSED`, `FAILED`, `WAITING`, `FINISHED`.

It is worth noting that the `WAITING` is not as intuitive as other status. So, what does it mean?

It will change from 0 to 1,2,3... when you have a scheduled job **temporarily finished**. Since the scheduled job will continue to run in future intervals, so we define a specific status `WAITING` for this situation.

## Experiments

Creating experiments is an important part of dashboard. Before we initially wanted to develop the dashboard, **conveniently creating experiments is a feature that we attach great importance to**.
In the following content, we will show you how to create a `NetworkChaos` experiment as an example.

### Create a New Experiment

Below is the creating page:

![Create Experiment](/img/chaos-dashboard/create-experiment.png)

The left half is the form of create experiment. You have to walk through four steps to create an experiment.

- Basic `(Basic info of the experiment, same as kubenetes object metadata)`
- Scope `(Limit the scope of the experiment)`
- Target `(Choose the chaos you want)`
- Schedule `(How to run the experiment you defined)`

We will not teach you by hand, Most of the creation operations are obvious. The important thing is we hope you can understand some interface design to help you better use the dashboard to create experiments.

<!-- TODO -->

The right half has three sections, which are:

- Load From Existing Experiments
- Load From Existing Archives
- Load From YAML File

Let's describe the right half. Usually, **you already have already defined some experiments by YAML files, and what if you want to use them by dashboard?**

**Or, you have already created or archived some experiments and you want to duplicate or recover the experiment?**

### Experiment detail

## Events

### Search Events

## Archives
