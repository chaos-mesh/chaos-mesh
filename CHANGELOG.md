# Chaos Mesh Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

For more information and how-to, see [RFC: Keep A Changelog](https://github.com/chaos-mesh/rfcs/blob/main/text/2022-01-17-keep-a-changelog.md).

## [Unreleased]

### Added

- Nothing

### Changed

- Removed docker registry mirror [#2797](https://github.com/chaos-mesh/chaos-mesh/pull/2797)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix default value for concurrencyPolicy [#2834](https://github.com/chaos-mesh/chaos-mesh/pull/2834)
- Bump toda to v0.2.3 [#3131](https://github.com/chaos-mesh/chaos-mesh/pull/3131)

### Security

- Nothing

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
