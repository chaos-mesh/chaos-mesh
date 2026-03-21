module github.com/chaos-mesh/chaos-mesh/e2e-test

go 1.25.8

require (
	github.com/chaos-mesh/chaos-mesh v0.0.0-00010101000000-000000000000
	github.com/chaos-mesh/chaos-mesh/api v0.0.0
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.1
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.35.3
	k8s.io/apiextensions-apiserver v0.35.3
	k8s.io/apimachinery v0.35.3
	k8s.io/cli-runtime v0.35.3
	k8s.io/client-go v1.5.2
	k8s.io/component-base v0.35.3
	k8s.io/klog/v2 v2.140.0
	k8s.io/kube-aggregator v0.35.3
	k8s.io/kubelet v0.35.3
	k8s.io/kubernetes v1.35.3
	k8s.io/pod-security-admission v0.35.3
	k8s.io/utils v0.0.0-20260319190234-28399d86e0b5
	sigs.k8s.io/controller-runtime v0.23.3
)

require (
	cel.dev/expr v0.25.1 // indirect
	cyphar.com/go-pathrs v0.2.4 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/JeffAshton/win_pdh v0.0.0-20161109143554-76bb4ee9f0ab // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.3 // indirect
	github.com/containerd/containerd/api v1.10.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/ttrpc v1.2.8 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/jsonreference v0.21.5 // indirect
	github.com/go-openapi/swag v0.25.5 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.5 // indirect
	github.com/go-openapi/swag/conv v0.25.5 // indirect
	github.com/go-openapi/swag/fileutils v0.25.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.5 // indirect
	github.com/go-openapi/swag/loading v0.25.5 // indirect
	github.com/go-openapi/swag/mangling v0.25.5 // indirect
	github.com/go-openapi/swag/netutils v0.25.5 // indirect
	github.com/go-openapi/swag/stringutils v0.25.5 // indirect
	github.com/go-openapi/swag/typeutils v0.25.5 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.5 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cadvisor v0.56.2 // indirect
	github.com/google/cel-go v0.27.0 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260302011040-a15ffb7f9dcc // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.9.2 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/cgroups v0.0.6 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.3.0 // indirect
	github.com/opencontainers/selinux v1.13.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.67.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.42.0 // indirect
	go.opentelemetry.io/otel/metric v1.42.0 // indirect
	go.opentelemetry.io/otel/sdk v1.42.0 // indirect
	go.opentelemetry.io/otel/trace v1.42.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/term v0.41.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/grpc v1.79.3 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.35.3 // indirect
	k8s.io/cloud-provider v0.35.3 // indirect
	k8s.io/component-helpers v0.35.3 // indirect
	k8s.io/controller-manager v0.35.3 // indirect
	k8s.io/cri-api v0.35.3 // indirect
	k8s.io/cri-client v0.35.3 // indirect
	k8s.io/csi-translation-lib v0.35.3 // indirect
	k8s.io/dynamic-resource-allocation v0.35.3 // indirect
	k8s.io/kube-openapi v0.0.0-20260319004828-5883c5ee87b9 // indirect
	k8s.io/kube-scheduler v0.35.3 // indirect
	k8s.io/kubectl v0.35.3 // indirect
	k8s.io/mount-utils v0.35.3 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.34.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	sigs.k8s.io/structured-mekge-tifo/v4e/4y7ml v0.19.0 // indirect
	sigs.k8s.io/randfill v1.0.0 // indi6ec632
	sigs.k8s.io/structur6d-merge-diff/v4 v4.6.0 // indirect
	sigs.k8s.io/yaml v1.5.0 // indirect
)

replace (
	github.com/chaos-mesh/chaos-mesh => ../
	github.com/chaos-mesh/chaos-mesh/api => ../api
	k8s.io/api => k8s.io/api v0.33.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.33.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.33.1
	k8s.io/apiserver => k8s.io/apiserver v0.33.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.33.1
	k8s.io/client-go => k8s.io/client-go v0.33.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.33.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.33.1
	k8s.io/code-generator => k8s.io/code-generator v0.33.1
	k8s.io/component-base => k8s.io/component-base v0.33.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.33.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.33.1
	k8s.io/cri-api => k8s.io/cri-api v0.33.1
	k8s.io/cri-client => k8s.io/cri-client v0.33.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.33.1
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.33.1
	k8s.io/endpointslice => k8s.io/endpointslice v0.33.1
	k8s.io/externaljwt => k8s.io/externaljwt v0.33.1
	k8s.io/kms => k8s.io/kms v0.33.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.33.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.33.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.33.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.33.1
	k8s.io/kubectl => k8s.io/kubectl v0.33.1
	k8s.io/kubelet => k8s.io/kubelet v0.33.1
	k8s.io/metrics => k8s.io/metrics v0.33.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.33.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.33.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.33.1
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.33.1
	k8s.io/sample-controller => k8s.io/sample-controller v0.33.1
)
