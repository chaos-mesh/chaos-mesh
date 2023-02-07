// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PhysicalMachineChaosAction represents the chaos action about physical machine.
type PhysicalMachineChaosAction string

var (
	PMStressCPUAction            PhysicalMachineChaosAction = "stress-cpu"
	PMStressMemAction            PhysicalMachineChaosAction = "stress-mem"
	PMDiskWritePayloadAction     PhysicalMachineChaosAction = "disk-write-payload"
	PMDiskReadPayloadAction      PhysicalMachineChaosAction = "disk-read-payload"
	PMDiskFillAction             PhysicalMachineChaosAction = "disk-fill"
	PMNetworkCorruptAction       PhysicalMachineChaosAction = "network-corrupt"
	PMNetworkDuplicateAction     PhysicalMachineChaosAction = "network-duplicate"
	PMNetworkLossAction          PhysicalMachineChaosAction = "network-loss"
	PMNetworkDelayAction         PhysicalMachineChaosAction = "network-delay"
	PMNetworkPartitionAction     PhysicalMachineChaosAction = "network-partition"
	PMNetworkBandwidthAction     PhysicalMachineChaosAction = "network-bandwidth"
	PMNetworkDNSAction           PhysicalMachineChaosAction = "network-dns"
	PMNetworkFloodAction         PhysicalMachineChaosAction = "network-flood"
	PMNetworkDownAction          PhysicalMachineChaosAction = "network-down"
	PMProcessAction              PhysicalMachineChaosAction = "process"
	PMJVMExceptionAction         PhysicalMachineChaosAction = "jvm-exception"
	PMJVMGCAction                PhysicalMachineChaosAction = "jvm-gc"
	PMJVMLatencyAction           PhysicalMachineChaosAction = "jvm-latency"
	PMJVMReturnAction            PhysicalMachineChaosAction = "jvm-return"
	PMJVMStressAction            PhysicalMachineChaosAction = "jvm-stress"
	PMJVMRuleDataAction          PhysicalMachineChaosAction = "jvm-rule-data"
	PMJVMMySQLAction             PhysicalMachineChaosAction = "jvm-mysql"
	PMClockAction                PhysicalMachineChaosAction = "clock"
	PMRedisExpirationAction      PhysicalMachineChaosAction = "redis-expiration"
	PMRedisPenetrationAction     PhysicalMachineChaosAction = "redis-penetration"
	PMRedisCacheLimitAction      PhysicalMachineChaosAction = "redis-cacheLimit"
	PMRedisSentinelRestartAction PhysicalMachineChaosAction = "redis-restart"
	PMRedisSentinelStopAction    PhysicalMachineChaosAction = "redis-stop"
	PMKafkaFillAction            PhysicalMachineChaosAction = "kafka-fill"
	PMKafkaFloodAction           PhysicalMachineChaosAction = "kafka-flood"
	PMKafkaIOAction              PhysicalMachineChaosAction = "kafka-io"
	PMHTTPAbortAction            PhysicalMachineChaosAction = "http-abort"
	PMHTTPDelayAction            PhysicalMachineChaosAction = "http-delay"
	PMHTTPConfigAction           PhysicalMachineChaosAction = "http-config"
	PMHTTPRequestAction          PhysicalMachineChaosAction = "http-request"
	PMFileCreateAction           PhysicalMachineChaosAction = "file-create"
	PMFileModifyPrivilegeAction  PhysicalMachineChaosAction = "file-modify"
	PMFileDeleteAction           PhysicalMachineChaosAction = "file-delete"
	PMFileRenameAction           PhysicalMachineChaosAction = "file-rename"
	PMFileAppendAction           PhysicalMachineChaosAction = "file-append"
	PMFileReplaceAction          PhysicalMachineChaosAction = "file-replace"
	PMVMAction                   PhysicalMachineChaosAction = "vm"
	PMUserDefinedAction          PhysicalMachineChaosAction = "user_defined"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="action",type=string,JSONPath=`.spec.action`
// +kubebuilder:printcolumn:name="duration",type=string,JSONPath=`.spec.duration`
// +chaos-mesh:experiment

// PhysicalMachineChaos is the Schema for the physical machine chaos API
type PhysicalMachineChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a physical machine chaos experiment
	Spec PhysicalMachineChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the chaos experiment
	Status PhysicalMachineChaosStatus `json:"status,omitempty"`
}

