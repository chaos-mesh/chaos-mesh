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
	"strings"

	authv1 "k8s.io/api/authorization/v1"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

var alwaysAllowedKind = []string{
	v1alpha1.KindAwsChaos,
	v1alpha1.KindPodNetworkChaos,
	v1alpha1.KindPodIOChaos,
	v1alpha1.KindGcpChaos,
	v1alpha1.KindPodHttpChaos,

	// TODO: check the auth for Schedule
	// The resouce will be created by the SA of controller-manager, so checking the auth of Schedule is needed.
	v1alpha1.KindSchedule,

	"Workflow",
	"WorkflowNode",
}

var authLog = ctrl.Log.WithName("validate-auth")

// +kubebuilder:webhook:path=/validate-auth,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=*,verbs=create;update,versions=v1alpha1,name=vauth.kb.io

// AuthValidator validates the authority
type AuthValidator struct {
	enabled bool
	authCli *authorizationv1.AuthorizationV1Client

	decoder *admission.Decoder

	clusterScoped         bool
	targetNamespace       string
	enableFilterNamespace bool
}

// NewAuthValidator returns a new AuthValidator
func NewAuthValidator(enabled bool, authCli *authorizationv1.AuthorizationV1Client,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) *AuthValidator {
	return &AuthValidator{
		enabled:               enabled,
		authCli:               authCli,
		clusterScoped:         clusterScoped,
		targetNamespace:       targetNamespace,
		enableFilterNamespace: enableFilterNamespace,
	}
}

// AuthValidator admits a pod iff a specific annotation exists.
func (v *AuthValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if !v.enabled {
		return admission.Allowed("")
	}

	username := req.UserInfo.Username
	groups := req.UserInfo.Groups
	requestKind := req.Kind.Kind

	if contains(alwaysAllowedKind, requestKind) {
		return admission.Allowed(fmt.Sprintf("skip the RBAC check for type %s", requestKind))
	}

	kind, ok := v1alpha1.AllKinds()[requestKind]
	if !ok {
		err := fmt.Errorf("kind %s is not support", requestKind)
		return admission.Errored(http.StatusBadRequest, err)
	}
	chaos := kind.Chaos.DeepCopyObject().(common.InnerObjectWithSelector)
	if chaos == nil {
		err := fmt.Errorf("kind %s is not support", requestKind)
		return admission.Errored(http.StatusBadRequest, err)
	}

	err := v.decoder.Decode(req, chaos)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	specs := chaos.GetSelectorSpecs()

	requireClusterPrivileges := false
	affectedNamespaces := make(map[string]struct{})

	for _, spec := range specs {
		var selector *v1alpha1.PodSelector
		if s, ok := spec.(*v1alpha1.ContainerSelector); ok {
			selector = &s.PodSelector
		}
		if p, ok := spec.(*v1alpha1.PodSelector); ok {
			selector = p
		}
		if selector == nil {
			return admission.Allowed("")
		}

		if selector.Selector.ClusterScoped() {
			requireClusterPrivileges = true
		}

		for _, namespace := range selector.Selector.AffectedNamespaces() {
			affectedNamespaces[namespace] = struct{}{}
		}
	}

	if requireClusterPrivileges {
		allow, err := v.auth(username, groups, "", requestKind)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if !allow {
			return admission.Denied(fmt.Sprintf("%s is forbidden on cluster", username))
		}
		authLog.Info("user have the privileges on cluster, auth validate passed", "user", username, "groups", groups, "namespace", affectedNamespaces)
	} else {
		for namespace := range affectedNamespaces {
			allow, err := v.auth(username, groups, namespace, requestKind)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}

			if !allow {
				return admission.Denied(fmt.Sprintf("%s is forbidden on namespace %s", username, namespace))
			}
		}

		authLog.Info("user have the privileges on namespace, auth validate passed", "user", username, "groups", groups, "namespace", affectedNamespaces)
	}

	return admission.Allowed("")
}

// AuthValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *AuthValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

func (v *AuthValidator) auth(username string, groups []string, namespace string, chaosKind string) (bool, error) {
	resourceName, err := v.resourceFor(chaosKind)
	if err != nil {
		return false, err
	}
	sar := authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "create",
				Group:     "chaos-mesh.org",
				Resource:  resourceName,
			},
			User:   username,
			Groups: groups,
		},
	}

	response, err := v.authCli.SubjectAccessReviews().Create(&sar)
	if err != nil {
		return false, err
	}

	return response.Status.Allowed, nil
}

func (v *AuthValidator) resourceFor(name string) (string, error) {
	// TODO: we should use RESTMapper, but it relates to many dependencies
	return strings.ToLower(name), nil
}

func contains(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}
