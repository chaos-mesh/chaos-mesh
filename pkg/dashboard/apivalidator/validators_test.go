// Copyright 2026 Chaos Mesh Authors.
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

package apivalidator

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScalarValidators(t *testing.T) {
	cases := []struct {
		name     string
		validate validator.Func
		valid    []string
		invalid  []string
	}{
		{"name", NameValid, []string{"pod-chaos", "a", strings.Repeat("a", 63)}, []string{"", strings.Repeat("a", 64), "has space", "a/b"}},
		{"cron", CronValid, []string{"", "*/5 * * * *", "@hourly"}, []string{"bad cron", "60 * * * *", "* * * * * *"}},
		{"duration", DurationValid, []string{"", "0", "10s", "1h30m", "-1s"}, []string{"tomorrow", "1", "1d"}},
		{"value", ValueValid, []string{"", "0", "0.5", "100", "1e3"}, []string{"-1", "bad", "NaN", "+Inf", "-Inf", "Infinity"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := validator.New()
			if err := v.RegisterValidation("test", tc.validate); err != nil {
				t.Fatal(err)
			}
			for _, value := range tc.valid {
				if err := v.Var(value, "test"); err != nil {
					t.Errorf("rejected valid input %q: %v", value, err)
				}
			}
			for _, value := range tc.invalid {
				if err := v.Var(value, "test"); err == nil {
					t.Errorf("accepted invalid input %q", value)
				}
			}
		})
	}
}

func TestSelectorValidators(t *testing.T) {
	cases := []struct {
		name     string
		validate validator.Func
		valid    []interface{}
		invalid  []interface{}
	}{
		{"namespace", NamespaceSelectorsValid, []interface{}{[]string(nil), []string{}, []string{"default", "chaos-mesh"}}, []interface{}{[]string{""}, []string{"invalid ns"}, []string{strings.Repeat("a", 64)}}},
		{"map", MapSelectorsValid, []interface{}{map[string]string(nil), map[string]string{}, map[string]string{"app.kubernetes.io/name": "chaos"}}, []interface{}{map[string]string{"invalid key": "value"}, map[string]string{"/bad": "value"}}},
		{"phase", PhaseSelectorsValid, []interface{}{[]string(nil), []string{"Running", "Pending", "Succeeded", "Failed", "Unknown"}}, []interface{}{[]string{"running"}, []string{""}, []string{"Running", "invalid"}}},
		{"pods", PodsValid, []interface{}{map[string][]string(nil), map[string][]string{"default": {"pod-0", "pod-1"}}}, []interface{}{map[string][]string{"invalid ns": {"pod"}}, map[string][]string{"default": {"invalid pod"}}}},
		{"physical machines", PhysicalMachineValid, []interface{}{map[string][]string(nil), map[string][]string{"default": {"machine-0"}}}, []interface{}{map[string][]string{"invalid ns": {"machine"}}, map[string][]string{"default": {"invalid machine"}}}},
		{"requirements", RequirementSelectorsValid,
			[]interface{}{[]metav1.LabelSelectorRequirement(nil), []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"a"}}, {Key: "other", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"b"}}, {Key: "exists", Operator: metav1.LabelSelectorOpExists}, {Key: "absent", Operator: metav1.LabelSelectorOpDoesNotExist}}},
			[]interface{}{[]metav1.LabelSelectorRequirement{{Key: "bad key", Operator: metav1.LabelSelectorOpExists}}, []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpIn}}, []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpNotIn}}, []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpExists, Values: []string{"a"}}}, []metav1.LabelSelectorRequirement{{Key: "app", Operator: metav1.LabelSelectorOpDoesNotExist, Values: []string{"a"}}}, []metav1.LabelSelectorRequirement{{Key: "app", Operator: "invalid"}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := validator.New()
			if err := v.RegisterValidation("test", tc.validate); err != nil {
				t.Fatal(err)
			}
			for _, value := range tc.valid {
				if err := v.Var(value, "test"); err != nil {
					t.Errorf("rejected valid input %#v: %v", value, err)
				}
			}
			for _, value := range tc.invalid {
				if err := v.Var(value, "test"); err == nil {
					t.Errorf("accepted invalid input %#v", value)
				}
			}
		})
	}
}

func TestRequiredFieldEqual(t *testing.T) {
	type request struct {
		Mode  string
		Value string `validate:"RequiredFieldEqual=Mode:fixed"`
	}
	v := validator.New()
	if err := v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		input request
		valid bool
	}{
		{request{Mode: "all"}, true}, {request{Mode: "fixed", Value: "1"}, true}, {request{Mode: "fixed"}, false},
	} {
		if err := v.Struct(tc.input); (err == nil) != tc.valid {
			t.Errorf("input %#v: error %v, want valid %v", tc.input, err, tc.valid)
		}
		if err := v.Struct(&tc.input); (err == nil) != tc.valid {
			t.Errorf("pointer input %#v: error %v, want valid %v", tc.input, err, tc.valid)
		}
	}
}
