// Copyright 2026 Chaos Mesh Authors.
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

package podhttpchaos

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
)

func TestHTTPChaosTLSConfiguration(t *testing.T) {
	for _, mode := range []string{"with CA", "without CA", "without TLS"} {
		t.Run(mode, func(t *testing.T) {
			tlsConfig := testHTTPChaosTLS()
			if mode == "without CA" {
				tlsConfig.CAName = nil
			} else if mode == "without TLS" {
				tlsConfig = nil
			}
			f := newHTTPChaosFixture(t, tlsConfig)
			if tlsConfig != nil {
				secret := f.secret.DeepCopy()
				if tlsConfig.CAName == nil {
					delete(secret.Data, "root.crt")
				}
				require.NoError(t, f.client.Create(context.Background(), secret))
				decoy := &v1.Secret{Name: f.key.Name, Namespace: f.key.Namespace, Data: map[string][]byte{
					"proxy.crt": []byte("wrong cert"), "proxy.key": []byte("wrong key"), "root.crt": []byte("wrong CA"),
				}}
				require.NoError(t, f.client.Create(context.Background(), decoy))
			}
			f.reconcile(t, false)
			f.checkRequest(t, tlsConfig, 101, 201)
			f.reconcile(t, false)
			require.Empty(t, f.daemon.requests)
		})
	}
}

func TestHTTPChaosTLSFailureThenRepair(t *testing.T) {
	cases := []struct {
		name    string
		missing bool
		getErr  error
		dataKey string
		empty   bool
	}{
		{name: "missing Secret", missing: true},
		{name: "forbidden Secret", getErr: apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "proxy-tls", errors.New("denied"))},
		{name: "transient API error", getErr: apierrors.NewServiceUnavailable("temporarily unavailable")},
		{name: "missing certificate", dataKey: "proxy.crt"},
		{name: "empty certificate", dataKey: "proxy.crt", empty: true},
		{name: "missing private key", dataKey: "proxy.key"},
		{name: "empty private key", dataKey: "proxy.key", empty: true},
		{name: "missing explicit CA", dataKey: "root.crt"},
		{name: "empty explicit CA", dataKey: "root.crt", empty: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tlsConfig := testHTTPChaosTLS()
			f := newHTTPChaosFixture(t, tlsConfig)
			if !tc.missing {
				secret := f.secret.DeepCopy()
				if tc.dataKey != "" {
					delete(secret.Data, tc.dataKey)
					if tc.empty {
						secret.Data[tc.dataKey] = []byte{}
					}
				}
				require.NoError(t, f.client.Create(context.Background(), secret))
			}
			f.client.secretErr = tc.getErr
			for range 2 {
				status := f.reconcile(t, true)
				require.EqualValues(t, 101, status.Pid)
				require.EqualValues(t, 201, status.StartTime)
				require.Contains(t, status.FailedMessage, "certificates/proxy-tls")
				if tc.dataKey != "" {
					require.Contains(t, status.FailedMessage, tc.dataKey)
				}
				if tc.getErr != nil {
					require.Contains(t, status.FailedMessage, tc.getErr.Error())
				}
				require.Empty(t, f.daemon.requests)
				select {
				case event := <-f.recorder.Events:
					require.Contains(t, event, "Warning Failed")
					require.Contains(t, event, status.FailedMessage)
				default:
					t.Fatal("missing failure event")
				}
			}

			f.client.secretErr = nil
			if tc.missing {
				require.NoError(t, f.client.Create(context.Background(), f.secret.DeepCopy()))
			} else {
				secret := &v1.Secret{}
				require.NoError(t, f.client.Get(context.Background(), client.ObjectKeyFromObject(f.secret), secret))
				secret.Data = f.secret.DeepCopy().Data
				require.NoError(t, f.client.Update(context.Background(), secret))
			}
			f.reconcile(t, false)
			f.checkRequest(t, tlsConfig, 101, 201)
			f.reconcile(t, false)
			require.Empty(t, f.daemon.requests)
		})
	}
}

