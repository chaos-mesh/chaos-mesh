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

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var alwaysAllowedKind = []string{
	v1alpha1.KindAWSChaos,
	v1alpha1.KindPodNetworkChaos,
	v1alpha1.KindPodIOChaos,
	v1alpha1.KindGCPChaos,
	v1alpha1.KindPodHttpChaos,
	v1alpha1.KindPhysicalMachine,
	v1alpha1.KindStatusCheck,
	v1alpha1.KindRemoteCluster,

	"WorkflowNode",
}

// +kubebuilder:webhook:path=/validate-auth,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=*,verbs=create;update,versions=v1alpha1,name=vauth.kb.io

// AuthValidator validates the authority
type AuthValidator struct {
	enabled bool
	authCli *authorizationv1.AuthorizationV1Client

	decoder *admission.Decoder

	clusterScoped         bool
	targetNamespace       string
	enableFilterNamespace bool
	logger                logr.Logger
}

// NewAuthValidator returns a new AuthValidator
func NewAuthValidator(enabled bool, authCli *authorizationv1.AuthorizationV1Client,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool, logger logr.Logger) *AuthValidator {
	return &AuthValidator{
		enabled:               enabled,
		authCli:               authCli,
		clusterScoped:         clusterScoped,
		targetNamespace:       targetNamespace,
		enableFilterNamespace: enableFilterNamespace,
		logger:                logger,
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

	kind, ok := v1alpha1.AllKindsIncludeScheduleAndWorkflow()[requestKind]
	if !ok {
		err := errors.Wrapf(errInvalidValue, "kind %s is not support", requestKind)
		return admission.Errored(http.StatusBadRequest, err)
	}
	chaos := kind.SpawnObject()

	err := v.decoder.Decode(req, chaos)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	requireClusterPrivileges, affectedNamespaces := affectedNamespaces(chaos)

	if requireClusterPrivileges {
		allow, err := v.auth(username, groups, "", requestKind)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if !allow {
			return admission.Denied(fmt.Sprintf("%s is forbidden on cluster", username))
		}
		v.logger.Info("user have the privileges on cluster, auth validate passed", "user", username, "groups", groups, "namespace", affectedNamespaces)
	} else {
		v.logger.Info("start validating user", "user", username, "groups", groups, "namespace", affectedNamespaces)

		for namespace := range affectedNamespaces {
			allow, err := v.auth(username, groups, namespace, requestKind)
			if err != nil {
				return admission.Errored(http.StatusBadRequest, err)
			}

			if !allow {
				return admission.Denied(fmt.Sprintf("%s is forbidden on namespace %s", username, namespace))
			}
		}

		v.logger.Info("user have the privileges on namespace, auth validate passed", "user", username, "groups", groups, "namespace", affectedNamespaces)
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

	// FIXME: get context from parameter
	response, err := v.authCli.SubjectAccessReviews().Create(context.TODO(), &sar, metav1.CreateOptions{})
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
