# Chaos Mesh Roadmap

This document defines the roadmap for Chaos Mesh development.

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

- [ ] Improve Chaos Dashboard and make it easier to use
- [ ] Support status checks. A status check is used to evaluate the health of your environment.
- [ ] Support defining the scenario to manage a group of chaos experiments. 
- [ ] Support generating the report for each chaos scenario.
- [ ] Add JVM chaos. Support injecting faults into Java applications.
- [ ] Add HTTP Chaos. Support injecting faults into http connections.
- [ ] Add GRPC Chaos. Support injecting faults into GRPC connections.
- [ ] Support injecting faults into native components of Kubernetes.

## Long-term

- [x] chaos-operator
- [x] chaos-dashboard
- [ ] chaos-verify
- [ ] chaos-cloud
