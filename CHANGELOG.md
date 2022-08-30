# Chaos Mesh Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

For more information and how-to, see [RFC: Keep A Changelog](https://github.com/chaos-mesh/rfcs/blob/main/text/2022-01-17-keep-a-changelog.md).

## [Unreleased]

### Added

- Nothing

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

## [v2.1.8] - 2020-08.30

### Added

- Nothing

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix recover bug when setting force recover to true [#3578](https://github.com/chaos-mesh/chaos-mesh/pull/3578)
- Uniformly use Enter to add item in LabelField [#3580](https://github.com/chaos-mesh/chaos-mesh/pull/3580)

## [2.1.7] - 2022-07-29

### Added

- Add `QPS` and `Burst` for Chaos Dashboard Configuration [#3476](https://github.com/chaos-mesh/chaos-mesh/pull/3476)

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Nothing

## [2.1.6] - 2022-07-22

### Added

- Add e2e build image cache [#3097](https://github.com/chaos-mesh/chaos-mesh/pull/3151)
- PhysicalMachineChaos: add recover command in process [#3468](https://github.com/chaos-mesh/chaos-mesh/pull/3468)

### Changed

- Must update CHANGELOG [#3181](https://github.com/chaos-mesh/chaos-mesh/pull/3181)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix: find the correct pid by CommName [#3336](https://github.com/chaos-mesh/chaos-mesh/pull/3336)

### Security

- Nothing

## [2.1.5] - 2022-04-18

### Added

- Nothing

### Changed

- Migrate e2e tests from self-hosted Jenkins to Github Action [#2986](https://github.com/chaos-mesh/chaos-mesh/pull/2986)
- Bump minimist from 1.2.5 to 1.2.6 in /ui [#3058](https://github.com/chaos-mesh/chaos-mesh/pull/3058)
- Specify image tag of `build-env` and `dev-env` for each branch [#3071](https://github.com/chaos-mesh/chaos-mesh/pull/3071)
- Bump toda to v0.2.3 [#3131](https://github.com/chaos-mesh/chaos-mesh/pull/3131)
- Specify image tag in e2e tests [#3147](https://github.com/chaos-mesh/chaos-mesh/pull/3147)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix `real_gettimeofday` on arm64 [#2849](https://github.com/chaos-mesh/chaos-mesh/pull/2849)
- Fix Github Action `upload-image` [#2935](https://github.com/chaos-mesh/chaos-mesh/pull/2935)
- Fix JVMChaos to handle the situation that the container which holds the JVM rules has been deleted [#2981](https://github.com/chaos-mesh/chaos-mesh/pull/2981)
- Fix typo in comments for Chaos API [#3109](https://github.com/chaos-mesh/chaos-mesh/pull/3109)

### Security

- Nothing

## [2.1.4] - 2022-03-21

### Added

- Add time skew for gettimeofday [#2742](https://github.com/chaos-mesh/chaos-mesh/pull/2742)

### Changed

- Removed docker registry mirror [#2797](https://github.com/chaos-mesh/chaos-mesh/pull/2797)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix default value for concurrencyPolicy [#2836](https://github.com/chaos-mesh/chaos-mesh/pull/2836)
- Fix PhysicalMachineChaos to make it able to create network bandwidth experiment. [#2850](https://github.com/chaos-mesh/chaos-mesh/pull/2850)
- Fix workflow emit new events after accomplished [#2911](https://github.com/chaos-mesh/chaos-mesh/pull/2911)
- Fix human unreadable logging timestamp [#2970](https://github.com/chaos-mesh/chaos-mesh/pull/2970) [#2973](https://github.com/chaos-mesh/chaos-mesh/pull/2973) [#2991](https://github.com/chaos-mesh/chaos-mesh/pull/2991)
- Fix default value of percent field in iochaos [#3018](https://github.com/chaos-mesh/chaos-mesh/pull/3018)
- Fix the unexpected CPU stress for StressChaos with cpu resource limit [#3102](https://github.com/chaos-mesh/chaos-mesh/pull/3102)

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
- Fix default value for concurrencyPolicy [#2622](https://github.com/chaos-mesh/chaos-mesh/pull/2622)
- Enable the webhooks for `Schedule` and `Workflow` [#2622](https://github.com/chaos-mesh/chaos-mesh/pull/2622)

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
