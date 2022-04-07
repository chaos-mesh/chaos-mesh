# Chaos Mesh Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

For more information and how-to, see [RFC: Keep A Changelog](https://github.com/chaos-mesh/rfcs/blob/main/text/2022-01-17-keep-a-changelog.md).

## [Unreleased]

### Added

- Add metrics for archived objects in chaos-dashboard [#2568](https://github.com/chaos-mesh/chaos-mesh/pull/2568)
- Add metrics for iptables, ipset and tc metrics in chaos-daemon [#2540](https://github.com/chaos-mesh/chaos-mesh/pull/2540)
- Add metrics for emitted event counter in chaos-controller-manager [#2435](https://github.com/chaos-mesh/chaos-mesh/pull/2435)
- Add metrics for grpc client [#2458](https://github.com/chaos-mesh/chaos-mesh/pull/2458)
- Add metrics for grpc and HTTP request duration histogram [#2543](https://github.com/chaos-mesh/chaos-mesh/pull/2543)
- Add metrics for bpm controlled processes [#2497](https://github.com/chaos-mesh/chaos-mesh/pull/2497)
- Provide additional printer columns for `action` and `duration` [#2526](https://github.com/chaos-mesh/chaos-mesh/pull/2526)
- Add PhysicalMachine CRD [#2587](https://github.com/chaos-mesh/chaos-mesh/pull/2587)
- New command `physical-machine` to `chaosctl` [#2624](https://github.com/chaos-mesh/chaos-mesh/pull/2624)
- Add status "Deleting" for chaos experiments on Chaos Dashboard [#2708](https://github.com/chaos-mesh/chaos-mesh/pull/2708)
- Add time skew for gettimeofday [#2742](https://github.com/chaos-mesh/chaos-mesh/pull/2742)
- Add support of the Unified cgroup mode (tested with containerd runtime only) for linux stress experiments [#2928](https://github.com/chaos-mesh/chaos-mesh/pull/2928)
- Add `StatusCheck` CRD [#2954](https://github.com/chaos-mesh/chaos-mesh/pull/2954)
- Add support for declaring ports in external targets in NetworkChaos experiments [#2932](https://github.com/chaos-mesh/chaos-mesh/pull/2932)

### Changed

- Use pipeline controller to serialize common controllers [#2465](https://github.com/chaos-mesh/chaos-mesh/pull/2465)
- Enable mTLS between chaos-controller-manager and chaosd [#2580](https://github.com/chaos-mesh/chaos-mesh/pull/2580)
- Rename Physics to Host in Chaos Dashboard [#2645](https://github.com/chaos-mesh/chaos-mesh/pull/2645)
- Retry oneshot chaos if it's not selected [#2618](https://github.com/chaos-mesh/chaos-mesh/pull/2618)
- Bump gopsutil to v3 [#2681](https://github.com/chaos-mesh/chaos-mesh/pull/2681)
- Add prefix for identifier of toda and tproxy in bpm [#2673](https://github.com/chaos-mesh/chaos-mesh/pull/2673)
- Bump toda to v0.2.2 [#2747](https://github.com/chaos-mesh/chaos-mesh/pull/2747)
- Bump go to 1.17 [#2754](https://github.com/chaos-mesh/chaos-mesh/pull/2754)
- Use github.com/pkg/errors to replace fmt.Errorf and "errors" [#2779](https://github.com/chaos-mesh/chaos-mesh/pull/2779)
- Kill chaos-tproxy while failing to apply config [#2672](https://github.com/chaos-mesh/chaos-mesh/pull/2672)
- JVMChaos: ignore AgentLoadException when install agent [#2701](https://github.com/chaos-mesh/chaos-mesh/pull/2701)
- Bump container-runtime to v0.11.0 [#2778](https://github.com/chaos-mesh/chaos-mesh/pull/2778)
- Bump kubernetes dependencies to v1.23.1 [#2778](https://github.com/chaos-mesh/chaos-mesh/pull/2778)
- Removed docker registry mirror [#2797](https://github.com/chaos-mesh/chaos-mesh/pull/2797)
- Use OpenAPI definitions to generate API Client and Form data in UI [2770](https://github.com/chaos-mesh/chaos-mesh/pull/2770)
- Refine logging in pkg/selector/pod [#3002](https://github.com/chaos-mesh/chaos-mesh/pull/3002)
- Add `envFollowKubernetesPattern` to handle k8s-like format env in helm templates [2955](https://github.com/chaos-mesh/chaos-mesh/pull/2955)
- Bump chaos-tproxy to v0.4.5 [#2555](https://github.com/chaos-mesh/chaos-mesh/pull/2555)
- Re-implement chaosctl based on ctrlserver [#2950](https://github.com/chaos-mesh/chaos-mesh/pull/2950)
- Fix wrong zero value of httpchaos replace-body-action[#2990](https://github.com/chaos-mesh/chaos-mesh/pull/2990)
- Bump gqlgen to v0.17.2 [#3038](https://github.com/chaos-mesh/chaos-mesh/pull/3038)
- Bump go to v1.18 [#3055](https://github.com/chaos-mesh/chaos-mesh/pull/3055)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Unable to load from saved objects [#2585](https://github.com/chaos-mesh/chaos-mesh/pull/2585)
- Fix helm install error [#2591](https://github.com/chaos-mesh/chaos-mesh/pull/2591)
- Fix helm conditions in ingress [#2604](https://github.com/chaos-mesh/chaos-mesh/pull/2604)
- Fix typo in NewExperiment [#2535](https://github.com/chaos-mesh/chaos-mesh/pull/2535)
- Fix chaos-kernel build, mark bcc version [#2693](https://github.com/chaos-mesh/chaos-mesh/pull/2693)
- Fix wrong field name of PhysicalMachineChaos on Chaos Dashboard [#2724](https://github.com/chaos-mesh/chaos-mesh/pull/2724)
- Fix field descriptions of GCPChaos [#2791](https://github.com/chaos-mesh/chaos-mesh/pull/2791)
- Fix default value for concurrencyPolicy [#2622](https://github.com/chaos-mesh/chaos-mesh/pull/2622)
- Enable the webhooks for `Schedule` and `Workflow` [#2622](https://github.com/chaos-mesh/chaos-mesh/pull/2622)
- Fix PhysicalMachineChaos to make it able to create network bandwidth experiment. [#2850](https://github.com/chaos-mesh/chaos-mesh/pull/2850)
- Fix workflow emit new events after accomplished [#2911](https://github.com/chaos-mesh/chaos-mesh/pull/2911)
- Fix human unreadable logging timestamp [#2808](https://github.com/chaos-mesh/chaos-mesh/pull/2808) [#2902](https://github.com/chaos-mesh/chaos-mesh/pull/2902) [#2973](https://github.com/chaos-mesh/chaos-mesh/pull/2973)
- Fix default value of percent field in iochaos [#3018](https://github.com/chaos-mesh/chaos-mesh/pull/3018)

### Security

- Nothing

## [2.1.3] - 2022-01-27

### Added

- Add status "Deleting" for chaos experiments on Chaos Dashboard [#2708](https://github.com/chaos-mesh/chaos-mesh/pull/2708)

### Changed

- Add prefix for identifier of toda and tproxy in bpm [#2673](https://github.com/chaos-mesh/chaos-mesh/pull/2673)
- Bump toda to v0.2.2 [#2747](https://github.com/chaos-mesh/chaos-mesh/pull/2747)
- Bump go to 1.17 [#2754](https://github.com/chaos-mesh/chaos-mesh/pull/2754)
- JVMChaos ignore AgentLoadException when install agent [#2701](https://github.com/chaos-mesh/chaos-mesh/pull/2701)
- Bump container-runtime to v0.11.0 [#2807](https://github.com/chaos-mesh/chaos-mesh/pull/2807)
- Bump kubernetes dependencies to v1.23.1 [#2807](https://github.com/chaos-mesh/chaos-mesh/pull/2807)
- Kill chaos-tproxy while failing to apply config [#2672](https://github.com/chaos-mesh/chaos-mesh/pull/2672)

### Fixed

- Fix wrong field name of PhysicalMachineChaos on Chaos Dashboard [#2724](https://github.com/chaos-mesh/chaos-mesh/pull/2724)
- Fix field descriptions of GCPChaos [#2791](https://github.com/chaos-mesh/chaos-mesh/pull/2791)
- Fix chaos experiment "not found" on Chaos Dashboard [#2698](https://github.com/chaos-mesh/chaos-mesh/pull/2698)

## [2.1.2] - 2021-12-29

### Changed

- Provide additional print columns for chaos experiments [#2526](https://github.com/chaos-mesh/chaos-mesh/pull/2526)
- Refactor pkg/time [#2570](https://github.com/chaos-mesh/chaos-mesh/pull/2570)
- Rename “physic” to “host” on Chaos Dashboard [#2645](https://github.com/chaos-mesh/chaos-mesh/pull/2645)
- Restructure UI codebase [#2590](https://github.com/chaos-mesh/chaos-mesh/pull/2590)
- Upgrade UI dependencies [#2685](https://github.com/chaos-mesh/chaos-mesh/pull/2685)
- Set default selector mode from “one” to “all” [#2680](https://github.com/chaos-mesh/chaos-mesh/pull/2792)
- Workflow now ordered by creation time [#2680](https://github.com/chaos-mesh/chaos-mesh/pull/2680)
- Set up codecov for testing coverage reports [#2679](https://github.com/chaos-mesh/chaos-mesh/pull/2679)
- Speed up e2e tests [#2617](https://github.com/chaos-mesh/chaos-mesh/pull/2617) [#2702](https://github.com/chaos-mesh/chaos-mesh/pull/2702)

### Fixed

- Fixed: error when using Schedule and PodChaos for injecting PodChaos as a cron job [#2618](https://github.com/chaos-mesh/chaos-mesh/pull/2618)
- Fixed: chaos-kernel build failure [#2693](https://github.com/chaos-mesh/chaos-mesh/pull/2693)

## [2.0.7] - 2022-01-27

### Added

- Add status "Deleting" for chaos experiments on Chaos Dashboard [#2708](https://github.com/chaos-mesh/chaos-mesh/pull/2708)

### Changed

- Add prefix for identifier of toda and tproxy in bpm [#2673](https://github.com/chaos-mesh/chaos-mesh/pull/2673)
- Kill chaos-tproxy while failing to apply config [#2672](https://github.com/chaos-mesh/chaos-mesh/pull/2672)

### Fixed

- Fix chaos experiment "not found" on Chaos Dashboard [#2698](https://github.com/chaos-mesh/chaos-mesh/pull/2698)
- Fix field descriptions of GCPChaos [#2791](https://github.com/chaos-mesh/chaos-mesh/pull/2791)

## [2.0.6] - 2021-12-29

### Changed

- Provide additional print columns for chaos experiments [#2526](https://github.com/chaos-mesh/chaos-mesh/pull/2526)
- Remove redundant codes [#2704](https://github.com/chaos-mesh/chaos-mesh/pull/2704)
- Speed up e2e tests #2617 [#2718](https://github.com/chaos-mesh/chaos-mesh/pull/2718)

### Fixed

- Fixed: error when using Schedule and PodChaos for injecting PodChaos as a cron job [#2618](https://github.com/chaos-mesh/chaos-mesh/pull/2618)
- Fixed: fail to recover when Chaos CR was deleted before appending finalizers [#2624](https://github.com/chaos-mesh/chaos-mesh/pull/2624)
- Fixed: chaos-kernel build failure [#2693](https://github.com/chaos-mesh/chaos-mesh/pull/2693)
- Fixed: Chaos Dashboard panic when creating StressChaos [#2655](https://github.com/chaos-mesh/chaos-mesh/pull/2655)
