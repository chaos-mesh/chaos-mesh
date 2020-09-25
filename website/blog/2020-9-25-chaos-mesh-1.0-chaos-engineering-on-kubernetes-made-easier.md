---
slug: /chaos-mesh-1.0-chaos-engineering-on-kubernetes-made-easier
title: 'Chaos Mesh 1.0: Chaos Engineering on Kubernetes Made Easier'
author: Chaos Mesh Maintainers
author_url: https://github.com/chaos-mesh
author_image_url: https://avatars1.githubusercontent.com/u/59082378?v=4
image: /img/chaos-mesh-1.0.png
tags: [Announcement, Chaos Mesh, Chaos Engineering]
---

![Chaos-Mesh-1.0 - Chaos-Engineering-on-Kubernetes-Made-Easier](/img/chaos-mesh-1.0.png)

Today, we are proud to announce the general availability of Chaos Mesh® 1.0, following its entry into CNCF as a [sandbox project](https://pingcap.com/blog/announcing-chaos-mesh-as-a-cncf-sandbox-project) in July, 2020.
<!--truncate-->

Chaos Mesh 1.0 is a major milestone in the project’s development. After 10 months of effort within the open-source community, Chaos Mesh is now ready in terms of functionality, scalability, and ease of use. Here are some highlights.

## Powerful chaos support 

[Chaos Mesh](https://chaos-mesh.org) originated in the testing framework of [TiDB](https://pingcap.com/products/tidb), a distributed database, so it takes into account the possible faults of a distributed system. Chaos Mesh provides comprehensive and fine-grained fault types, covering the Pod, the network, system I/O, and the kernel. Chaos experiments are defined in YAML, which is fast and easy to use.

Chaos Mesh 1.0 supports the following fault types:

* clock-skew: Simulates clock skew
* container-kill: Simulates the container being killed
* cpu-burn: Simulates CPU pressure
* io-attribution-override: Simulates file exceptions
* io-fault: Simulates file system I/O errors
* io-latency: Simulates file system I/O latency 
* kernel-injection: Simulates kernel failures
* memory-burn: Simulates memory pressure
* network-corrupt: Simulates network packet corruption
* network-duplication: Simulates network packet duplication
* network-latency: Simulate network latency
* network-loss: Simulates network loss
* network-partition: Simulates network partition
* pod-failure: Simulates continuous unavailability of Kubernetes Pods
* pod-kill: Simulates the Kubernetes Pod being killed

## Visual chaos orchestration  

The Chaos Dashboard component is a one-stop web interface for Chaos Mesh users to orchestrate chaos experiments. Previously, Chaos Dashboard was only available for testing TiDB. With Chaos Mesh 1.0, it is available to everyone. Chaos Dashboard greatly simplifies the complexity of chaos experiments. With only a few mouse clicks, you can define the scope of the chaos experiment, specify the type of chaos injection, define scheduling rules, and observe the results of the chaos experiment—all in the same web interface.

![Chaos Dashboard](/img/chaos-dashboard.gif)

## Grafana plug-in for enhanced observability

To further improve the observability of chaos experiments, Chaos Mesh 1.0 includes a Grafana plug-in to allow you to directly display real-time chaos experiment information on your application monitoring panel. Currently, the chaos experiment information is displayed as annotations. This way, you can simultaneously observe the running status of the application and the current chaos experiment information.

![Chaos status and application status on Grafana](/img/chaos-status.png)

## Safe and controllable chaos 

When we conduct chaos experiments, it is vital that we keep strict control over the chaos scope or “blast radius.” Chaos Mesh 1.0 not only provides a wealth of selectors to accurately control the scope of the experiment, but it also enables you to set protected Namespaces to protect important applications. You can also use Namespace permissions to limit the scope of Chaos Mesh to a specific Namespace. Together, these features make chaos experiments with Chaos Mesh safe and controllable.

## Try it out now 

You can quickly deploy Chaos Mesh in your Kubernetes environment through the `install.sh` script or the Helm tool. For specific installation steps, please refer to the [Chaos Mesh Getting Started](https://chaos-mesh.org/docs/installation/installation) document. In addition, thanks to the [Katakoda interactive tutorial](https://chaos-mesh.org/interactiveTutorial), you can also quickly get your hands on Chaos Mesh without having to deploy it.     

If you haven’t upgraded to 1.0 GA, please refer to the [1.0 Release Notes](https://github.com/chaos-mesh/chaos-mesh/releases/tag/v1.0) for the changes and upgrade guidelines.

## Thanks 

Thanks to all our Chaos Mesh [contributors](https://github.com/chaos-mesh/chaos-mesh/graphs/contributors)!  

If you are interested in Chaos Mesh, you’re welcome to join us by submitting issues, or contributing code, documentation, or articles. We look forward to your participation and feedback!

