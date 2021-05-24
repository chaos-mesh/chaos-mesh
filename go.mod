module github.com/chaos-mesh/chaos-mesh

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Microsoft/go-winio v0.4.11 // indirect
	github.com/Microsoft/hcsshim v0.0.0-20190417211021-672e52e9209d // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/aws/aws-sdk-go-v2 v0.31.0
	github.com/aws/aws-sdk-go-v2/config v0.4.0
	github.com/aws/aws-sdk-go-v2/credentials v0.2.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v0.31.0
	github.com/chaos-mesh/chaos-mesh/api/v1alpha1 v0.0.0
	github.com/chaos-mesh/k8s_dns_chaos v0.2.0
	github.com/containerd/cgroups v0.0.0-20200404012852-53ba5634dc0f
	github.com/containerd/containerd v1.2.3
	github.com/containerd/continuity v0.0.0-20200107194136-26c1120b8d41 // indirect
	github.com/containerd/cri v1.11.1 // indirect
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00 // indirect
	github.com/containerd/typeurl v0.0.0-20200115183213-fe1d0d650e42 // indirect
	github.com/docker/docker v1.4.2-0.20180625184442-8e610b2b55bf
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/ethereum/go-ethereum v1.9.8
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/mux v1.7.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/golang-lru v0.5.3
	github.com/jinzhu/gorm v1.9.12
	github.com/joomcode/errorx v1.0.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/mgechev/revive v1.0.2-0.20200225072153-6219ca02fffb
	github.com/mitchellh/mapstructure v1.3.3
	github.com/moby/locker v1.0.1
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.9.0
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.0-rc9 // indirect
	github.com/pingcap/check v0.0.0-20191216031241-8a5a85928f12 // indirect
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/failpoint v0.0.0-20200210140405-f8f9fb234798
	github.com/pingcap/log v0.0.0-20200117041106-d28c14d3b1cd // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil v0.0.0-20180427012116-c95755e4bcd7
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14
	github.com/swaggo/gin-swagger v1.2.0
	github.com/swaggo/swag v1.6.7
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/vishvananda/netlink v1.0.0
	github.com/vishvananda/netns v0.0.0-20171111001504-be1fbeda1936 // indirect
	go.uber.org/fx v1.12.0
	go.uber.org/zap v1.15.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20200616195046-dc31b401abb5
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/api v0.13.0
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.23.0
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/cli-runtime v0.17.0
	k8s.io/client-go v0.18.2
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.4.1
)

replace github.com/chaos-mesh/chaos-mesh/api/v1alpha1 => ./api/v1alpha1

go 1.15
