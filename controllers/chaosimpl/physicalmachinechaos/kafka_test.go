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

package physicalmachinechaos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func TestKafkaPayload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kafka request payload")
}

var _ = Describe("Kafka requests", func() {
	DescribeTable("preserves broker and action settings", func(action, settings, expected string) {
		secure := config.ControllerCfg.ChaosdSecurityMode
		config.ControllerCfg.ChaosdSecurityMode = false
		DeferCleanup(func() { config.ControllerCfg.ChaosdSecurityMode = secure })

		payloads := make(chan []byte, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Equal(http.MethodPost))
			Expect(r.URL.Path).To(Equal("/api/attack/kafka"))
			body, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			payloads <- body
			w.WriteHeader(http.StatusOK)
		}))
		DeferCleanup(server.Close)

		obj := &v1alpha1.PhysicalMachineChaos{}
		specification := fmt.Sprintf(`{"action":"kafka-%s","uid":"kafka-%s-test","kafka-%s":%s}`, action, action, action, settings)
		Expect(json.Unmarshal([]byte(specification), &obj.Spec)).To(Succeed())
		obj.Spec.Address = []string{server.URL}
		var logs bytes.Buffer
		impl := &Impl{Log: zap.New(zap.WriteTo(&logs))}
		phase, err := impl.Apply(context.Background(), 0, []*v1alpha1.Record{{Id: server.URL}}, obj)
		Expect(err).NotTo(HaveOccurred())
		Expect(phase).To(Equal(v1alpha1.Injected))
		var payload []byte
		Expect(payloads).To(Receive(&payload))
		Expect(payload).To(MatchJSON(expected))
		Expect(logs.String()).To(ContainSubstring("HTTP request"))
		Expect(logs.String()).To(ContainSubstring(server.URL))
		Expect(logs.String()).NotTo(ContainSubstring("kafka-test-password"))
	},
		Entry("authenticated flood", "flood",
			`{"topic":"orders","host":"broker","port":9092,"username":"worker","password":"kafka-test-password","authMechanism":"sasl/plain","messageSize":128,"threads":2}`,
			`{"uid":"kafka-flood-test","action":"flood","topic":"orders","host":"broker","port":9092,"username":"worker","password":"kafka-test-password","authMechanism":"sasl/plain","messageSize":128,"threads":2}`),
		Entry("fill config", "fill",
			`{"topic":"orders","host":"broker","port":9092,"username":"worker","password":"kafka-test-password","authMechanism":"sasl/scram-sha-256","configFile":"/etc/kafka/server.properties","reloadCommand":"systemctl reload kafka","messageSize":128,"maxBytes":4096}`,
			`{"uid":"kafka-fill-test","action":"fill","topic":"orders","host":"broker","port":9092,"username":"worker","password":"kafka-test-password","authMechanism":"sasl/scram-sha-256","configFile":"/etc/kafka/server.properties","reloadCommand":"systemctl reload kafka","messageSize":128,"maxBytes":4096}`),
		Entry("IO broker discovery", "io",
			`{"topic":"orders","host":"broker","port":9092,"username":"worker","password":"kafka-test-password","authMechanism":"sasl/scram-sha-512","configFile":"/etc/kafka/server.properties","nonReadable":true,"nonWritable":true}`,
			`{"uid":"kafka-io-test","action":"io","topic":"orders","host":"broker","port":9092,"username":"worker","password":"kafka-test-password","authMechanism":"sasl/scram-sha-512","configFile":"/etc/kafka/server.properties","nonReadable":true,"nonWritable":true}`),
	)
})
