# Chaos Mesh Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

For more information and how-to, see [RFC: Keep A Changelog](https://github.com/chaos-mesh/rfcs/blob/main/text/2022-01-17-keep-a-changelog.md).

## [Unreleased]

### Added

- Allow annotations on chaos-controller-manager and chaos-daemon ServiceAccount [#4106](https://github.com/chaos-mesh/chaos-mesh/pull/4106)
- Support for deploying chaos-dashboard under the subpath [#4093](https://github.com/chaos-mesh/chaos-mesh/pull/4093)
- Support more rate units for networkchaos [#4129](https://github.com/chaos-mesh/chaos-mesh/pull/4129)
- Support for deploying chaos-dashboard with sidecar containers in helm chart [#4164](https://github.com/chaos-mesh/chaos-mesh/pull/4164)
- Add `values.schema.json` [#4205](https://github.com/chaos-mesh/chaos-mesh/pull/4205)
- Add [`GreptimeDB`](https://greptime.com) to ADOPTERS.md [#4245](https://github.com/chaos-mesh/chaos-mesh/pull/4245)
- Support configurable chaos-dns-server pod affinities[#4260](https://github.com/chaos-mesh/chaos-mesh/pull/4260)
- Add experiment `name`, `namespace` and `kind` to "apply chaos" and "recover chaos" log messages [4278](https://github.com/chaos-mesh/chaos-mesh/pull/4278)
- Support for watch remote chaos experiments to locally chaos  [#4188](https://github.com/chaos-mesh/chaos-mesh/pull/4188)
- Support configuration of logging through helm chart / environment [#4249](https://github.com/chaos-mesh/chaos-mesh/pull/4249)

### Changed

- Change `awschaos/*/impl.go` to use `aws_session_token` with static credentials provider so that users can perform awschaos using [temporary AWS credentials](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_use-resources.html) [#4066](https://github.com/chaos-mesh/chaos-mesh/pull/4066)
- Enable consistency for ghcr.io registry [#4091](https://github.com/chaos-mesh/chaos-mesh/pull/4091) [#4134](https://github.com/chaos-mesh/chaos-mesh/pull/4134)
- Prevent mousewheel scroll from changing numeric input values. [#4133](https://github.com/chaos-mesh/chaos-mesh/pull/4133)
- Set the database column type to text for 'Workflow/Experiment/Schedule' [#4151](https://github.com/chaos-mesh/chaos-mesh/pull/4151)
- Bump go to v1.20 [#4157](https://github.com/chaos-mesh/chaos-mesh/pull/4157)
- Upgrade docker/login-action to v2 [#4167](https://github.com/chaos-mesh/chaos-mesh/pull/4167)
- Update k8s dependencies to 1.28.0 [#4154](https://github.com/chaos-mesh/chaos-mesh/pull/4154)
- Update lru dependency [#4189](https://github.com/chaos-mesh/chaos-mesh/pull/4189)
- Update swag dependency [#4191](https://github.com/chaos-mesh/chaos-mesh/pull/4191)
- Update ginkgo to 2.12.0 [#4190](https://github.com/chaos-mesh/chaos-mesh/pull/4190)
- Update k8s controller-runtime dependency [#4198](https://github.com/chaos-mesh/chaos-mesh/pull/4198)
- Automatically remove the token from the dashboard when it expires [#4193](https://github.com/chaos-mesh/chaos-mesh/pull/4193)
- Optimize `allInjected` and `allRecovered` states when targets are not selected [#4199](https://github.com/chaos-mesh/chaos-mesh/pull/4199)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Set `replicas: 1` automatically when HA is not enabled [#4079](https://github.com/chaos-mesh/chaos-mesh/pull/4079)
- Remove redundant where statements [#4084](https://github.com/chaos-mesh/chaos-mesh/pull/4084)
- Fix dashboard panic when exec delete action by finish_time [#4100](https://github.com/chaos-mesh/chaos-mesh/pull/4100)
- Fix remote cluster cannot upgrade helm release [#4075](https://github.com/chaos-mesh/chaos-mesh/pull/4075)
- Fix goroutine leak [#4229](https://github.com/chaos-mesh/chaos-mesh/pull/4229)
- Remove the duplicate `make test` [#4234](https://github.com/chaos-mesh/chaos-mesh/pull/4234)
- Fix daemon-server `SetDNSServer` endpoint to validate provided server address [#4246](https://github.com/chaos-mesh/chaos-mesh/pull/4246)
- Fix serial workflow node reconciler not to spawn the next child after status check abort [#4286](https://github.com/chaos-mesh/chaos-mesh/pull/4286)

### Security

- Remove `-k` from `curl` command lines in chaos-daemon Dockerfile [#4241](https://github.com/chaos-mesh/chaos-mesh/pull/4241)
- Disable GraphQL API in chaos-controller-manager by default, and only apply `pods/exec` RBAC rule if enabled [#4247](https://github.com/chaos-mesh/chaos-mesh/pull/4247)

## [2.6.0] - 2023-05-30

### Added

- Install offline Helm Chart for a multi-cluster [#3897](https://github.com/chaos-mesh/chaos-mesh/pull/3897)

### Changed

- Change CoreDNS listen port from 53 to 5353 [#4022](https://github.com/chaos-mesh/chaos-mesh/pull/4022)
- Bump go to v1.19.3 [#3770](https://github.com/chaos-mesh/chaos-mesh/pull/3770)
- Change ubuntu version from latest to 20.04 [#3817](https://github.com/chaos-mesh/chaos-mesh/pull/3817)
- Switch views between k8s and hosts nodes [#3830](https://github.com/chaos-mesh/chaos-mesh/pull/3830)
- New CI for finding merge conflicts [#3850](https://github.com/chaos-mesh/chaos-mesh/pull/3850)
- Upgrade byteman-helper to v4.0.20 [#3863](https://github.com/chaos-mesh/chaos-mesh/pull/3863)
- Helm: change default webhook port to 10250 [#3877](https://github.com/chaos-mesh/chaos-mesh/pull/3877)
- Upgrade base image for chaos-mesh to alpine:3.17 [#3893](https://github.com/chaos-mesh/chaos-mesh/pull/3893)
- Slow down releasing the latest version [#3900](https://github.com/chaos-mesh/chaos-mesh/pull/3900)
- Update k8s.io dependencies to v0.26.1 [#3902](https://github.com/chaos-mesh/chaos-mesh/pull/3902)
- Update sigs.k8s.io/controller-runtime to v0.14.1 and sigs.k8s.io/controller-tools to v0.11.1 [#3902](https://github.com/chaos-mesh/chaos-mesh/pull/3902)
- Change the package manager from `yarn` to `pnpm`. [#3965](https://github.com/chaos-mesh/chaos-mesh/pull/3965)
- Upgrade DNS CoreDNS image url to ghcr.io [#3488](https://github.com/chaos-mesh/chaos-mesh/pull/3488)
- Upgrade OS image for chaos-daemon container image [#3905](https://github.com/chaos-mesh/chaos-mesh/pull/3905)
- Replace openapi-generator with Orval and React Query [#3748](https://github.com/chaos-mesh/chaos-mesh/pull/3748)
- Cleanup makefile and provide `make help` [#3888](https://github.com/chaos-mesh/chaos-mesh/pull/3888)
- Remove `IN_DOCKER` environment variable in `Makefile` [#3992](https://github.com/chaos-mesh/chaos-mesh/pull/3992)
- Refine TTL config of Chaos dashboard [#4008](https://github.com/chaos-mesh/chaos-mesh/pull/4008)
- `pause` would return non zero exit code when the subcommand failed [#4018](https://github.com/chaos-mesh/chaos-mesh/pull/4018)
- use helm values to set chaos-daemon capabilities [#4030](https://github.com/chaos-mesh/chaos-mesh/pull/4030)
- Build binaries locally with `local/` prefix targets in `Makefile` [#4004](https://github.com/chaos-mesh/chaos-mesh/pull/4004)
- Use kubectl cluster-info dump to enhance e2e profiling [#3759](https://github.com/chaos-mesh/chaos-mesh/pull/3759)
- Upgrade fx event logger [#4036](https://github.com/chaos-mesh/chaos-mesh/pull/4036)
- Refine logging in `pkg/selector/physicalmachine` [#4037](https://github.com/chaos-mesh/chaos-mesh/pull/4037)
- Setup OWNERS and OWNERS_ALIASES [#4039](https://github.com/chaos-mesh/chaos-mesh/pull/4039)

### Deprecated

- Nothing

### Removed

- Remove no needed file crd-v1beta1.yaml [#3807](https://github.com/chaos-mesh/chaos-mesh/pull/3807)
- Remove useless kubebuilder comment in webhook [#3816](https://github.com/chaos-mesh/chaos-mesh/pull/3816)
- Remove the unused inject-v1-pod webhook. [#3885](https://github.com/chaos-mesh/chaos-mesh/pull/3885)

### Fixed

- Fix version comparison in install.sh [#3901](https://github.com/chaos-mesh/chaos-mesh/pull/3901)
- Fix stuck dashboard updates when using ReadWriteOnce PVCs [#3876](https://github.com/chaos-mesh/chaos-mesh/issues/3876)
- Fix MySQL NO_ZERO_IN_DATE by using `*time.Time` to represent finish time [#4056](https://github.com/chaos-mesh/chaos-mesh/pull/4056)

### Security

- Bump go to v1.19.7 to fix CVE-2022-41723 [#3978](https://github.com/chaos-mesh/chaos-mesh/pull/3978) [#3981](https://github.com/chaos-mesh/chaos-mesh/pull/3981)
- Bump github.com/opencontainers/runc from 1.1.4 to 1.1.5 [#3987](https://github.com/chaos-mesh/chaos-mesh/pull/3987)

## [2.5.0] - 2022-11-22

### Added

- Add `controller.securityContext` and `dashboard.securityContext` to Helm chart [#3603](https://github.com/chaos-mesh/chaos-mesh/pull/3603)
- Add `RemoteCluster` resource type [#3342](https://github.com/chaos-mesh/chaos-mesh/pull/3342)
- Add `clusterregistry` package to help developers to develop multi-cluster reconciler [#3342](https://github.com/chaos-mesh/chaos-mesh/pull/3342)
- Add features about integration with helm to install Chaos Mesh in remote cluster [#3384](https://github.com/chaos-mesh/chaos-mesh/pull/3384)
- Add new CI "Manually Sign Container Images" to sign existing container images [#3708](https://github.com/chaos-mesh/chaos-mesh/pull/3708)
- Install and uninstall chaos mesh in remote cluster through `RemoteCluster` resource [#3414](https://github.com/chaos-mesh/chaos-mesh/pull/3414)
- MultiCluster: support inject / recover on remote cluster [#3453](https://github.com/chaos-mesh/chaos-mesh/pull/3453)
- Add TLS support for HTTPChaos [#3549](https://github.com/chaos-mesh/chaos-mesh/pull/3549)

### Changed

- Use the next generation `New Workflow` UI by default [#3718](https://github.com/chaos-mesh/chaos-mesh/pull/3718)
- StressChaos: Support cgroup v2 for docker and crio [#3698](https://github.com/chaos-mesh/chaos-mesh/pull/3698)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Remove the explicit use of pingcap/log [#3674](https://github.com/chaos-mesh/chaos-mesh/pull/3674)
- Fix typo in controller error message [#3704](https://github.com/chaos-mesh/chaos-mesh/pull/3704)
- Fix panic when logging, log kvs as pair [#3716](https://github.com/chaos-mesh/chaos-mesh/pull/3716)
- Fix timechaos not injected into the child process [#3725](https://github.com/chaos-mesh/chaos-mesh/pull/3725)
- Ignore `ScheduleSkipRemoveHistory` events to fix the memory of controller-manager keep increasing [#3761](https://github.com/chaos-mesh/chaos-mesh/issues/3761)
- Update `is mandatory` to true in a swagger comment [#3743](https://github.com/chaos-mesh/chaos-mesh/pull/3743)
- Enable mode when creating PhysicalMachineChaos with addresses in UI [#3797](https://github.com/chaos-mesh/chaos-mesh/pull/3797)

### Security

- Sign images and generate sbom when uploading images in CI [#3766](https://github.com/chaos-mesh/chaos-mesh/pull/3766)

## [2.4.2] - 2022-11-07

### Added

- Nothing

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix timechaos not injected into the child process [#3730](https://github.com/chaos-mesh/chaos-mesh/pull/3730)
- Update `is mandatory` to true in a swagger comment [#3743](https://github.com/chaos-mesh/chaos-mesh/pull/3743)

### Security

- Nothing

## [2.4.1] - 2022-09-27

### Added

- Nothing

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix: mark 2.4.1 not as pre-release version[#3679](https://github.com/chaos-mesh/chaos-mesh/pull/3679)

### Security

- Nothing

## [2.4.0] - 2022-09-23

### Added

- Add support for `PhysicalMachine` in UI [#3624](https://github.com/chaos-mesh/chaos-mesh/pull/3624)

### Changed

- Replace io/ioutil package with os package. [#3539](https://github.com/chaos-mesh/chaos-mesh/pull/3539)
- Refine logging in pkg/dashboard/apiserver/event, moved package level variable log into struct Service as a field. [#3528](https://github.com/chaos-mesh/chaos-mesh/pull/3528)
- Refine logging in pkg/dashboard/apiserver/auth/gcp, moved package level variable log into struct Service as a field [#3527](https://github.com/chaos-mesh/chaos-mesh/pull/3527)
- Change e2e config settings to append "pause image" args. [#3567](https://github.com/chaos-mesh/chaos-mesh/pull/3567)
- Update display of disabled scope in UI [#3621](https://github.com/chaos-mesh/chaos-mesh/pull/3621)
- Make the Scope render conditionally [#3622](https://github.com/chaos-mesh/chaos-mesh/pull/3622)
- Refine logging, remove the usage of klog in chaosctl [#3628](https://github.com/chaos-mesh/chaos-mesh/pull/3628)
- Dashboard: Fix rbac.yaml for token generation verbs/resources [#3370](https://github.com/chaos-mesh/chaos-mesh/pull/3370)
- Refine logging, remove the usage of klog in event recorder [#3629](https://github.com/chaos-mesh/chaos-mesh/pull/3629)
- Use int64 to restore latency for BlockChaos [#3638](https://github.com/chaos-mesh/chaos-mesh/pull/3638)
- Remove CRD v1beta1 support [#3630](https://github.com/chaos-mesh/chaos-mesh/pull/3630)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix NetworkChaos fail with @ifXX in the device [#3605](https://github.com/chaos-mesh/chaos-mesh/pull/3605)
- Fix BlockChaos can't show Chinese name. [#3536](https://github.com/chaos-mesh/chaos-mesh/pull/3536)
- Add `omitempty` JSON tag to optional fields of the CRD objects. [#3531](https://github.com/chaos-mesh/chaos-mesh/pull/3531)
- Fix "sidecar config" e2e test cases run failed in some scenario.[#3564](https://github.com/chaos-mesh/chaos-mesh/pull/3564)
- Fix Integration test with bumping kubectl version. [#3589](https://github.com/chaos-mesh/chaos-mesh/pull/3589)

### Security

- Nothing

## [2.3.3] - 2022-11-07

### Added

- Nothing

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Remove `limit` action of `BlockChaos` in the dashboard [#3655](https://github.com/chaos-mesh/chaos-mesh/pull/3655)
- Sync PhysicalMachineChaos to API client and forms [#3660](https://github.com/chaos-mesh/chaos-mesh/pull/3660)
- Fix timechaos not injected into the child process [#3729](https://github.com/chaos-mesh/chaos-mesh/pull/3729)
- Update `is mandatory` to true in a swagger comment [#3743](https://github.com/chaos-mesh/chaos-mesh/pull/3743)

### Security

- Nothing

## [2.3.2] - 2022-09-20

### Added

- Nothing

### Changed

- Use int64 to restore latency for BlockChaos [#3638](https://github.com/chaos-mesh/chaos-mesh/pull/3638)

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Fix: update helm chart annotation artifacthub.io/prerelease to false [#3626](https://github.com/chaos-mesh/chaos-mesh/pull/3626)

### Security

- Nothing

## [2.3.1] - 2022-09-02

### Added

- Nothing

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Protect chaos available namespace and filter namespaces if needed [#3473](https://github.com/chaos-mesh/chaos-mesh/pull/3473)
- Respect flag `enableProfiling` and do not register profiler endpoints when it's false [#3474](https://github.com/chaos-mesh/chaos-mesh/pull/3474)
- Fix the blank screen after creating chaos experiment with "By YAML" [#3489](https://github.com/chaos-mesh/chaos-mesh/pull/3489)
- Update hint text about the manual token generating process for Kubernetes 1.24+ [#3505](https://github.com/chaos-mesh/chaos-mesh/pull/3505)
- Fix IOChaos `containerNames` field in UI [#3533](https://github.com/chaos-mesh/chaos-mesh/pull/3533)
- Fix BlockChaos can't show Chinese name. [#3536](https://github.com/chaos-mesh/chaos-mesh/pull/3536)
- Fix recover bug when setting force recover to true [#3578](https://github.com/chaos-mesh/chaos-mesh/pull/3578)
- Fix generate token failed on chaos dashboard [#3595](https://github.com/chaos-mesh/chaos-mesh/pull/3595)

### Security

- Nothing

## [2.3.0] - 2022-07-29

### Added

- Add more status for record [#3170](https://github.com/chaos-mesh/chaos-mesh/pull/3170)
- Add `chaosDaemon.updateStrategy` to Helm chart to allow configuring `DaemonSetUpdateStrategy` for chaos-daemon [#3108](https://github.com/chaos-mesh/chaos-mesh/pull/3108)
- Add AArch64 support for TimeChaos [#3088](https://github.com/chaos-mesh/chaos-mesh/pull/3088)
- Add integration test and link test on arm [#3177](https://github.com/chaos-mesh/chaos-mesh/pull/3177)
- Add `spec.privateKey.rotationPolicy` to Certificates, to comply with requirements in cert-manager 1.8 [#3325](https://github.com/chaos-mesh/chaos-mesh/pull/3325)
- Add `clusterregistry` package to help developers to develop multi-cluster reconciler [#3342](https://github.com/chaos-mesh/chaos-mesh/pull/3342)
- Support `Suspend` in next generation `New Workflow`'s UI [#3254](https://github.com/chaos-mesh/chaos-mesh/pull/3254)
- Add helm annotations for Artifact Hub [#3355](https://github.com/chaos-mesh/chaos-mesh/pull/3355)
- Add implementation of blockchaos in chaos-daemon [#2907](https://github.com/chaos-mesh/chaos-mesh/pull/2907)
- Bump chaos-tproxy to v0.5.1 [#3412](https://github.com/chaos-mesh/chaos-mesh/pull/3412)
- Allow importing external workflows and copying flow nodes in next generation `New Workflow` [#3368](https://github.com/chaos-mesh/chaos-mesh/pull/3368)
- Add `QPS` and `Burst` for Chaos Dashboard Configuration [#3476](https://github.com/chaos-mesh/chaos-mesh/pull/3476)
- Add guide and example for monitoring Chaos Mesh [#3030](https://github.com/chaos-mesh/chaos-mesh/pull/3030)
- Support `KernelChaos` in `AutoForm` [#3449](https://github.com/chaos-mesh/chaos-mesh/pull/3449)
- Sync latest Chaosd and PhysicalMachineChaos [#3477](https://github.com/chaos-mesh/chaos-mesh/pull/3477)
- Add accept-tcp-flag to network delay in PysicalMachineChaos [#3588](https://github.com/chaos-mesh/chaos-mesh/pull/3588)

### Changed

- Helm charts: update validate-auth to chaos-mesh-validation-auth [#3193](https://github.com/chaos-mesh/chaos-mesh/pull/3193)
- Helm charts: support latest api version of dashboard ingress [#3066](https://github.com/chaos-mesh/chaos-mesh/pull/3066)
- Update shell script to support shellchecks [#3230](https://github.com/chaos-mesh/chaos-mesh/pull/3230)
- CI: build dev-env and build-env for e2e tests if required [#3264](https://github.com/chaos-mesh/chaos-mesh/pull/3264)
- CI: version unrelated manifests [#3293](https://github.com/chaos-mesh/chaos-mesh/pull/3293)
- Bump chaos-tproxy to v0.4.6 [#3272](https://github.com/chaos-mesh/chaos-mesh/pull/3272)
- Helm charts: using 0.0.0 as version and appVersion [#3311](https://github.com/chaos-mesh/chaos-mesh/pull/3311)
- Add a comment to the flag size of memory stress in the dashboard [#3359](https://github.com/chaos-mesh/chaos-mesh/pull/3359)
- Refine logging in pkg/dashboard/store, removed global the log [#3143](https://github.com/chaos-mesh/chaos-mesh/pull/3143)
- Renamed namespace from chaos-testing to chaos-mesh [#3353](https://github.com/chaos-mesh/chaos-mesh/pull/3353)
- Use ContainerSelector in kernel chaos [#3395](https://github.com/chaos-mesh/chaos-mesh/pull/3395)
- Make possible to have more than one dns chaos server [#3381](https://github.com/chaos-mesh/chaos-mesh/pull/3381)
- Helm charts: Relax allowedHostPaths in chaos-daemon PSP [#3350](https://github.com/chaos-mesh/chaos-mesh/pull/3350)
- Run build image ci on self-hosted machine [#3429](https://github.com/chaos-mesh/chaos-mesh/pull/3429)
- Simplified logic and add test case about finalizers. [#3422](https://github.com/chaos-mesh/chaos-mesh/pull/3422)
- Update API requests with OpenAPI generated client [#2926](https://github.com/chaos-mesh/chaos-mesh/pull/2926)
- Implement some missing methods in ctrl server [#3462](https://github.com/chaos-mesh/chaos-mesh/pull/3462)
- Use `net.Interfaces()` to implement `getAllInterfaces()` [#3484](https://github.com/chaos-mesh/chaos-mesh/pull/3484)

### Deprecated

- Nothing

### Removed

- Removed extra import of common pkg in chaosctl/cmd/logs.go
- Removed unused local function from statuscheck/manager.go [#3228](https://github.com/chaos-mesh/chaos-mesh/pull/3228)
- Removed ui build and test for arm64 [#3305](https://github.com/chaos-mesh/chaos-mesh/pull/3305)
- Remove sed need (SC2001) [#3248](https://github.com/chaos-mesh/chaos-mesh/pull/3248)
- Removed not used clientset in cmd/chaos-controller-manager/main.go [#3334](https://github.com/chaos-mesh/chaos-mesh/pull/3334)
- Removed not used globalCacheReader in cmd/chaos-controller-manager/provider/controller.go [#3343](https://github.com/chaos-mesh/chaos-mesh/pull/3343)
- Removed unsupported action comments of blockchaos [#3435](https://github.com/chaos-mesh/chaos-mesh/pull/3435)

### Fixed

- Update description of memory stressors [#3225](https://github.com/chaos-mesh/chaos-mesh/pull/3225)
- Isolate `target` field and `Scope` when creating `NetworkChaos` in UI [#3223](https://github.com/chaos-mesh/chaos-mesh/issues/3221)
- Add arm64 architecture to ci_skip to pass required test [#3305](https://github.com/chaos-mesh/chaos-mesh/pull/3305)
- Adapt install.sh for kubectl/kubernetes cluster greater than 1.24 [#3177](https://github.com/chaos-mesh/chaos-mesh/pull/3177)
- SC2166: Use || or && rather than -o or -a [#3235](https://github.com/chaos-mesh/chaos-mesh/pull/3235)
- SC2206: Use quote to prevent word splitting/globbing [#3234](https://github.com/chaos-mesh/chaos-mesh/pull/3234)
- Fix make check does not respect the env-images.yaml [#3210] (<https://github.com/chaos-mesh/chaos-mesh/pull/3210>)
- SC2004: Remove unnecessary $ on arithmetic variables [#3247](https://github.com/chaos-mesh/chaos-mesh/pull/3247)
- PhysicalMachineChaos: update stress options type [#3347](https://github.com/chaos-mesh/chaos-mesh/pull/3347)
- PhysicalMachineChaos: remove validate for IP and host for delay, loss, duplicate, corruption [#3483](https://github.com/chaos-mesh/chaos-mesh/pull/3483)
- StressChaos: run `pause` before `choom` [#3405](https://github.com/chaos-mesh/chaos-mesh/pull/3405)
- JVMChaos: update the error message that can be ignored [#3415](https://github.com/chaos-mesh/chaos-mesh/pull/3415)
- Fix Workflow Validating Webhook Panic [#3413](https://github.com/chaos-mesh/chaos-mesh/pull/3413)
- Overwrite $IMAGE_BUILD_ENV_TAG with $IMAGE_TAG-$ARCH in `upload_env_image.yml` github action [#3444](https://github.com/chaos-mesh/chaos-mesh/pull/3444)
- Add a judgement of `enterNS` in `getAllInterfaces()` [#3459](https://github.com/chaos-mesh/chaos-mesh/pull/3459)
- Fix JVMChaos loading missing jar file for injection [#3491](https://github.com/chaos-mesh/chaos-mesh/pull/3491)

### Security

- Nothing

## [2.2.3] - 2022-08-04

### Added

- Add `QPS` and `Burst` for Chaos Dashboard Configuration [#3476](https://github.com/chaos-mesh/chaos-mesh/pull/3476)

### Changed

- Implement some missing methods in ctrl server [#3470](https://github.com/chaos-mesh/chaos-mesh/pull/3470)

### Deprecated

- Nothing

### Changed

- Nothing

### Removed

- Nothing

### Fixed

- Protect chaos available namespace and filter namespaces if needed [#3473](https://github.com/chaos-mesh/chaos-mesh/pull/3473)
- Respect flag `enableProfiling` and do not register profiler endpoints when it's false [#3474](https://github.com/chaos-mesh/chaos-mesh/pull/3474)
- Fix the blank screen after creating chaos experiment with "By YAML" [#3489](https://github.com/chaos-mesh/chaos-mesh/pull/3489)
- Update hint text about the manual token generating process for Kubernetes 1.24+ [#3505](https://github.com/chaos-mesh/chaos-mesh/pull/3505)

### Security

- Nothing

## [2.2.2] - 2022-07-07

### Changed

- Bump chaos-tproxy to v0.5.1 [#3426](https://github.com/chaos-mesh/chaos-mesh/pull/3426)
- Run build image ci on self-hosted machine [#3429](https://github.com/chaos-mesh/chaos-mesh/pull/3429)

## [2.2.1] - 2022-06-29

### Added

- JVMChaos: support inject fault into MySQL client [#3189](https://github.com/chaos-mesh/chaos-mesh/pull/3189)

### Changed

- Nothing

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- PhysicalMachineChaos: update stress options type [#3354](https://github.com/chaos-mesh/chaos-mesh/pull/3354)
- StressChaos: run `pause` before `choom` [#3405](https://github.com/chaos-mesh/chaos-mesh/pull/3405)

### Security

- Nothing

## [2.2.0] - 2022-04-29

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
- Add forced recovery of httpchaos, iochaos, stresschaos, and networkchaos for chaosctl [#2992](https://github.com/chaos-mesh/chaos-mesh/pull/2992)
- Add namespace and pod name in failed event for podxxxchaos crd [#3178](https://github.com/chaos-mesh/chaos-mesh/pull/3178)
- Add next generation `New Workflow` in UI [#3185](https://github.com/chaos-mesh/chaos-mesh/pull/3185)
- JVMChaos: support inject fault into MySQL client [#3189](https://github.com/chaos-mesh/chaos-mesh/pull/3189)

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
- Bump toda to v0.2.3 [#3131](https://github.com/chaos-mesh/chaos-mesh/pull/3131)
- refactor: rename reconcileContext to reconcileInfo [#3154](https://github.com/chaos-mesh/chaos-mesh/pull/3154)
- Migrate e2e tests from self-hosted Jenkins to Github Action [#2986](https://github.com/chaos-mesh/chaos-mesh/pull/2986)
- Bump minimist from 1.2.5 to 1.2.6 in /ui [#3058](https://github.com/chaos-mesh/chaos-mesh/pull/3058)
- Specify image tag of `build-env` and `dev-env` for each branch [#3071](https://github.com/chaos-mesh/chaos-mesh/pull/3071)
- Specify image tag in e2e tests [#3147](https://github.com/chaos-mesh/chaos-mesh/pull/3147)
- Must update CHANGELOG [#3148](https://github.com/chaos-mesh/chaos-mesh/pull/3148)
- Use chaosDaemon.mtls.enabled instead of dashboard.securityMode for chaos-daemon mtls [#3168](https://github.com/chaos-mesh/chaos-mesh/pull/3168)
- Helm charts: component chaos-dashboard use certain service account and roles [#3145](https://github.com/chaos-mesh/chaos-mesh/pull/3145)
- Refactor helm charts template, split out webhook configuration and secrets [#3159](https://github.com/chaos-mesh/chaos-mesh/pull/3159)
- Helm charts: apply webhook.FailurePolicy to all the webhooks with default value `Fail` [#3184](https://github.com/chaos-mesh/chaos-mesh/pull/3184)
- Bump memStress from v0.2.1 to v0.3 [#3186](https://github.com/chaos-mesh/chaos-mesh/pull/3186)
- Helm charts: configure ca bundle for webhook explicitly [#3190](https://github.com/chaos-mesh/chaos-mesh/pull/3190)
- Refine logging in pkg/selector/generic/namespace [#3214](https://github.com/chaos-mesh/chaos-mesh/pull/3214)

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
- Fix `real_gettimeofday` on arm64 [#2849](https://github.com/chaos-mesh/chaos-mesh/pull/2849)
- Fix Github Action `upload-image` [#2935](https://github.com/chaos-mesh/chaos-mesh/pull/2935)
- Fix JVMChaos to handle the situation that the container which holds the JVM rules has been deleted [#2981](https://github.com/chaos-mesh/chaos-mesh/pull/2981)
- Fix typo in comments for Chaos API [#3109](https://github.com/chaos-mesh/chaos-mesh/pull/3109)

### Security

- Nothing

## [v2.1.8] - 2022-08-30

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

- Fix default value for concurrencyPolicy [#2622](https://github.com/chaos-mesh/chaos-mesh/pull/2622)
- Enable the webhooks for `Schedule` and `Workflow` [#2622](https://github.com/chaos-mesh/chaos-mesh/pull/2622)
- Fix PhysicalMachineChaos to make it able to create network bandwidth experiment. [#2850](https://github.com/chaos-mesh/chaos-mesh/pull/2850)
- Fix workflow emit new events after accomplished [#2911](https://github.com/chaos-mesh/chaos-mesh/pull/2911)
- Fix human unreadable logging timestamp [#2808](https://github.com/chaos-mesh/chaos-mesh/pull/2808) [#2902](https://github.com/chaos-mesh/chaos-mesh/pull/2902) [#2973](https://github.com/chaos-mesh/chaos-mesh/pull/2973)
- Fix default value of percent field in iochaos [#3018](https://github.com/chaos-mesh/chaos-mesh/pull/3018)
- Fix the unexpected CPU stress for StressChaos with cpu resource limit [#3102](https://github.com/chaos-mesh/chaos-mesh/pull/3102)
- Fix the bug that create JVMChaos failed in workflow [#3156](https://github.com/chaos-mesh/chaos-mesh/pull/3156)

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
