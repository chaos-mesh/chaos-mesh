# Chaos Mesh Roadmap

This document is intended to describe high-level plans for the Chaos Mesh project and is neither comprehensive nor prescriptive. For a more granular view of planned work, please refer to the project's upcoming [milestones](https://github.com/chaos-mesh/chaos-mesh/milestones).

## v1.0

- [x] Support time skew chaos. Simulate time jumping forward or backward.
- [x] Add container kill chaos. Simulate killing a specified container in a multi-container pod.
- [x] Add CPU chaos. Simulate CPU being busy.
- [x] Add memory chaos. Simulate memory allocation failure.
- [x] Make scheduler optional. Support single time chaos triggering.
- [x] Support helm-less install.
- [x] Support force clean finalizer with annotation.
- [x] Support the basic version of Chaos Dashboard.

## v2.0

- [x] Improve Chaos Dashboard and make it easier to use
- [x] Support status checks. A status check is used to evaluate the health of your environment.
- [x] Support defining the scenario to manage a group of chaos experiments.
- [x] Support generating the report for each chaos scenario.
- [x] Add JVM chaos. Support injecting faults into Java applications.
- [x] Add HTTP Chaos. Support injecting faults into http connections.
- [ ] ~~Add GRPC Chaos. Support injecting faults into GRPC connections.~~
- [ ] ~~Support injecting faults into native components of Kubernetes.~~

## Medium term

- [x] Manage and schedule chaos experiments on Kubernetes targets and non-Kubernetes targets on a unified dashboard.
- [x] Improve JVMChaos and support dynamic injection.
- [ ] Support injecting faults into native components of Kubernetes.
- [ ] More comprehensive status inspection mechanism and reports.
- [ ] Improve observability via events logs and metrics.
- [ ] Improve authentication system, and support using GCP/AWS account to log in chaos dashboard.
- [ ] Add GRPC Chaos. Support injecting faults into GRPC connections.
- [ ] A new component to force recovery chaos experiments, and avoid experiments going out of control.
- [ ] Build a hub for users sharing their own chaos workflow and chaos types.
- [ ] Support doing chaos experiments on multiple Kubernetes clusters.
- [ ] Provide a plugin approach to extend complex chaos types, such as RabbitMQChaos, RedisChaos...
- [ ] Continue to enrich fault types.

