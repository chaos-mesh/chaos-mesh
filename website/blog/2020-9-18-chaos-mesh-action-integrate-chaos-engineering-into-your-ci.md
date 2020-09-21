---
slug: /chaos-mesh-action-integrate-chaos-engineering-into -your-ci
title: 'chaos-mesh-action: Integrate Chaos Engineering into Your CI'
author: Xiang Wang
author_title: Contributor of Chaos Mesh
author_url: https://github.com/WangXiangUSTC
author_image_url: https://avatars3.githubusercontent.com/u/5793595?v=4
image: /img/automated_testing_framework.png
tags: [Chaos Mesh, Chaos Engineering, GitHub Action, CI]
---

![chaos-mesh-action - Integrate-Chaos-Engineering-into-Your-CI](/img/chaos-mesh-action.png)

[Chaos Mesh](https://chaos-mesh.org) is a cloud-native chaos testing platform that orchestrates chaos in Kubernetes environments. While it’s well received in the community with its rich fault injection types and easy-to-use dashboard, it was difficult  to use Chaos Mesh with end-to-end testing or the continuous integration (CI) process. As a result, problems introduced during system development could not be discovered before the release.

In this article, I will share how we use chaos-mesh-action, a GitHub action to integrate Chaos Mesh into the CI process.
<!--truncate-->

chaos-mesh-action is available on [GitHub market](https://github.com/marketplace/actions/chaos-mesh), and the source code is on [GitHub](https://github.com/chaos-mesh/chaos-mesh-action).

## Design of chaos-mesh-action

[GitHub Action](https://docs.github.com/en/actions) is a CI/CD feature natively supported by GitHub, through which we can easily build automated and customized software development workflows in the GitHub repository. 

Combined with GitHub actions, Chaos Mesh can be more easily integrated into the daily development and testing of the system, thus guaranteeing that each code submission on GitHub is bug-free and won’t damage existing code. The following figure shows chaos-mesh-action integrated into the CI workflow:

![chaos-mesh-action integrate in the CI workflow](/img/chaos-mesh-action-integrate-in-the-ci-workflow.png)

## Using chaos-mesh-action in GitHub workflow

[chaos-mesh-action](https://github.com/marketplace/actions/chaos-mesh) works in Github workflows. A GitHub workflow is a configurable automated process that you can set up in your repository to build, test, package, release, or deploy any GitHub project. To integrate Chaos Mesh in your CI, do the following:

1. Design a workflow.
2. Create a workflow.
3. Run the workflow.

### Design a workflow

Before you design a workflow, you must consider the following issues:

* What functions are we going to test in this workflow?
* What types of faults will we inject? 
* How do we verify the correctness of the system?

As an example, let’s design a simple test workflow that includes the following steps: 

1. Create two Pods in a Kubernetes cluster.
2. Ping one pod from the other. 
3. Use Chaos Mesh to inject network delay chaos and test whether the ping command is affected.

### Create the workflow

After you design the workflow, the next step is to create it. 

1. Navigate to the GitHub repository that contains the software you want to test.
2. To start creating a workflow, click **Actions**, and then click the **"New workflow**" button:

![Creating a workflow](/img/creating-a-workflow.png)

A workflow is essentially the configuration of jobs that take place sequentially and automatically. Note that the jobs are configured in a single file. For better illustration, we split the script into different job groups as shown below: 

*   Set the workflow name and trigger rules.

    This job names the workflow "Chaos.” When the code is pushed to the master branch or a pull request is submitted to the master branch, this workflow is triggered.

    ```yaml
    name: Chaos

    on:
     push:
       branches:
         - master
     pull_request:
       branches:
         - master
    ```

*   Install the CI-related environment.

    This configuration specifies the operating system (Ubuntu), and that it uses [helm/kind-action](https://github.com/marketplace/actions/kind-cluster) to create a Kind cluster. Then, it outputs related information about the cluster. Finally, it checks out the GitHub repository for the workflow to access. 

    ```yaml
    jobs:
     build:
       runs-on: ubuntu-latest
       steps:

       - name: Creating kind cluster
         uses: helm/kind-action@v1.0.0-rc.1

       - name: Print cluster information
         run: |
           kubectl config view
           kubectl cluster-info
           kubectl get nodes
           kubectl get pods -n kube-system
           helm version
           kubectl version

       - uses: actions/checkout@v2
    ```

*   Deploy the application.

    In our example, this job deploys an application that creates two Kubernetes Pods.

    ```yaml
    - name: Deploy an application
         run: |
           kubectl apply -f https://raw.githubusercontent.com/chaos-mesh/apps/master/ping/busybox-statefulset.yaml
    ```

*   Inject chaos with chaos-mesh-action.

    ```yaml
    - name: Run chaos mesh action
        uses: chaos-mesh/chaos-mesh-action@xiang/refine_script
        env:
          CFG_BASE64: YXBpVmVyc2lvbjogY2hhb3MtbWVzaC5vcmcvdjFhbHBoYTEKa2luZDogTmV0d29ya0NoYW9zCm1ldGFkYXRhOgogIG5hbWU6IG5ldHdvcmstZGVsYXkKICBuYW1lc3BhY2U6IGJ1c3lib3gKc3BlYzoKICBhY3Rpb246IGRlbGF5ICMgdGhlIHNwZWNpZmljIGNoYW9zIGFjdGlvbiB0byBpbmplY3QKICBtb2RlOiBhbGwKICBzZWxlY3RvcjoKICAgIHBvZHM6CiAgICAgIGJ1c3lib3g6CiAgICAgICAgLSBidXN5Ym94LTAKICBkZWxheToKICAgIGxhdGVuY3k6ICIxMG1zIgogIGR1cmF0aW9uOiAiNXMiCiAgc2NoZWR1bGVyOgogICAgY3JvbjogIkBldmVyeSAxMHMiCiAgZGlyZWN0aW9uOiB0bwogIHRhcmdldDoKICAgIHNlbGVjdG9yOgogICAgICBwb2RzOgogICAgICAgIGJ1c3lib3g6CiAgICAgICAgICAtIGJ1c3lib3gtMQogICAgbW9kZTogYWxsCg==
    ```

    With chaos-mesh-action, the installation of Chaos Mesh and the injection of chaos complete automatically. You simply need to prepare the chaos configuration that you intend to use to get its Base64 representation. Here, we want to inject network delay chaos into the Pods, so we use the original chaos configuration as follows:

    ```yaml
    apiVersion: chaos-mesh.org/v1alpha1
    kind: NetworkChaos
    metadata:
     name: network-delay
     namespace: busybox
    spec:
     action: delay # the specific chaos action to inject
     mode: all
     selector:
       pods:
         busybox:
           - busybox-0
     delay:
       latency: "10ms"
     duration: "5s"
     scheduler:
       cron: "@every 10s"
     direction: to
     target:
       selector:
         pods:
           busybox:
             - busybox-1
       mode: all
    ```

    You can obtain the Base64 value of the above chaos configuration file using the following command:

    ```shell
    $ base64 chaos.yaml
    ```

*   Verify the system correctness.

    In this job,  the workflow pings one Pod from the other and observes the changes in network delay.

    ```yaml
    - name: Verify
         run: |
           echo "do some verification"
           kubectl exec busybox-0 -it -n busybox -- ping -c 30 busybox-1.busybox.busybox.svc
    ```

### Run the workflow

Now that the workflow is configured, we can trigger it by submitting a pull request to the master branch. When the workflow completes, the verification job outputs of the results that look similar to the following:

```shell
do some verification
Unable to use a TTY - input is not a terminal or the right kind of file
PING busybox-1.busybox.busybox.svc (10.244.0.6): 56 data bytes
64 bytes from 10.244.0.6: seq=0 ttl=63 time=0.069 ms
64 bytes from 10.244.0.6: seq=1 ttl=63 time=10.136 ms
64 bytes from 10.244.0.6: seq=2 ttl=63 time=10.192 ms
64 bytes from 10.244.0.6: seq=3 ttl=63 time=10.129 ms
64 bytes from 10.244.0.6: seq=4 ttl=63 time=10.120 ms
64 bytes from 10.244.0.6: seq=5 ttl=63 time=0.070 ms
64 bytes from 10.244.0.6: seq=6 ttl=63 time=0.073 ms
64 bytes from 10.244.0.6: seq=7 ttl=63 time=0.111 ms
64 bytes from 10.244.0.6: seq=8 ttl=63 time=0.070 ms
64 bytes from 10.244.0.6: seq=9 ttl=63 time=0.077 ms
……
```

The output indicates a regular series of 10-millisecond delays that last about 5 seconds each. This is consistent with the chaos configuration we injected into chaos-mesh-action.  

## Current status and next steps

At present, we have applied chaos-mesh-action to the [TiDB Operator](https://github.com/pingcap/tidb-operator) project. The workflow is injected with the Pod chaos to verify the restart function of the specified instances of the operator. The purpose is to ensure that tidb-operator can work normally when the pods of the operator are randomly deleted by the injected faults. You can view the [TiDB Operator page](https://github.com/pingcap/tidb-operator/actions?query=workflow%3Achaos) for more details.

In the future, we plan to apply chaos-mesh-action to more tests to ensure the stability of TiDB and related components. You are welcome to create your own workflow using chaos-mesh-action.

If you find a bug or think something is missing, feel free to file an issue, open a pull request (PR), or join us on the [#project-chaos-mesh](https://join.slack.com/t/cloud-native/shared_invite/zt-fyy3b8up-qHeDNVqbz1j8HDY6g1cY4w) channel in the [CNCF](https://www.cncf.io/) slack workspace. 
