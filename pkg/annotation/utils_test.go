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

package annotation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPodChaos(name string, action v1alpha1.PodChaosAction) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1alpha1.PodChaosSpec{Action: action},
	}
}

func TestGenKeyForImage_Normal(t *testing.T) {
	pc := newPodChaos("test-chaos", v1alpha1.PodKillAction)
	key := GenKeyForImage(pc, "app", false)
	expected := fmt.Sprintf("%s-%s-%s-%s-image", AnnotationPrefix, "test-chaos", v1alpha1.PodKillAction, "app-normal")
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestGenKeyForImage_Init(t *testing.T) {
	pc := newPodChaos("test-chaos", v1alpha1.PodFailureAction)
	key := GenKeyForImage(pc, "app", true)
	expected := fmt.Sprintf("%s-%s-%s-%s-image", AnnotationPrefix, "test-chaos", v1alpha1.PodFailureAction, "app-init")
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestGenKeyForImage_FallbackWhenTooLong(t *testing.T) {
	// name long enough to push imageKey beyond 63 chars
	longName := "this-is-a-very-long-chaos-experiment-name-that-exceeds-limit"
	pc := newPodChaos(longName, v1alpha1.ContainerKillAction)
	containerName := "mycontainer"
	key := GenKeyForImage(pc, containerName, false)

	// the full key would be >63 chars so it should fall back to containerName+"-normal"
	if len(key) > 63 {
		t.Errorf("key length %d exceeds 63 chars: %q", len(key), key)
	}
	if !strings.HasSuffix(key, "-normal") {
		t.Errorf("expected fallback key to end with -normal, got %q", key)
	}
}

func TestGenKeyForImage_ExactlyAtLimit(t *testing.T) {
	pc := newPodChaos("chaos", v1alpha1.PodKillAction)
	key := GenKeyForImage(pc, "c", false)
	// short enough to not trigger fallback
	expected := fmt.Sprintf("%s-%s-%s-%s-image", AnnotationPrefix, "chaos", v1alpha1.PodKillAction, "c-normal")
	if key != expected {
		t.Errorf("expected %q, got %q", expected, key)
	}
}

func TestGenKeyForWebhook(t *testing.T) {
	cases := []struct {
		name     string
		prefix   string
		podName  string
		expected string
	}{
		{"standard", "chaos-mesh", "my-pod", "chaos-mesh-my-pod"},
		{"empty prefix", "", "my-pod", "-my-pod"},
		{"empty podName", "chaos-mesh", "", "chaos-mesh-"},
		{"both empty", "", "", "-"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GenKeyForWebhook(tc.prefix, tc.podName)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
