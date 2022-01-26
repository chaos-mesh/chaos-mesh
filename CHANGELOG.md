# Chaos Mesh Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

For more information and how-to, see [RFC: Keep A Changelog](https://github.com/chaos-mesh/rfcs/blob/main/text/2022-01-17-keep-a-changelog.md).

## [Unreleased]

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

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix wrong field name of PhysicalMachineChaos on Chaos Dashboard [#2724](https://github.com/chaos-mesh/chaos-mesh/pull/2724)
- Fix field descriptions of GCPChaos [#2791](https://github.com/chaos-mesh/chaos-mesh/pull/2791)
- Fix chaos experiment "not found" on Chaos Dashboard [#2698](https://github.com/chaos-mesh/chaos-mesh/pull/2698)

### Security

- Nothing

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
