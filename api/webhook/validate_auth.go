// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authv1 "k8s.io/api/authorization/v1"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

var authLog = ctrl.Log.WithName("validate-auth")

// +kubebuilder:webhook:path=/validate-auth,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=*,verbs=create;update,versions=v1alpha1,name=vauth.kb.io

// AuthValidator validates the authority
type AuthValidator struct {
	Enabled bool
	Client  client.Client
	Reader  client.Reader
	AuthCli *authorizationv1.AuthorizationV1Client

	decoder *admission.Decoder

	ClusterScoped     bool
	TargetNamespace   string
	AllowedNamespaces string
	IgnoredNamespaces string
}

// AuthValidator admits a pod iff a specific annotation exists.
func (v *AuthValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if !v.Enabled {
		return admission.Allowed("")
	}

	username := req.UserInfo.Username
	groups := req.UserInfo.Groups
	chaosKind := req.Kind.Kind

	//var chaos v1alpha1.ChaosValidator

	if chaosKind == v1alpha1.KindAwsChaos || chaosKind == v1alpha1.KindPodNetworkChaos {
		return admission.Allowed("")
	}

	chaos := v1alpha1.GetChaosValidator(chaosKind)
	oldChaos := v1alpha1.GetChaosValidator(chaosKind)
	if chaos == nil {
		err := fmt.Errorf("kind %s is not support", chaosKind)
		return admission.Errored(http.StatusBadRequest, err)
	}

	err := v.decoder.Decode(req, chaos)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	specs := chaos.GetSelectSpec()

	if req.Operation == admissionv1beta1.Update {
		// when selector is not changed, don't need to validate auth again

		err = v.decoder.DecodeRaw(req.OldObject, oldChaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		oldSpecs := oldChaos.GetSelectSpec()
		selectSpecChanged := false
		if len(specs) == len(oldSpecs) {
			for i, spec := range specs {
				if !reflect.DeepEqual(oldSpecs[i].GetSelector(), spec.GetSelector()) {
					selectSpecChanged = true
					break
				}
			}
		} else {
			selectSpecChanged = true
		}

		if !selectSpecChanged {
			authLog.Info("chaos updated but select spec not changed, auth validate passed")
			return admission.Allowed("")
		}
	}

	affectedNamespaces := make(map[string]struct{})

	for _, spec := range specs {
		pods, err := selector.SelectPods(context.Background(), v.Client, v.Reader, spec.GetSelector(), v.ClusterScoped, v.TargetNamespace, v.AllowedNamespaces, v.IgnoredNamespaces)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		for _, pod := range pods {
			affectedNamespaces[pod.Namespace] = struct{}{}
		}

		// may not exist pod under selector namespace, but still need to validate the privileges
		for _, namespace := range spec.GetSelector().Namespaces {
			affectedNamespaces[namespace] = struct{}{}
		}
	}

	for namespace := range affectedNamespaces {
		allow, err := v.auth(username, groups, namespace)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if !allow {
			return admission.Denied(fmt.Sprintf("%s is forbidden on namespace %s", username, namespace))
		}
	}

	authLog.Info("user have the privileges on all namespace, auth validate passed", "user", username, "groups", groups, "namespaces", affectedNamespaces)
	return admission.Allowed("")
}

// AuthValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *AuthValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

func (v *AuthValidator) auth(username string, groups []string, namespace string) (bool, error) {
	sar := authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "create",
				Group:     "chaos-mesh.org",
				Resource:  "*",
			},
			User:   username,
			Groups: groups,
		},
	}

	response, err := v.AuthCli.SubjectAccessReviews().Create(&sar)
	if err != nil {
		return false, err
	}

	return response.Status.Allowed, nil
}
