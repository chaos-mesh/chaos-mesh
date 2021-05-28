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

package sidecar

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/e2econst"
)

func createTemplateConfig(
	ctx context.Context,
	cli client.Client,
	name string,
	data map[string]string,
) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e2econst.ChaosMeshNamespace,
			Name:      name,
			Labels: map[string]string{
				"app.kubernetes.io/component": "template",
			},
		},
		Data: data,
	}
	return cli.Create(ctx, cm)
}

func createInjectionConfig(
	ctx context.Context,
	cli client.Client,
	ns, name string,
	data map[string]string,
) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels: map[string]string{
				"app.kubernetes.io/component": "webhook",
			},
		},
		Data: data,
	}
	return cli.Create(ctx, cm)
}

// enableWebhook enables webhook on the specific namespace
func enableWebhook(ns string) error {
	args := []string{"label", "ns", ns, "--overwrite", "admission-webhook=enabled"}
	out, err := exec.New().Command("kubectl", args...).CombinedOutput()
	if err != nil {
		klog.Fatalf("Failed to run 'kubectl %s'\nCombined output: %q\nError: %v", strings.Join(args, " "), string(out), err)
	}
	return nil
}
