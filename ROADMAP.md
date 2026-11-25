# Chaos Mesh Roadmap

This document is intended to describe high-level plans for the Chaos Mesh project and is neither comprehensive nor prescriptive. For a more granular view of planned work, please refer to the project's upcoming [milestones](https://github.com/chaos-mesh/chaos-mesh/milestones).

## v1.0

- [x] Support time skew chaos. Simulate time jumping forward or backward.
- [x] Add container kill chaos. Simulate killing a specified container in a multi-container pod.
- [x] Add CPU chaos. Simulate CPU being busy.
- [x] Add memory chaos. Simulate memory allocation failure.
- [x] Make scheduler optional. Support single time chaos triggering.
- [x] Support installation without Helm.
- [x] Support forcing finalizer cleanup with an annotation.
- [x] Support the basic version of Chaos Dashboard.

## v2.0

- [x] Improve Chaos Dashboard and make it easier to use.
- [x] Support status checks. A status check is used to evaluate the health of your environment.
- [x] Support defining the scenario to manage a group of chaos experiments.
- [x] Support generating the report for each chaos scenario.
- [x] Add JVM chaos. Support injecting faults into Java applications.
- [x] Add HTTP Chaos. Support injecting faults into HTTP connections.
- [ ] ~~Add gRPC Chaos. Support injecting faults into gRPC connections.~~
- [ ] ~~Support injecting faults into native components of Kubernetes.~~

## Medium term

- [x] Manage and schedule chaos experiments on Kubernetes targets and non-Kubernetes targets on a unified dashboard.
- [x] Improve JVMChaos and support dynamic injection.
- [ ] Support injecting faults into native components of Kubernetes.
- [ ] More comprehensive status inspection mechanism and reports.
- [ ] Improve observability via events, logs, and metrics.
- [ ] Improve the authentication system and support using a GCP or AWS account to log in to the Chaos Dashboard.
- [ ] Add gRPC Chaos. Support injecting faults into gRPC connections.
- [ ] A new component to force recovery of chaos experiments and avoid experiments going out of control.
- [ ] Build a hub for users sharing their own chaos workflow and chaos types.
- [x] Support doing chaos experiments on multiple Kubernetes clusters.
- [ ] Provide a plugin approach to extend complex chaos types, such as RabbitMQChaos and RedisChaos.
- [ ] Continue to enrich fault types.
