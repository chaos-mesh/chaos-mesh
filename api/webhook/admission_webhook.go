/*
Copyright 2018 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhook

import (
	"context"
	"fmt"
	"net/http"

	//corev1 "k8s.io/api/core/v1"
	authv1 "k8s.io/api/authorization/v1"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

var log2 = ctrl.Log.WithName("validate-webhook")

// +kubebuilder:webhook:path=/validate-auth,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=*,verbs=create;update,versions=v1alpha1,name=vauth.kb.io

// AuthValidator validates the authority
type AuthValidator struct {
	Client client.Client
	Reader client.Reader

	decoder *admission.Decoder

	ClusterScoped     bool
	TargetNamespace   string
	AllowedNamespaces string
	IgnoredNamespaces string
}

// AuthValidator admits a pod iff a specific annotation exists.
func (v *AuthValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log2.Info("Get request from chaos mesh:", "req", req)
	/*
		pod := &corev1.Pod{}

		err := v.decoder.Decode(req, pod)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		key := "example-mutating-admission-webhook"
		anno, found := pod.Annotations[key]
		if !found {
			return admission.Denied(fmt.Sprintf("missing annotation %s", key))
		}
		if anno != "foo" {
			return admission.Denied(fmt.Sprintf("annotation %s did not have value %q", key, "foo"))
		}
	*/
	var spec selector.SelectSpec
	needAuth := true
	username := req.UserInfo.Username
	groups := req.UserInfo.Groups
	chaosKind := req.Kind.Kind

	switch chaosKind {
	case v1alpha1.KindPodChaos:
		chaos := &v1alpha1.PodChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindIoChaos:
		chaos := &v1alpha1.IoChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindNetworkChaos:
		chaos := &v1alpha1.NetworkChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindTimeChaos:
		chaos := &v1alpha1.TimeChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindKernelChaos:
		chaos := &v1alpha1.KernelChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindStressChaos:
		chaos := &v1alpha1.StressChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindDNSChaos:
		chaos := &v1alpha1.DNSChaos{}
		err := v.decoder.Decode(req, chaos)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		spec = &chaos.Spec

	case v1alpha1.KindAwsChaos:
		needAuth = false
	default:
		err := fmt.Errorf("kind %s is not support", chaosKind)

		return admission.Errored(http.StatusBadRequest, err)
	}

	if !needAuth {
		return admission.Allowed("")
	}

	pods, err := selector.SelectAndFilterPods(context.Background(), v.Client, v.Reader, spec, v.ClusterScoped, v.TargetNamespace, v.AllowedNamespaces, v.IgnoredNamespaces)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	log2.Info("select pods in webhook", "pods", pods)

	namespaceMap := make(map[string]struct{})
	for _, pod := range pods {
		namespaceMap[pod.Namespace] = struct{}{}
	}
	for namespace := range namespaceMap {
		allow, err := v.auth(username, groups, namespace)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		if !allow {
			return admission.Denied(fmt.Sprintf("%s don't have privileges on namespace %s", username, namespace))
		}
	}

	return admission.Allowed("")
}

// PodValidator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *AuthValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

func (v *AuthValidator) auth(username string, groups []string, namespace string) (bool, error) {

	config := ctrl.GetConfigOrDie()
	authCli, err := authorizationv1.NewForConfig(config)

	sar := authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "create",

				Group:    "chaos-mesh.org",
				Resource: "*",
			},
			User: username,
			Groups: groups,
		},
	}

	response, err := authCli.SubjectAccessReviews().Create(&sar)
	if err != nil {
		return false, err
	}

	return response.Status.Allowed, nil
}
