# Chaos Mesh Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

- Nothing

### Security

- Nothing

## [2.1.2] - 2021-12-29

### Changed

- Provide additional print columns for chaos experiments #2526
- Refactor pkg/time #2570
- Rename “physic” to “host” on Chaos Dashboard #2645
- Restructure UI codebase #2590
- Upgrade UI dependencies #2685
- Set default selector mode from “one” to “all” #2680
- Workflow now ordered by creation time #2680
- Set up codecov for testing coverage reports #2679
- Speed up e2e tests #2617 #2702

### Fixed

- Fixed: error when using Schedule and PodChaos for injecting PodChaos as a cron job #2618
- Fixed: chaos-kernel build failure #2693

## [2.0.6] - 2021-12-29

### Changed

- Provide additional print columns for chaos experiments #2526
- Remove redundant codes #2704
- Speed up e2e tests #2617 #2718

### Fixed

- Fixed: error when using Schedule and PodChaos for injecting PodChaos as a cron job #2618
- Fixed: fail to recover when Chaos CR was deleted before appending finalizers #2624
- Fixed: chaos-kernel build failure #2693
- Fixed: Chaos Dashboard panic when creating StressChaos #2655