func TestHTTPChaosTLSRejectedThenSecretRepair(t *testing.T) {
	tlsConfig := testHTTPChaosTLS()
	f := newHTTPChaosFixture(t, tlsConfig)
	require.NoError(t, f.client.Create(context.Background(), f.secret.DeepCopy()))
	for range 2 {
		f.daemon.outcomes <- httpDaemonOutcome{response: &pb.ApplyHttpChaosResponse{
			StatusCode: http.StatusBadRequest, Error: "invalid TLS contents", Instance: 123, StartTime: 456,
		}}
	}
	for attempt := range 2 {
		got := f.reconcile(t, true)
		require.EqualValues(t, 123, got.Pid)
		require.EqualValues(t, 456, got.StartTime)
		require.EqualValues(t, 7, got.ObservedGeneration)
		require.Contains(t, got.FailedMessage, "status(400)")
		require.Contains(t, got.FailedMessage, "invalid TLS contents")
		if attempt == 0 {
			f.checkRequest(t, tlsConfig, 101, 201)
		} else {
			f.checkRequest(t, tlsConfig, 123, 456)
		}
	}

	secret := &v1.Secret{}
	require.NoError(t, f.client.Get(context.Background(), client.ObjectKeyFromObject(f.secret), secret))
	secret.Data = map[string][]byte{
		"proxy.crt": []byte("repaired certificate"), "proxy.key": []byte("repaired private key"), "root.crt": []byte("repaired CA"),
	}
	require.NoError(t, f.client.Update(context.Background(), secret))
	f.secret = secret
	got := f.reconcile(t, false)
	require.EqualValues(t, 7, got.ObservedGeneration)
	f.checkRequest(t, tlsConfig, 123, 456)
	f.reconcile(t, false)
	require.Empty(t, f.daemon.requests)
}

func TestHTTPChaosTLSRPCErrorKeepsIdentity(t *testing.T) {
	tlsConfig := testHTTPChaosTLS()
	f := newHTTPChaosFixture(t, tlsConfig)
	require.NoError(t, f.client.Create(context.Background(), f.secret.DeepCopy()))
	f.daemon.outcomes <- httpDaemonOutcome{err: status.Error(codes.Unavailable, "daemon disconnected")}
	got := f.reconcile(t, true)
	require.EqualValues(t, 101, got.Pid)
	require.EqualValues(t, 201, got.StartTime)
	require.Contains(t, got.FailedMessage, "daemon disconnected")
	f.checkRequest(t, tlsConfig, 101, 201)
	f.reconcile(t, false)
	f.checkRequest(t, tlsConfig, 101, 201)
}

func TestHTTPChaosTLSRejectedReconfigurationKeepsIdentity(t *testing.T) {
	tlsConfig := testHTTPChaosTLS()
	f := newHTTPChaosFixture(t, tlsConfig)
	require.NoError(t, f.client.Create(context.Background(), f.secret.DeepCopy()))
	f.reconcile(t, false)
	f.checkRequest(t, tlsConfig, 101, 201)

	obj := &v1alpha1.PodHttpChaos{}
	require.NoError(t, f.client.Get(context.Background(), f.key, obj))
	obj.Generation++
	obj.Spec.Rules[0].Actions.Delay = new(string)
	*obj.Spec.Rules[0].Actions.Delay = "1s"
	require.NoError(t, f.client.Update(context.Background(), obj))
	f.daemon.outcomes <- httpDaemonOutcome{response: &pb.ApplyHttpChaosResponse{
		StatusCode: http.StatusBadRequest, Error: "reconfiguration rejected", Instance: 123, StartTime: 456,
	}}
	got := f.reconcile(t, true)
	require.EqualValues(t, 123, got.Pid)
	require.EqualValues(t, 456, got.StartTime)
	require.EqualValues(t, 8, got.ObservedGeneration)
	require.Contains(t, got.FailedMessage, "reconfiguration rejected")
	f.checkRequest(t, tlsConfig, 123, 456)

	secret := &v1.Secret{}
	require.NoError(t, f.client.Get(context.Background(), client.ObjectKeyFromObject(f.secret), secret))
	secret.Data["proxy.crt"] = []byte("replacement certificate")
	require.NoError(t, f.client.Update(context.Background(), secret))
	f.secret = secret
	got = f.reconcile(t, false)
	require.EqualValues(t, 8, got.ObservedGeneration)
	f.checkRequest(t, tlsConfig, 123, 456)
	f.reconcile(t, false)
	require.Empty(t, f.daemon.requests)
}

