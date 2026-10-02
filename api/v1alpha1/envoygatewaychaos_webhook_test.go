// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestEnvoyGatewayChaosSpecValidation(t *testing.T) {
	httpStatus := int32(503)
	grpcStatus := int32(14)
	path := field.NewPath("spec")

	tests := []struct {
		name string
		spec EnvoyGatewayChaosSpec
		ok   bool
	}{
		{
			name: "delay",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayHTTPRoute},
				Fault:  EnvoyGatewayFault{Delay: &EnvoyGatewayDelay{FixedDelay: "250ms", Percentage: 10}},
			},
			ok: true,
		},
		{
			name: "http abort",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayHTTPRoute},
				Fault:  EnvoyGatewayFault{Abort: &EnvoyGatewayAbort{HTTPStatus: &httpStatus, Percentage: 10}},
			},
			ok: true,
		},
		{
			name: "grpc abort",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayGRPCRoute},
				Fault:  EnvoyGatewayFault{Abort: &EnvoyGatewayAbort{GRPCStatus: &grpcStatus, Percentage: 10}},
			},
			ok: true,
		},
		{name: "missing fault", spec: EnvoyGatewayChaosSpec{Target: EnvoyGatewayTarget{Kind: EnvoyGatewayHTTPRoute}}},
		{
			name: "missing abort status",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayHTTPRoute},
				Fault:  EnvoyGatewayFault{Abort: &EnvoyGatewayAbort{Percentage: 10}},
			},
		},
		{
			name: "both abort statuses",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayHTTPRoute},
				Fault:  EnvoyGatewayFault{Abort: &EnvoyGatewayAbort{HTTPStatus: &httpStatus, GRPCStatus: &grpcStatus, Percentage: 10}},
			},
		},
		{
			name: "grpc status on http route",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayHTTPRoute},
				Fault:  EnvoyGatewayFault{Abort: &EnvoyGatewayAbort{GRPCStatus: &grpcStatus, Percentage: 10}},
			},
		},
		{
			name: "http status on grpc route",
			spec: EnvoyGatewayChaosSpec{
				Target: EnvoyGatewayTarget{Kind: EnvoyGatewayGRPCRoute},
				Fault:  EnvoyGatewayFault{Abort: &EnvoyGatewayAbort{HTTPStatus: &httpStatus, Percentage: 10}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			errs := test.spec.Validate(nil, path)
			if test.ok {
				g.Expect(errs).To(BeEmpty())
			} else {
				g.Expect(errs).NotTo(BeEmpty())
			}
		})
	}
}
