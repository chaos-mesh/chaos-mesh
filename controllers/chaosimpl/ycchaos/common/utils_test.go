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

package common

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestParseYCChaosAndSelector(t *testing.T) {
	g := NewWithT(t)
	log := zap.New(zap.UseDevMode(true))

	testCases := []struct {
		name        string
		obj         v1alpha1.InnerObject
		records     []*v1alpha1.Record
		index       int
		expectError bool
		expectedID  string
	}{
		{
			name: "valid YCChaos and selector",
			obj: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ycchaos",
					Namespace: "default",
				},
				Spec: v1alpha1.YCChaosSpec{
					Action: v1alpha1.ComputeStop,
					YCSelector: v1alpha1.YCSelector{
						ComputeInstance: "test-instance-id",
					},
				},
			},
			records: []*v1alpha1.Record{
				{
					Id: `{"computeInstance":"test-instance-id"}`,
				},
			},
			index:       0,
			expectError: false,
			expectedID:  "test-instance-id",
		},
		{
			name: "invalid object type",
			obj: &v1alpha1.PodChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-podchaos",
				},
			},
			records: []*v1alpha1.Record{
				{
					Id: `{"computeInstance":"test-instance-id"}`,
				},
			},
			index:       0,
			expectError: true,
		},
		{
			name: "invalid JSON in record",
			obj: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ycchaos",
				},
			},
			records: []*v1alpha1.Record{
				{
					Id: `invalid-json`,
				},
			},
			index:       0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ycchaos, selector, err := ParseYCChaosAndSelector(tc.obj, tc.records, tc.index, log)

			if tc.expectError {
				g.Expect(err).Should(HaveOccurred())
				g.Expect(ycchaos).Should(BeNil())
				g.Expect(selector).Should(BeNil())
			} else {
				g.Expect(err).ShouldNot(HaveOccurred())
				g.Expect(ycchaos).ShouldNot(BeNil())
				g.Expect(selector).ShouldNot(BeNil())
				g.Expect(selector.ComputeInstance).Should(Equal(tc.expectedID))
			}
		})
	}
}

func TestGetYandexCloudSDK(t *testing.T) {
	g := NewWithT(t)

	// Create a fake service account key
	fakeKey := map[string]interface{}{
		"id":                "test-key-id",
		"service_account_id": "test-sa-id",
		"private_key":       "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB\n-----END PRIVATE KEY-----",
		"public_key":        "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu1SU1L7VLPHCgQC7VJT\n-----END PUBLIC KEY-----",
	}
	keyBytes, _ := json.Marshal(fakeKey)

	testCases := []struct {
		name        string
		ycchaos     *v1alpha1.YCChaos
		secret      *v1.Secret
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing secret name",
			ycchaos: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ycchaos",
					Namespace: "default",
				},
				Spec: v1alpha1.YCChaosSpec{
					SecretName: nil,
				},
			},
			expectError: true,
			errorMsg:    "secret name is required",
		},
		{
			name: "secret not found",
			ycchaos: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ycchaos",
					Namespace: "default",
				},
				Spec: v1alpha1.YCChaosSpec{
					SecretName: stringPtr("nonexistent-secret"),
				},
			},
			expectError: true,
			errorMsg:    "fail to get cloud secret",
		},
		{
			name: "missing sa-key.json in secret",
			ycchaos: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ycchaos",
					Namespace: "default",
				},
				Spec: v1alpha1.YCChaosSpec{
					SecretName: stringPtr("test-secret"),
				},
			},
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"other-key": []byte("some-data"),
				},
			},
			expectError: true,
			errorMsg:    "sa-key.json not found in secret",
		},
		{
			name: "invalid service account key format",
			ycchaos: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ycchaos",
					Namespace: "default",
				},
				Spec: v1alpha1.YCChaosSpec{
					SecretName: stringPtr("test-secret"),
				},
			},
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"sa-key.json": []byte("invalid-json"),
				},
			},
			expectError: true,
			errorMsg:    "fail to parse service account key",
		},
		{
			name: "valid secret with service account key",
			ycchaos: &v1alpha1.YCChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ycchaos",
					Namespace: "default",
				},
				Spec: v1alpha1.YCChaosSpec{
					SecretName: stringPtr("test-secret"),
				},
			},
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"sa-key.json": keyBytes,
				},
			},
			expectError: true, // Will fail at SDK build stage due to fake key
			errorMsg:    "fail to build Yandex Cloud SDK",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = v1.AddToScheme(scheme)
			_ = v1alpha1.AddToScheme(scheme)

			var objects []runtime.Object
			if tc.secret != nil {
				objects = append(objects, tc.secret)
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			sdk, err := GetYandexCloudSDK(context.Background(), client, tc.ycchaos)

			if tc.expectError {
				g.Expect(err).Should(HaveOccurred())
				g.Expect(err.Error()).Should(ContainSubstring(tc.errorMsg))
				g.Expect(sdk).Should(BeNil())
			} else {
				g.Expect(err).ShouldNot(HaveOccurred())
				g.Expect(sdk).ShouldNot(BeNil())
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
