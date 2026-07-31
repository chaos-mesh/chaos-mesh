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

import "k8s.io/apimachinery/pkg/util/validation/field"

// Validate checks fault presence and route-specific abort status fields.
func (in *EnvoyGatewayChaosSpec) Validate(_ interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	faultPath := path.Child("fault")

	if in.Fault.Delay == nil && in.Fault.Abort == nil {
		allErrs = append(allErrs, field.Required(faultPath, "at least one of delay or abort must be configured"))
	}

	if in.Fault.Abort == nil {
		return allErrs
	}

	abortPath := faultPath.Child("abort")
	hasHTTPStatus := in.Fault.Abort.HTTPStatus != nil
	hasGRPCStatus := in.Fault.Abort.GRPCStatus != nil
	if hasHTTPStatus == hasGRPCStatus {
		allErrs = append(allErrs, field.Invalid(abortPath, in.Fault.Abort, "exactly one of httpStatus or grpcStatus must be configured"))
		return allErrs
	}

	switch in.Target.Kind {
	case EnvoyGatewayHTTPRoute:
		if hasGRPCStatus {
			allErrs = append(allErrs, field.Invalid(abortPath.Child("grpcStatus"), *in.Fault.Abort.GRPCStatus, "grpcStatus cannot be used with HTTPRoute"))
		}
	case EnvoyGatewayGRPCRoute:
		if hasHTTPStatus {
			allErrs = append(allErrs, field.Invalid(abortPath.Child("httpStatus"), *in.Fault.Abort.HTTPStatus, "httpStatus cannot be used with GRPCRoute"))
		}
	}

	return allErrs
}