type httpDaemonOutcome struct {
	response *pb.ApplyHttpChaosResponse
	err      error
}

type recordingHTTPDaemon struct {
	pb.UnimplementedChaosDaemonServer
	requests chan *pb.ApplyHttpChaosRequest
	outcomes chan httpDaemonOutcome
}

//revive:disable-next-line:var-naming Preserve the generated protobuf RPC contract.
func (d *recordingHTTPDaemon) ApplyHttpChaos(_ context.Context, req *pb.ApplyHttpChaosRequest) (*pb.ApplyHttpChaosResponse, error) {
	d.requests <- req
	select {
	case outcome := <-d.outcomes:
		return outcome.response, outcome.err
	default:
	}
	return &pb.ApplyHttpChaosResponse{StatusCode: http.StatusOK, Instance: 123, StartTime: 456}, nil
}

type secretErrorClient struct {
	client.Client
	secretErr error
}

func (c *secretErrorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*v1.Secret); ok && c.secretErr != nil {
		return c.secretErr
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

type httpChaosFixture struct {
	key        types.NamespacedName
	client     *secretErrorClient
	reconciler *Reconciler
	daemon     *recordingHTTPDaemon
	recorder   *record.FakeRecorder
	secret     *v1.Secret
}

func testHTTPChaosTLS() *v1alpha1.PodHttpChaosTLS {
	caName := "root.crt"
	return &v1alpha1.PodHttpChaosTLS{SecretName: "proxy-tls", SecretNamespace: "certificates", CertName: "proxy.crt", KeyName: "proxy.key", CAName: &caName}
}

func newHTTPChaosFixture(t *testing.T, tlsConfig *v1alpha1.PodHttpChaosTLS) *httpChaosFixture {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	daemon := &recordingHTTPDaemon{requests: make(chan *pb.ApplyHttpChaosRequest, 8), outcomes: make(chan httpDaemonOutcome, 8)}
	server := grpc.NewServer()
	pb.RegisterChaosDaemonServer(server, daemon)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	oldConfig := config.ControllerCfg
	testConfig := *oldConfig
	testConfig.Namespace = "chaos-mesh"
	testConfig.ChaosDaemonPort = listener.Addr().(*net.TCPAddr).Port
	testConfig.ChaosMeshCACert = ""
	config.ControllerCfg = &testConfig
	t.Cleanup(func() { config.ControllerCfg = oldConfig })

	scheme := runtime.NewScheme()
	require.NoError(t, v1.AddToScheme(scheme))
	require.NoError(t, discoveryv1.AddToScheme(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	key := types.NamespacedName{Namespace: "application", Name: "target-pod"}
	nodeName := "test-node"
	abort := true
	pod := &v1.Pod{Name: key.Name, Namespace: key.Namespace, Spec: v1.PodSpec{NodeName: nodeName}, Status: v1.PodStatus{
		ContainerStatuses: []v1.ContainerStatus{{Name: "app", ContainerID: "containerd://target"}},
	}}
	chaos := &v1alpha1.PodHttpChaos{
		Name: key.Name, Namespace: key.Namespace, Generation: 7,
		Spec: v1alpha1.PodHttpChaosSpec{TLS: tlsConfig, Rules: []v1alpha1.PodHttpChaosRule{{Port: 8443,
			Target: v1alpha1.PodHttpRequest, Actions: v1alpha1.PodHttpChaosActions{Abort: &abort}}}},
		Status: v1alpha1.PodHttpChaosStatus{Pid: 101, StartTime: 201},
	}
	endpoints := &discoveryv1.EndpointSlice{Name: "chaos-daemon-test", Namespace: testConfig.Namespace,
		AddressType: discoveryv1.AddressTypeIPv4, Endpoints: []discoveryv1.Endpoint{{Addresses: []string{"127.0.0.1"}, NodeName: &nodeName}}}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, chaos, endpoints).
		WithStatusSubresource(&v1alpha1.PodHttpChaos{}).Build()
	cli := &secretErrorClient{Client: fakeClient}
	recorder := record.NewFakeRecorder(8)
	return &httpChaosFixture{
		key: key, client: cli, daemon: daemon, recorder: recorder,
		reconciler: &Reconciler{Client: cli, Recorder: recorder, Log: logr.Discard(), ChaosDaemonClientBuilder: &chaosdaemon.ClientBuilder{Reader: fakeClient}},
		secret: &v1.Secret{Name: "proxy-tls", Namespace: "certificates", Data: map[string][]byte{
			"proxy.crt": {0, 45, 127, 128, 255}, "proxy.key": {255, 128, 127, 45, 0}, "root.crt": {45, 0, 255, 127, 128},
		}},
	}
}

func (f *httpChaosFixture) reconcile(t *testing.T, failed bool) v1alpha1.PodHttpChaosStatus {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := f.reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: f.key})
	require.NoError(t, err)
	require.Equal(t, failed, result.Requeue)
	require.Zero(t, result.RequeueAfter)
	obj := &v1alpha1.PodHttpChaos{}
	require.NoError(t, f.client.Get(ctx, f.key, obj))
	require.Equal(t, obj.Generation, obj.Status.ObservedGeneration)
	if failed {
		require.NotEmpty(t, obj.Status.FailedMessage)
	} else {
		require.Empty(t, obj.Status.FailedMessage)
		require.EqualValues(t, 123, obj.Status.Pid)
		require.EqualValues(t, 456, obj.Status.StartTime)
	}
	return obj.Status
}

