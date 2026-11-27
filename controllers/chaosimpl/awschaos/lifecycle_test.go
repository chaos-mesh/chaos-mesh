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

package awschaos_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos/detachvolume"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos/ec2restart"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/awschaos/ec2stop"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/aws"
)

func TestEC2ActionsAndRecovery(t *testing.T) {
	for _, scenario := range []struct {
		name, apply, recover string
	}{
		{"stop", "StopInstances", "StartInstances"},
		{"restart", "RebootInstances", ""},
		{"detach", "DetachVolume", "AttachVolume"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			g := NewWithT(t)
			requests := make(chan url.Values, 8)
			reject := make(chan string, 1)
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.ParseForm() != nil || r.Header.Get("Authorization") == "" || r.Header.Get("X-Amz-Security-Token") != "fixture-token" {
					http.Error(w, "invalid EC2 request", http.StatusBadRequest)
					return
				}
				requests <- r.PostForm
				action := r.PostForm.Get("Action")
				select {
				case fail := <-reject:
					if action != fail {
						http.Error(w, "unexpected action", http.StatusBadRequest)
						return
					}
					w.Header().Set("Content-Type", "text/xml")
					w.WriteHeader(http.StatusForbidden)
					_, _ = fmt.Fprint(w, "<Response><Errors><Error><Code>UnauthorizedOperation</Code><Message>fixture denied</Message></Error></Errors></Response>")
					return
				default:
				}
				w.Header().Set("Content-Type", "text/xml")
				_, _ = fmt.Fprintf(w, "<%sResponse xmlns=\"http://ec2.amazonaws.com/doc/2016-11-15/\"><requestId>fixture</requestId><return>true</return></%sResponse>", action, action)
			}))
			t.Cleanup(api.Close)
			t.Setenv("AWS_ENDPOINT_URL_EC2", api.URL)
			t.Setenv("AWS_IGNORE_CONFIGURED_ENDPOINT_URLS", "false")
			t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
			t.Setenv("AWS_MAX_ATTEMPTS", "1")
			t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/config")
			t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/credentials")

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "credentials", Namespace: "fixture"},
				Data: map[string][]byte{
					"aws_access_key_id": []byte("fixture-access"), "aws_secret_access_key": []byte("fixture-secret"),
					"aws_session_token": []byte("fixture-token"),
				},
			}
			kube := fake.NewClientBuilder().WithObjects(secret).Build()
			implementations := map[string]impltypes.ChaosImpl{
				"stop": ec2stop.NewImpl(kube, logr.Discard()), "restart": ec2restart.NewImpl(kube, logr.Discard()),
				"detach": detachvolume.NewImpl(kube, logr.Discard()),
			}
			impl := implementations[scenario.name]
			selected := v1alpha1.AWSSelector{AWSRegion: "us-east-1", Ec2Instance: "i-fixture",
				EbsVolume: new("vol-fixture"), DeviceName: new("/dev/sdf")}
			chaos := &v1alpha1.AWSChaos{ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace}}
			chaos.Spec.AWSSelector = selected
			if scenario.name == "stop" {
				selected.Endpoint = &api.URL
			}
			chaos.Spec.SecretName = &secret.Name
			targets, err := aws.New().Select(t.Context(), &selected)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(targets).To(HaveLen(1))
			records := []*v1alpha1.Record{{Id: targets[0].Id()}}

			checkRequest := func(action string) {
				t.Helper()
				var sent url.Values
				g.Expect(requests).To(Receive(&sent))
				g.Expect(sent.Get("Action")).To(Equal(action))
				if scenario.name == "detach" {
					g.Expect(sent.Get("InstanceId")).To(Equal(selected.Ec2Instance))
					g.Expect(sent.Get("VolumeId")).To(Equal("vol-fixture"))
					g.Expect(sent.Get("Device")).To(Equal("/dev/sdf"))
					if action == "DetachVolume" {
						g.Expect(sent.Get("Force")).To(Equal("true"))
					}
				} else {
					g.Expect(sent.Get("InstanceId.1")).To(Equal(selected.Ec2Instance))
				}
			}

			reject <- scenario.apply
			phase, err := impl.Apply(t.Context(), 0, records, chaos)
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			g.Expect(err).To(MatchError(ContainSubstring("fixture denied")))
			checkRequest(scenario.apply)
			g.Expect(impl.Apply(t.Context(), 0, records, chaos)).To(Equal(v1alpha1.Injected))
			checkRequest(scenario.apply)

			if scenario.recover != "" {
				reject <- scenario.recover
				phase, err = impl.Recover(t.Context(), 0, records, chaos)
				g.Expect(phase).To(Equal(v1alpha1.Injected))
				g.Expect(err).To(MatchError(ContainSubstring("fixture denied")))
				checkRequest(scenario.recover)
			}
			g.Expect(impl.Recover(t.Context(), 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
			if scenario.recover != "" {
				checkRequest(scenario.recover)
			}
			g.Expect(requests).To(BeEmpty())

			invalid := []*v1alpha1.Record{{Id: "invalid"}}
			_, err = impl.Apply(t.Context(), 0, invalid, chaos)
			g.Expect(err).To(HaveOccurred())
			if scenario.recover != "" {
				_, err = impl.Recover(t.Context(), 0, invalid, chaos)
				g.Expect(err).To(HaveOccurred())
			}
			chaos.Spec.SecretName = new("missing")
			phase, err = impl.Apply(t.Context(), 0, records, chaos)
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
			if scenario.recover != "" {
				phase, err = impl.Recover(t.Context(), 0, records, chaos)
				g.Expect(phase).To(Equal(v1alpha1.Injected))
				g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
			}
			g.Expect(requests).To(BeEmpty())
		})
	}
}