// PhysicalMachineChaosSpec defines the desired state of PhysicalMachineChaos
type PhysicalMachineChaosSpec struct {
	// +kubebuilder:validation:Enum=stress-cpu;stress-mem;disk-read-payload;disk-write-payload;disk-fill;network-corrupt;network-duplicate;network-loss;network-delay;network-partition;network-dns;network-bandwidth;network-flood;network-down;process;jvm-exception;jvm-gc;jvm-latency;jvm-return;jvm-stress;jvm-rule-data;jvm-mysql;clock;redis-expiration;redis-penetration;redis-cacheLimit;redis-restart;redis-stop;kafka-fill;kafka-flood;kafka-io;file-create;file-modify;file-delete;file-rename;file-append;file-replace;vm;user_defined
	Action PhysicalMachineChaosAction `json:"action"`

	PhysicalMachineSelector `json:",inline"`

	// ExpInfo string `json:"expInfo"`
	ExpInfo `json:",inline"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty" webhook:"Duration"`

	// RemoteCluster represents the remote cluster where the chaos will be deployed
	// +optional
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// PhysicalMachineChaosStatus defines the observed state of PhysicalMachineChaos
type PhysicalMachineChaosStatus struct {
	ChaosStatus `json:",inline"`
}

func (obj *PhysicalMachineChaos) GetSelectorSpecs() map[string]interface{} {
	return map[string]interface{}{
		".": &obj.Spec.PhysicalMachineSelector,
	}
}

type PhysicalMachineSelector struct {
	// DEPRECATED: Use Selector instead.
	// Only one of Address and Selector could be specified.
	// +optional
	Address []string `json:"address,omitempty"`

	// Selector is used to select physical machines that are used to inject chaos action.
	// +optional
	Selector PhysicalMachineSelectorSpec `json:"selector,omitempty"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode SelectorMode `json:"mode"`

	// Value is required when the mode is set to `FixedMode` / `FixedPercentMode` / `RandomMaxPercentMode`.
	// If `FixedMode`, provide an integer of physical machines to do chaos action.
	// If `FixedPercentMode`, provide a number from 0-100 to specify the percent of physical machines the server can do chaos action.
	// IF `RandomMaxPercentMode`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value,omitempty"`
}

// PhysicalMachineSelectorSpec defines some selectors to select objects.
// If the all selectors are empty, all objects will be used in chaos experiment.
type PhysicalMachineSelectorSpec struct {
	GenericSelectorSpec `json:",inline"`

	// PhysicalMachines is a map of string keys and a set values that used to select physical machines.
	// The key defines the namespace which physical machine belong,
	// and each value is a set of physical machine names.
	// +optional
	PhysicalMachines map[string][]string `json:"physicalMachines,omitempty"`
}

func (spec *PhysicalMachineSelectorSpec) Empty() bool {
	if spec == nil {
		return true
	}
	if len(spec.AnnotationSelectors) != 0 || len(spec.FieldSelectors) != 0 || len(spec.LabelSelectors) != 0 ||
		len(spec.Namespaces) != 0 || len(spec.PhysicalMachines) != 0 || len(spec.ExpressionSelectors) != 0 {
		return false
	}
	return true
}

type ExpInfo struct {
	// the experiment ID
	// +optional
	UID string `json:"uid,omitempty" swaggerignore:"true"`

	// the subAction, generate automatically
	// +optional
	Action string `json:"action,omitempty" swaggerignore:"true"`

	// +ui:form:when=action=='stress-cpu'
	// +optional
	StressCPU *StressCPUSpec `json:"stress-cpu,omitempty"`

	// +ui:form:when=action=='stress-mem'
	// +optional
	StressMemory *StressMemorySpec `json:"stress-mem,omitempty"`

	// +ui:form:when=action=='disk-read-payload'
	// +optional
	DiskReadPayload *DiskPayloadSpec `json:"disk-read-payload,omitempty"`

	// +ui:form:when=action=='disk-write-payload'
	// +optional
	DiskWritePayload *DiskPayloadSpec `json:"disk-write-payload,omitempty"`

	// +ui:form:when=action=='disk-fill'
	// +optional
	DiskFill *DiskFillSpec `json:"disk-fill,omitempty"`

	// +ui:form:when=action=='network-corrupt'
	// +optional
	NetworkCorrupt *NetworkCorruptSpec `json:"network-corrupt,omitempty"`

	// +ui:form:when=action=='network-duplicate'
	// +optional
	NetworkDuplicate *NetworkDuplicateSpec `json:"network-duplicate,omitempty"`

	// +ui:form:when=action=='network-loss'
	// +optional
	NetworkLoss *NetworkLossSpec `json:"network-loss,omitempty"`

	// +ui:form:when=action=='network-delay'
	// +optional
	NetworkDelay *NetworkDelaySpec `json:"network-delay,omitempty"`

	// +ui:form:when=action=='network-partition'
	// +optional
	NetworkPartition *NetworkPartitionSpec `json:"network-partition,omitempty"`

	// +ui:form:when=action=='network-dns'
	// +optional
	NetworkDNS *NetworkDNSSpec `json:"network-dns,omitempty"`

	// +ui:form:when=action=='network-bandwidth'
	// +optional
	NetworkBandwidth *NetworkBandwidthSpec `json:"network-bandwidth,omitempty"`

	// +ui:form:when=action=='network-flood'
	// +optional
	NetworkFlood *NetworkFloodSpec `json:"network-flood,omitempty"`

	// +ui:form:when=action=='network-down'
	// +optional
	NetworkDown *NetworkDownSpec `json:"network-down,omitempty"`

	// +ui:form:when=action=='process'
	// +optional
	Process *ProcessSpec `json:"process,omitempty"`

	// +ui:form:when=action=='jvm-exception'
	// +optional
	JVMException *JVMExceptionSpec `json:"jvm-exception,omitempty"`

	// +ui:form:when=action=='jvm-gc'
	// +optional
	JVMGC *JVMGCSpec `json:"jvm-gc,omitempty"`

	// +ui:form:when=action=='jvm-latency'
	// +optional
	JVMLatency *JVMLatencySpec `json:"jvm-latency,omitempty"`

	// +ui:form:when=action=='jvm-return'
	// +optional
	JVMReturn *JVMReturnSpec `json:"jvm-return,omitempty"`

	// +ui:form:when=action=='jvm-stress'
	// +optional
	JVMStress *JVMStressSpec `json:"jvm-stress,omitempty"`

	// +ui:form:when=action=='jvm-rule-data'
	// +optional
	JVMRuleData *JVMRuleDataSpec `json:"jvm-rule-data,omitempty"`

	// +ui:form:when=action=='jvm-mysql'
	// +optional
	JVMMySQL *PMJVMMySQLSpec `json:"jvm-mysql,omitempty"`

	// +ui:form:when=action=='clock'
	// +optional
	Clock *ClockSpec `json:"clock,omitempty"`

	// +ui:form:when=action=='redis-expiration'
	// +optional
	RedisExpiration *RedisExpirationSpec `json:"redis-expiration,omitempty"`

	// +ui:form:when=action=='redis-penetration'
	// +optional
	RedisPenetration *RedisPenetrationSpec `json:"redis-penetration,omitempty"`

	// +ui:form:when=action=='redis-cacheLimit'
	// +optional
	RedisCacheLimit *RedisCacheLimitSpec `json:"redis-cacheLimit,omitempty"`

	// +ui:form:when=action=='redis-restart'
	// +optional
	RedisSentinelRestart *RedisSentinelRestartSpec `json:"redis-restart,omitempty"`

	// +ui:form:when=action=='redis-stop'
	// +optional
	RedisSentinelStop *RedisSentinelStopSpec `json:"redis-stop,omitempty"`

	// +ui:form:when=action=='kafka-fill'
	// +optional
	KafkaFill *KafkaFillSpec `json:"kafka-fill,omitempty"`

	// +ui:form:when=action=='kafka-flood'
	// +optional
	KafkaFlood *KafkaFloodSpec `json:"kafka-flood,omitempty"`

	// +ui:form:when=action=='kafka-io'
	// +optional
	KafkaIO *KafkaIOSpec `json:"kafka-io,omitempty"`

	// +ui:form:when=action=='http-abort'
	// +optional
	HTTPAbort *HTTPAbortSpec `json:"http-abort,omitempty"`

	// +ui:form:when=action=='http-delay'
	// +optional
	HTTPDelay *HTTPDelaySpec `json:"http-delay,omitempty"`

	// +ui:form:when=action=='http-config'
	// +optional
	HTTPConfig *HTTPConfigSpec `json:"http-config,omitempty"`

	// +ui:form:when=action=='http-request'
	// +optional
	HTTPRequest *HTTPRequestSpec `json:"http-request,omitempty"`

	// +ui:form:when=action=='file-create'
	// +optional
	FileCreate *FileCreateSpec `json:"file-create,omitempty"`

	// +ui:form:when=action=='file-modify'
	// +optional
	FileModifyPrivilege *FileModifyPrivilegeSpec `json:"file-modify,omitempty"`

	// +ui:form:when=action=='file-delete'
	// +optional
	FileDelete *FileDeleteSpec `json:"file-delete,omitempty"`

	// +ui:form:when=action=='file-create'
	// +optional
	FileRename *FileRenameSpec `json:"file-rename,omitempty"`

	// +ui:form:when=action=='file-append'
	// +optional
	FileAppend *FileAppendSpec `json:"file-append,omitempty"`

	// +ui:form:when=action=='file-replace'
	// +optional
	FileReplace *FileReplaceSpec `json:"file-replace,omitempty"`

	// +ui:form:when=action=='vm'
	// +optional
	VM *VMSpec `json:"vm,omitempty"`

	// +ui:form:when=action=='user_defined'
	// +optional
	UserDefined *UserDefinedSpec `json:"user_defined,omitempty"`
}

type StressCPUSpec struct {
	// specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100 is full loading.
	Load int `json:"load,omitempty"`
	// specifies N workers to apply the stressor.
	Workers int `json:"workers,omitempty"`
	// extend stress-ng options
	Options []string `json:"options,omitempty"`
}

type StressMemorySpec struct {
	// specifies N bytes consumed per vm worker, default is the total available memory.
	// One can specify the size as % of total available memory or in units of B, KB/KiB, MB/MiB, GB/GiB, TB/TiB..
	Size string `json:"size,omitempty"`
	// extend stress-ng options
	Options []string `json:"options,omitempty"`
}

type DiskFileSpec struct {
	// specifies how many units of data will write into the file path. support unit: c=1, w=2, b=512, kB=1000,
	// K=1024, MB=1000*1000, M=1024*1024, GB=1000*1000*1000, G=1024*1024*1024 BYTES. example : 1M | 512kB
	Size string `json:"size,omitempty"`
	// specifies the location to fill data in. if path not provided,
	// payload will read/write from/into a temp file, temp file will be deleted after writing
	Path string `json:"path,omitempty"`
}

type DiskPayloadSpec struct {
	DiskFileSpec `json:",inline"`

	// specifies the number of process work on writing, default 1, only 1-255 is valid value
	PayloadProcessNum uint8 `json:"payload-process-num,omitempty"`
}

type DiskFillSpec struct {
	DiskFileSpec `json:",inline"`

	// fill disk by fallocate
	FillByFallocate bool `json:"fill-by-fallocate,omitempty"`
}

type NetworkCommonSpec struct {
	// correlation is percentage (10 is 10%)
	Correlation string `json:"correlation,omitempty"`
	// the network interface to impact
	Device string `json:"device,omitempty"`
	// only impact egress traffic from these source ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010.
	// it can only be used in conjunction with -p tcp or -p udp
	SourcePort string `json:"source-port,omitempty"`
	// only impact egress traffic to these destination ports, use a ',' to separate or to indicate the range, such as 80, 8001:8010.
	// it can only be used in conjunction with -p tcp or -p udp
	EgressPort string `json:"egress-port,omitempty"`
	// only impact egress traffic to these IP addresses
	IPAddress string `json:"ip-address,omitempty"`
	// only impact traffic using this IP protocol, supported: tcp, udp, icmp, all
	IPProtocol string `json:"ip-protocol,omitempty"`
	// only impact traffic to these hostnames
	Hostname string `json:"hostname,omitempty"`
}

type NetworkCorruptSpec struct {
	NetworkCommonSpec `json:",inline"`

	// percentage of packets to corrupt (10 is 10%)
	Percent string `json:"percent,omitempty"`
}

type NetworkDuplicateSpec struct {
	NetworkCommonSpec `json:",inline"`

	// percentage of packets to duplicate (10 is 10%)
	Percent string `json:"percent,omitempty"`
}

type NetworkLossSpec struct {
	NetworkCommonSpec `json:",inline"`

	// percentage of packets to loss (10 is 10%)
	Percent string `json:"percent,omitempty"`
}

type NetworkDelaySpec struct {
	NetworkCommonSpec `json:",inline"`

	// jitter time, time units: ns, us (or µs), ms, s, m, h.
	Jitter string `json:"jitter,omitempty"`
	// delay egress time, time units: ns, us (or µs), ms, s, m, h.
	Latency string `json:"latency,omitempty"`
	// only the packet which match the tcp flag can be accepted, others will be dropped.
	// only set when the IPProtocol is tcp, used for partition.
	AcceptTCPFlags string `json:"accept-tcp-flags,omitempty"`
}

type NetworkPartitionSpec struct {
	// the network interface to impact
	Device string `json:"device,omitempty"`
	// only impact traffic to these hostnames
	Hostname string `json:"hostname,omitempty"`
	// only impact egress traffic to these IP addresses
	IPAddress string `json:"ip-address,omitempty"`
	// specifies the partition direction, values can be 'from', 'to'.
	// 'from' means packets coming from the 'IPAddress' or 'Hostname' and going to your server,
	// 'to' means packets originating from your server and going to the 'IPAddress' or 'Hostname'.
	Direction string `json:"direction,omitempty"`
	// only impact egress traffic to these IP addresses
	IPProtocol string `json:"ip-protocol,omitempty"`
	// only the packet which match the tcp flag can be accepted, others will be dropped.
	// only set when the IPProtocol is tcp, used for partition.
	AcceptTCPFlags string `json:"accept-tcp-flags,omitempty"`
}

type NetworkDNSSpec struct {
	// update the DNS server in /etc/resolv.conf with this value
	DNSServer string `json:"dns-server,omitempty"`
	// map specified host to this IP address
	DNSIp string `json:"dns-ip,omitempty"`
	// map this host to specified IP
	DNSDomainName string `json:"dns-domain-name,omitempty"`
}

type NetworkBandwidthSpec struct {
	Rate string `json:"rate"`
	// +kubebuilder:validation:Minimum=1
	Limit uint32 `json:"limit"`
	// +kubebuilder:validation:Minimum=1
	Buffer uint32 `json:"buffer"`

	Peakrate *uint64 `json:"peakrate,omitempty"`
	Minburst *uint32 `json:"minburst,omitempty"`

	Device    string `json:"device,omitempty"`
	IPAddress string `json:"ip-address,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
}

type NetworkFloodSpec struct {
	// The speed of network traffic, allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second
	Rate string `json:"rate"`
	// Generate traffic to this IP address
	IPAddress string `json:"ip-address,omitempty"`
	// Generate traffic to this port on the IP address
	Port string `json:"port,omitempty"`
	// The number of iperf parallel client threads to run
	Parallel int32 `json:"parallel,omitempty"`
	// The number of seconds to run the iperf test
	Duration string `json:"duration"`
}

type NetworkDownSpec struct {
	// The network interface to impact
	Device string `json:"device,omitempty"`
	// NIC down time, time units: ns, us (or µs), ms, s, m, h.
	Duration string `json:"duration,omitempty"`
}

type ProcessSpec struct {
	// the process name or the process ID
	Process string `json:"process,omitempty"`
	// the signal number to send
	Signal int `json:"signal,omitempty"`

	// the command to be run when recovering experiment
	RecoverCmd string `json:"recoverCmd,omitempty"`
}

type JVMExceptionSpec struct {
	JVMCommonSpec      `json:",inline"`
	JVMClassMethodSpec `json:",inline"`

	// the exception which needs to throw for action `exception`
	ThrowException string `json:"exception,omitempty"`
}

type JVMStressSpec struct {
	JVMCommonSpec `json:",inline"`

	// the CPU core number need to use, only set it when action is stress
	CPUCount int `json:"cpu-count,omitempty"`

	// the memory type need to locate, only set it when action is stress, the value can be 'stack' or 'heap'
	MemoryType string `json:"mem-type,omitempty"`
}

type JVMGCSpec struct {
	JVMCommonSpec `json:",inline"`
}

type JVMLatencySpec struct {
	JVMCommonSpec      `json:",inline"`
	JVMClassMethodSpec `json:",inline"`

	// the latency duration for action 'latency', unit ms
	LatencyDuration int `json:"latency,omitempty"`
}

type JVMReturnSpec struct {
	JVMCommonSpec      `json:",inline"`
	JVMClassMethodSpec `json:",inline"`

	// the return value for action 'return'
	ReturnValue string `json:"value,omitempty"`
}

type JVMRuleDataSpec struct {
	JVMCommonSpec `json:",inline"`

	// RuleData used to save the rule file's data, will use it when recover
	RuleData string `json:"rule-data,omitempty"`
}

type PMJVMMySQLSpec struct {
	JVMCommonSpec `json:",inline"`

	JVMMySQLSpec `json:",inline"`

	// The exception which needs to throw for action `exception`
	// or the exception message needs to throw in action `mysql`
	ThrowException string `json:"exception,omitempty"`

	// The latency duration for action 'latency'
	// or the latency duration in action `mysql`
	LatencyDuration int `json:"latency,omitempty"`
}

type ClockSpec struct {
	// the pid of target program.
	Pid int `json:"pid,omitempty"`
	// specifies the length of time offset.
	TimeOffset string `json:"time-offset,omitempty"`
	// the identifier of the particular clock on which to act.
	// More clock description in linux kernel can be found in man page of clock_getres, clock_gettime, clock_settime.
	// Muti clock ids should be split with ","
	ClockIdsSlice string `json:"clock-ids-slice,omitempty"`
}

type RedisCommonSpec struct {
	// The adress of Redis server
	Addr string `json:"addr,omitempty"`
	// The password of Redis server
	Password string `json:"password,omitempty"`
}

type RedisExpirationSpec struct {
	RedisCommonSpec `json:",inline"`
	// The expiration of the keys
	Expiration string `json:"expiration,omitempty"`
	// The keys to be expired
	Key string `json:"key,omitempty"`
	// Additional options for `expiration`
	Option string `json:"option,omitempty"`
}

type RedisPenetrationSpec struct {
	RedisCommonSpec `json:",inline"`
	// The number of requests to be sent
	RequestNum int `json:"requestNum,omitempty"`
}

type RedisCacheLimitSpec struct {
	RedisCommonSpec `json:",inline"`
	// The size of `maxmemory`
	Size string `json:"cacheSize,omitempty"`
	// Specifies maxmemory as a percentage of the original value
	Percent string `json:"percent,omitempty"`
}

type RedisSentinelRestartSpec struct {
	RedisCommonSpec `json:",inline"`
	// The path of Sentinel conf
	Conf string `json:"conf,omitempty"`
	// The control flag determines whether to flush config
	FlushConfig bool `json:"flushConfig,omitempty"`
	// The path of `redis-server` command-line tool
	RedisPath bool `json:"redisPath,omitempty"`
}

type RedisSentinelStopSpec struct {
	RedisCommonSpec `json:",inline"`
	// The path of Sentinel conf
	Conf string `json:"conf,omitempty"`
	// The control flag determines whether to flush config
	FlushConfig bool `json:"flushConfig,omitempty"`
	// The path of `redis-server` command-line tool
	RedisPath bool `json:"redisPath,omitempty"`
}

type KafkaCommonSpec struct {
	// The topic to attack
	Topic string `json:"topic,omitempty"`
	// The host of kafka server
	Host string `json:"host,omitempty"`
	// The port of kafka server
	Port uint16 `json:"port,omitempty"`
	// The username of kafka client
	Username string `json:"username,omitempty"`
	// The password of kafka client
	Password string `json:"password,omitempty"`
}

type KafkaFillSpec struct {
	KafkaCommonSpec `json:",inline"`
	// The size of each message
	MessageSize uint `json:"messageSize,omitempty"`
	// The max bytes to fill
	MaxBytes uint64 `json:"maxBytes,omitempty"`
	// The command to reload kafka config
	ReloadCommand string `json:"reloadCommand,omitempty"`
}

type KafkaFloodSpec struct {
	KafkaCommonSpec `json:",inline"`
	// The size of each message
	MessageSize uint `json:"messageSize,omitempty"`
	// The number of worker threads
	Threads uint `json:"threads,omitempty"`
}

type KafkaIOSpec struct {
	// The topic to attack
	Topic string `json:"topic,omitempty"`
	// The path of server config
	ConfigFile string `json:"configFile,omitempty"`
	// Make kafka cluster non-readable
	NonReadable bool `json:"nonReadable,omitempty"`
	// Make kafka cluster non-writable
	NonWritable bool `json:"nonWritable,omitempty"`
}

type HTTPCommonSpec struct {
	// Composed with one of the port of HTTP connection, we will only attack HTTP connection with port inside proxy_ports
	ProxyPorts []uint `json:"proxy_ports"`
	// HTTP target: Request or Response
	Target string `json:"target"`
	// The TCP port that the target service listens on
	Port int32 `json:"port,omitempty"`
	// Match path of Uri with wildcard matches
	Path string `json:"path,omitempty"`
	// HTTP method
	Method string `json:"method,omitempty"`
	// Code is a rule to select target by http status code in response
	Code string `json:"code,omitempty"`
}

type HTTPAbortSpec struct {
	HTTPCommonSpec `json:",inline"`
}

type HTTPDelaySpec struct {
	HTTPCommonSpec `json:",inline"`
	// Delay represents the delay of the target request/response
	Delay string `json:"delay"`
}

type HTTPConfigSpec struct {
	// The config file path
	FilePath string `json:"file_path,omitempty"`
}

// used for HTTP request, now only support GET
type HTTPRequestSpec struct {
	// Request to send"
	URL string `json:"url,omitempty"`
	// Enable connection pool
	EnableConnPool bool `json:"enable-conn-pool,omitempty"`
	// The number of requests to send
	Count int `json:"count,omitempty"`
}

type FileCreateSpec struct {
	// FileName is the name of the file to be created, modified, deleted, renamed, or appended.
	FileName string `json:"file-name,omitempty"`
	// DirName is the directory name to create or delete.
	DirName string `json:"dir-name,omitempty"`
}

type FileModifyPrivilegeSpec struct {
	// FileName is the name of the file to be created, modified, deleted, renamed, or appended.
	FileName string `json:"file-name,omitempty"`
	// Privilege is the file privilege to be set.
	Privilege uint32 `json:"privilege,omitempty"`
}

type FileDeleteSpec struct {
	// FileName is the name of the file to be created, modified, deleted, renamed, or appended.
	FileName string `json:"file-name,omitempty"`
	// DirName is the directory name to create or delete.
	DirName string `json:"dir-name,omitempty"`
}

type FileRenameSpec struct {
	// SourceFile is the name need to be renamed.
	SourceFile string `json:"source-file,omitempty"`
	// DestFile is the name to be renamed.
	DestFile string `json:"dest-file,omitempty"`
}

type FileAppendSpec struct {
	// FileName is the name of the file to be created, modified, deleted, renamed, or appended.
	FileName string `json:"file-name,omitempty"`
	// Data is the data for append.
	Data string `json:"data,omitempty"`
	// Count is the number of times to append the data.
	Count int `json:"count,omitempty"`
}

type FileReplaceSpec struct {
	// FileName is the name of the file to be created, modified, deleted, renamed, or appended.
	FileName string `json:"file-name,omitempty"`
	// OriginStr is the origin string of the file.
	OriginStr string `json:"origin-string,omitempty"`
	// DestStr is the destination string of the file.
	DestStr string `json:"dest-string,omitempty"`
	// Line is the line number of the file to be replaced.
	Line int `json:"line,omitempty"`
}

type VMSpec struct {
	// The name of the VM to be injected
	VMName string `json:"vm-name,omitempty"`
}

type UserDefinedSpec struct {
	// The command to be executed when attack
	AttackCmd string `json:"attackCmd,omitempty"`
	// The command to be executed when recover
	RecoverCmd string `json:"recoverCmd,omitempty"`
}