func (f *httpChaosFixture) checkRequest(t *testing.T, tlsConfig *v1alpha1.PodHttpChaosTLS, pid, startTime int64) {
	t.Helper()
	require.Len(t, f.daemon.requests, 1)
	req := <-f.daemon.requests
	require.Equal(t, "containerd://target", req.ContainerId)
	require.Equal(t, []uint32{8443}, req.ProxyPorts)
	require.True(t, req.EnterNS)
	require.Equal(t, pid, req.Instance)
	require.Equal(t, startTime, req.StartTime)
	var rules []v1alpha1.PodHttpChaosBaseRule
	require.NoError(t, json.Unmarshal([]byte(req.Rules), &rules))
	require.Len(t, rules, 1)
	require.Equal(t, v1alpha1.PodHttpRequest, rules[0].Target)
	require.NotNil(t, rules[0].Actions.Abort)
	require.True(t, *rules[0].Actions.Abort)
	if tlsConfig == nil {
		require.Empty(t, req.Tls)
		return
	}
	var wire map[string]struct {
		Type  string `json:"type"`
		Value []int  `json:"value"`
	}
	require.NoError(t, json.Unmarshal([]byte(req.Tls), &wire))
	keys := map[string]string{"cert_file": "proxy.crt", "key_file": "proxy.key"}
	if tlsConfig.CAName != nil {
		keys["ca_file"] = "root.crt"
	}
	require.Len(t, wire, len(keys))
	for field, secretKey := range keys {
		item, ok := wire[field]
		require.True(t, ok, field)
		require.Equal(t, "Contents", item.Type)
		require.NotNil(t, item.Value)
		expected := make([]int, len(f.secret.Data[secretKey]))
		for i, value := range f.secret.Data[secretKey] {
			expected[i] = int(value)
		}
		require.Equal(t, expected, item.Value, field)
	}
	var got tproxyconfig.TLSConfig
	require.NoError(t, json.Unmarshal([]byte(req.Tls), &got))
	require.Equal(t, tproxyconfig.TLSConfigItem{Type: "Contents", Value: f.secret.Data["proxy.crt"]}, got.CertFile)
	require.Equal(t, tproxyconfig.TLSConfigItem{Type: "Contents", Value: f.secret.Data["proxy.key"]}, got.KeyFile)
	if tlsConfig.CAName == nil {
		require.Nil(t, got.CAFile)
	} else {
		require.Equal(t, &tproxyconfig.TLSConfigItem{Type: "Contents", Value: f.secret.Data["root.crt"]}, got.CAFile)
	}
}
