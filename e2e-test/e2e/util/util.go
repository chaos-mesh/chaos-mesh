// Copyright 2020 Chaos Mesh Authors.
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

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	aggregatorclientset "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// WaitForAPIServicesAvailable waits for apiservices to be available
func WaitForAPIServicesAvailable(client aggregatorclientset.Interface, selector labels.Selector) error {
	isAvailable := func(status apiregistrationv1.APIServiceStatus) bool {
		if status.Conditions == nil {
			return false
		}
		for _, condition := range status.Conditions {
			if condition.Type == apiregistrationv1.Available {
				return condition.Status == apiregistrationv1.ConditionTrue
			}
		}
		return false
	}
	return wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
		apiServiceList, err := client.ApiregistrationV1().APIServices().List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err != nil {
			return false, err
		}
		for _, apiService := range apiServiceList.Items {
			if !isAvailable(apiService.Status) {
				framework.Logf("APIService %q is not available yet", apiService.Name)
				return false, nil
			}
		}
		for _, apiService := range apiServiceList.Items {
			framework.Logf("APIService %q is available", apiService.Name)
		}
		return true, nil
	})
}

// WaitForCRDsEstablished waits for all CRDs to be established
func WaitForCRDsEstablished(client apiextensionsclientset.Interface, selector labels.Selector) error {
	isEstablished := func(status apiextensionsv1beta1.CustomResourceDefinitionStatus) bool {
		if status.Conditions == nil {
			return false
		}
		for _, condition := range status.Conditions {
			if condition.Type == apiextensionsv1beta1.Established {
				return condition.Status == apiextensionsv1beta1.ConditionTrue
			}
		}
		return false
	}
	return wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
		crdList, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err != nil {
			return false, err
		}
		for _, crd := range crdList.Items {
			if !isEstablished(crd.Status) {
				framework.Logf("CRD %q is not established yet", crd.Name)
				return false, nil
			}
		}
		for _, crd := range crdList.Items {
			framework.Logf("CRD %q is established", crd.Name)
		}
		return true, nil
	})
}

// WaitDeploymentReady waits for all pods which controlled by deployment to be ready.
func WaitDeploymentReady(name, namespace string, cli kubernetes.Interface) error {
	return wait.Poll(2*time.Second, 5*time.Minute, func() (done bool, err error) {
		d, err := cli.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if d.Status.AvailableReplicas != *d.Spec.Replicas {
			return false, nil
		}
		if d.Status.UpdatedReplicas != *d.Spec.Replicas {
			return false, nil
		}
		return true, nil
	})
}

func PauseChaos(ctx context.Context, cli client.Client, chaos runtime.Object) error {
	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{v1alpha1.PauseAnnotationKey: "true"},
		},
	})
	return cli.Patch(ctx, chaos, client.ConstantPatch(types.MergePatchType, mergePatch))
}

func UnPauseChaos(ctx context.Context, cli client.Client, chaos runtime.Object) error {
	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{v1alpha1.PauseAnnotationKey: "false"},
		},
	})
	return cli.Patch(ctx, chaos, client.ConstantPatch(types.MergePatchType, mergePatch))
}

func WaitE2EHelperReady(c http.Client, port uint16) error {
	return wait.Poll(2*time.Second, 5*time.Minute, func() (done bool, err error) {
		if _, err = c.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err != nil {
			return false, nil
		}
		return true, nil
	})
}
