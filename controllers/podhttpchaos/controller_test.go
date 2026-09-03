// Copyright Chaos Mesh Authors.
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
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/httpchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tproxyconfig"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type httpChaosDaemon struct {
	pb.ChaosDaemonClient
	requests []*pb.ApplyHttpChaosRequest
}

func (d *httpChaosDaemon) Close() error { return nil }

func (d *httpChaosDaemon) ApplyHttpChaos(_ context.Context, req *pb.ApplyHttpChaosRequest, _ ...grpc.CallOption) (*pb.ApplyHttpChaosResponse, error) {
	d.requests = append(d.requests, req)
	return &pb.ApplyHttpChaosResponse{Instance: 123, StartTime: 456, StatusCode: http.StatusOK}, nil
}

func TestReconcileMissingTLSData(t *testing.T) {
	for _, missingKey := range []string{"tls.crt", "tls.key", "ca.crt"} {
		t.Run(missingKey, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			require.NoError(t, v1.AddToScheme(scheme))
			require.NoError(t, v1alpha1.AddToScheme(scheme))

			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "target"},
				Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{{
					Name: "app", ContainerID: "containerd://app",
				}}},
			}
			abort := true
			child := &v1alpha1.PodHttpChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: pod.Namespace, Name: pod.Name, Generation: 2},
				Spec: v1alpha1.PodHttpChaosSpec{
					TLS: &v1alpha1.PodHttpChaosTLS{
						SecretNamespace: pod.Namespace, SecretName: pod.Name,
						CertName: "tls.crt", KeyName: "tls.key",
					},
					Rules: []v1alpha1.PodHttpChaosRule{{
						Source: "default/experiment", Port: 443,
						PodHttpChaosBaseRule: v1alpha1.PodHttpChaosBaseRule{
							Target:  v1alpha1.PodHttpRequest,
							Actions: v1alpha1.PodHttpChaosActions{Abort: &abort},
						},
					}},
				},
			}
			if missingKey == "ca.crt" {
				child.Spec.TLS.CAName = &missingKey
			}
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Namespace: pod.Namespace, Name: pod.Name},
				Data: map[string][]byte{
					"tls.crt": []byte("certificate"), "tls.key": []byte("private key"), "ca.crt": []byte("CA certificate"),
				},
			}
			missingValue := secret.Data[missingKey]
			delete(secret.Data, missingKey)
			c := fake.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(&v1alpha1.PodHttpChaos{}).
				WithObjects(pod, child, secret).Build()
			daemon := &httpChaosDaemon{}
			defer mock.With("MockChaosDaemonClient", daemon)()
			require.NotNil(t, mock.On("MockChaosDaemonClient"), "enable pkg/mock failpoints before running this test")
			r := &Reconciler{
				Client: c, Log: logr.Discard(), Recorder: record.NewFakeRecorder(10),
				ChaosDaemonClientBuilder: &chaosdaemon.ChaosDaemonClientBuilder{},
			}
			request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(child)}
			parent := &v1alpha1.HTTPChaos{
				Status: v1alpha1.HTTPChaosStatus{Instances: map[string]int64{"default/target": child.Generation}},
			}
			impl := httpchaos.NewImpl(c, nil, logr.Discard()).Impl
			records := []*v1alpha1.Record{{Id: "default/target", Phase: "Not Injected/Wait"}}

			for attempt := 0; attempt < 2; attempt++ {
				require.NotPanics(t, func() {
					result, err := r.Reconcile(ctx, request)
					require.NoError(t, err)
					require.True(t, result.Requeue, "Secret updates are not watched, so failed injection must retry")
				})
				require.NoError(t, c.Get(ctx, request.NamespacedName, child))
				require.Contains(t, child.Status.FailedMessage, missingKey)
				require.Equal(t, child.Generation, child.Status.ObservedGeneration)
				require.Empty(t, daemon.requests)
				phase, err := impl.Apply(ctx, 0, records, parent)
				require.Error(t, err)
				require.NotEqual(t, v1alpha1.Injected, phase)
			}

			// Repairing only the Secret must allow the same generation to be injected.
			require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(secret), secret))
			secret.Data[missingKey] = missingValue
			require.NoError(t, c.Update(ctx, secret))
			result, err := r.Reconcile(ctx, request)
			require.NoError(t, err)
			require.Equal(t, ctrl.Result{}, result)
			require.Len(t, daemon.requests, 1)
			require.Equal(t, "containerd://app", daemon.requests[0].ContainerId)
			var tlsConfig tproxyconfig.TLSConfig
			require.NoError(t, json.Unmarshal([]byte(daemon.requests[0].Tls), &tlsConfig))
			require.Equal(t, secret.Data["tls.crt"], tlsConfig.CertFile.Value)
			require.Equal(t, secret.Data["tls.key"], tlsConfig.KeyFile.Value)
			if child.Spec.TLS.CAName != nil {
				require.NotNil(t, tlsConfig.CAFile)
				require.Equal(t, secret.Data["ca.crt"], tlsConfig.CAFile.Value)
			} else {
				require.Nil(t, tlsConfig.CAFile)
			}
			require.NoError(t, c.Get(ctx, request.NamespacedName, child))
			require.Empty(t, child.Status.FailedMessage)
			require.EqualValues(t, 123, child.Status.Pid)
			require.EqualValues(t, 456, child.Status.StartTime)
			phase, err := impl.Apply(ctx, 0, records, parent)
			require.NoError(t, err)
			require.Equal(t, v1alpha1.Injected, phase)

			_, err = r.Reconcile(ctx, request)
			require.NoError(t, err)
			require.Len(t, daemon.requests, 1)
		})
	}
}
