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

package dnschaos

import (
	"context"
	"net"
	"strings"
	"testing"

	dnspb "github.com/chaos-mesh/k8s_dns_chaos/pb"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type recordingDNSServer struct {
	dnspb.UnimplementedDNSServer
	requests chan *dnspb.SetDNSChaosRequest
}

func (s *recordingDNSServer) SetDNSChaos(_ context.Context, request *dnspb.SetDNSChaosRequest) (*dnspb.DNSChaosResponse, error) {
	s.requests <- request
	return &dnspb.DNSChaosResponse{Result: true}, nil
}

func TestSetDNSServerRulesStaticRPC(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	service := &recordingDNSServer{requests: make(chan *dnspb.SetDNSChaosRequest, 3)}
	server := grpc.NewServer()
	dnspb.RegisterDNSServer(server, service)
	t.Cleanup(server.Stop)
	go server.Serve(listener)
	impl := &Impl{Log: logr.Discard()}
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "client", Namespace: "default"}}
	port := listener.Addr().(*net.TCPAddr).Port
	err = impl.setDNSServerRules("127.0.0.1", port, "static-rule", pod, v1alpha1.StaticAction, []string{"example.com", "v6.example.com"}, []string{"V6.EXAMPLE.COM.=2001:0db8::1", "Example.COM=192.0.2.10"})
	if err != nil {
		t.Fatal(err)
	}
	got := <-service.requests
	want := &dnspb.SetDNSChaosRequest{
		Name: "static-rule", Action: "static", Pods: []*dnspb.Pod{{Name: "client", Namespace: "default"}}, Patterns: []string{"example.com", "v6.example.com"},
		IpDomainMaps: []*dnspb.IpDomainMap{{Domain: "example.com", Ip: "192.0.2.10"}, {Domain: "v6.example.com", Ip: "2001:db8::1"}},
	}
	if !proto.Equal(got, want) {
		t.Fatalf("request = %v, want %v", got, want)
	}
	for _, action := range []v1alpha1.DNSChaosAction{v1alpha1.ErrorAction, v1alpha1.RandomAction} {
		if err := impl.setDNSServerRules("127.0.0.1", port, "existing-rule", pod, action, nil, []string{"ignored legacy value"}); err != nil {
			t.Fatal(err)
		}
		if request := <-service.requests; len(request.IpDomainMaps) != 0 || request.Action != string(action) {
			t.Fatalf("nonstatic request = %v", request)
		}
	}
}

func TestSetDNSServerRulesRejectsInvalidMappingsBeforeRPC(t *testing.T) {
	impl := &Impl{Log: logr.Discard()}
	for _, mappings := range [][]string{nil, {"example.com"}, {"example.com=bad"}, {"Example.com.=192.0.2.1", "example.com=192.0.2.2"}} {
		err := impl.setDNSServerRules("invalid-server-address", 1, "invalid-rule", nil, v1alpha1.StaticAction, nil, mappings)
		if err == nil || !strings.Contains(err.Error(), "domainIpMapping") {
			t.Fatalf("mappings %v returned %v", mappings, err)
		}
	}
}
