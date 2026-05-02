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
//

package apivalidator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newValidate(t *testing.T) *validator.Validate {
	t.Helper()
	v := validator.New()
	v.RegisterValidation("NameValid", NameValid)
	v.RegisterValidation("CronValid", CronValid)
	v.RegisterValidation("DurationValid", DurationValid)
	v.RegisterValidation("NamespaceSelectorsValid", NamespaceSelectorsValid)
	v.RegisterValidation("MapSelectorsValid", MapSelectorsValid)
	v.RegisterValidation("RequirementSelectorsValid", RequirementSelectorsValid)
	v.RegisterValidation("PhaseSelectorsValid", PhaseSelectorsValid)
	v.RegisterValidation("ValueValid", ValueValid)
	v.RegisterValidation("PodsValid", PodsValid)
	v.RegisterValidation("PhysicalMachineValid", PhysicalMachineValid)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)
	return v
}

type nameStruct struct {
	Name string `validate:"NameValid"`
}

func TestNameValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid simple name", "my-chaos", true},
		{"valid with dot", "chaos.test", true},
		{"valid underscore", "chaos_test", true},
		{"empty name", "", false},
		{"too long (64 chars)", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", false},
		{"invalid chars", "chaos@test!", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(nameStruct{Name: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}


type cronStruct struct {
	Cron string `validate:"CronValid"`
}

func TestCronValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{"empty cron (allowed)", "", true},
		{"valid cron", "0 * * * *", true},
		{"valid every minute", "* * * * *", true},
		{"invalid cron", "invalid-cron", false},
		{"too few fields", "* * *", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(cronStruct{Cron: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- DurationValid ----------

type durationStruct struct {
	Duration string `validate:"DurationValid"`
}

func TestDurationValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{"empty duration (allowed)", "", true},
		{"valid seconds", "30s", true},
		{"valid minutes", "5m", true},
		{"valid hours", "1h", true},
		{"invalid duration", "notaduration", false},
		{"invalid number only", "100", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(durationStruct{Duration: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

type namespaceSelectorStruct struct {
	Namespaces []string `validate:"NamespaceSelectorsValid"`
}

func TestNamespaceSelectorsValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input []string
		valid bool
	}{
		{"empty list", []string{}, true},
		{"valid namespaces", []string{"default", "kube-system"}, true},
		{"empty string in list", []string{"default", ""}, false},
		{"name too long", []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, false},
		{"invalid chars", []string{"default@chaos"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(namespaceSelectorStruct{Namespaces: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- MapSelectorsValid ----------

type mapSelectorStruct struct {
	Labels map[string]string `validate:"MapSelectorsValid"`
}

func TestMapSelectorsValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input map[string]string
		valid bool
	}{
		{"nil map (allowed)", nil, true},
		{"valid labels", map[string]string{"app": "chaos", "env": "test"}, true},
		{"valid prefixed label", map[string]string{"chaos-mesh.org/type": "pod"}, true},
		{"invalid key", map[string]string{"invalid key!": "value"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(mapSelectorStruct{Labels: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}


type requirementSelectorStruct struct {
	Requirements []metav1.LabelSelectorRequirement `validate:"RequirementSelectorsValid"`
}

func TestRequirementSelectorsValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input []metav1.LabelSelectorRequirement
		valid bool
	}{
		{"nil slice (allowed)", nil, true},
		{"valid In operator", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{"chaos"}},
		}, true},
		{"valid NotIn operator", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: metav1.LabelSelectorOpNotIn, Values: []string{"chaos"}},
		}, true},
		{"valid Exists operator", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: metav1.LabelSelectorOpExists},
		}, true},
		{"valid DoesNotExist operator", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: metav1.LabelSelectorOpDoesNotExist},
		}, true},
		{"In with no values", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: metav1.LabelSelectorOpIn, Values: []string{}},
		}, false},
		{"Exists with values", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: metav1.LabelSelectorOpExists, Values: []string{"chaos"}},
		}, false},
		{"invalid key", []metav1.LabelSelectorRequirement{
			{Key: "invalid key!", Operator: metav1.LabelSelectorOpExists},
		}, false},
		{"unsupported operator", []metav1.LabelSelectorRequirement{
			{Key: "app", Operator: "Unknown"},
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(requirementSelectorStruct{Requirements: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- PhaseSelectorsValid ----------

type phaseSelectorStruct struct {
	Phases []string `validate:"PhaseSelectorsValid"`
}

func TestPhaseSelectorsValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input []string
		valid bool
	}{
		{"empty list", []string{}, true},
		{"valid Running", []string{"Running"}, true},
		{"valid Failed", []string{"Failed"}, true},
		{"valid Pending", []string{"Pending"}, true},
		{"valid Succeeded", []string{"Succeeded"}, true},
		{"valid Unknown", []string{"Unknown"}, true},
		{"multiple valid", []string{"Running", "Failed"}, true},
		{"invalid phase", []string{"InvalidPhase"}, false},
		{"mixed valid and invalid", []string{"Running", "BadPhase"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(phaseSelectorStruct{Phases: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- ValueValid ----------

type valueStruct struct {
	Value string `validate:"ValueValid"`
}

func TestValueValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input string
		valid bool
	}{
		{"empty value (allowed)", "", true},
		{"valid integer", "50", true},
		{"valid float", "0.5", true},
		{"valid zero", "0", true},
		{"invalid string", "abc", false},
		{"negative value", "-1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(valueStruct{Value: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- PodsValid ----------

type podsStruct struct {
	Pods map[string][]string `validate:"PodsValid"`
}

func TestPodsValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input map[string][]string
		valid bool
	}{
		{"nil map (allowed)", nil, true},
		{"valid pods", map[string][]string{"default": {"pod-1", "pod-2"}}, true},
		{"invalid namespace", map[string][]string{"invalid@ns": {"pod-1"}}, false},
		{"invalid pod name", map[string][]string{"default": {"pod@invalid"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(podsStruct{Pods: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- PhysicalMachineValid ----------

type physicalMachineStruct struct {
	Machines map[string][]string `validate:"PhysicalMachineValid"`
}

func TestPhysicalMachineValid(t *testing.T) {
	v := newValidate(t)
	cases := []struct {
		name  string
		input map[string][]string
		valid bool
	}{
		{"nil map (allowed)", nil, true},
		{"valid machines", map[string][]string{"default": {"machine-1", "machine-2"}}, true},
		{"invalid namespace", map[string][]string{"invalid@ns": {"machine-1"}}, false},
		{"invalid machine name", map[string][]string{"default": {"machine@bad"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(physicalMachineStruct{Machines: tc.input})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- RequiredFieldEqualValid (string param field) ----------

type requiredFieldStruct struct {
	Action string `validate:""`
	Target string `validate:"RequiredFieldEqual=Action:NetworkChaos"`
}

func TestRequiredFieldEqualValid(t *testing.T) {
	v := newValidate(t)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)

	cases := []struct {
		name   string
		action string
		target string
		valid  bool
	}{
		{"action matches, target present", "NetworkChaos", "some-target", true},
		{"action matches, target empty", "NetworkChaos", "", false},
		{"action does not match, target empty (allowed)", "PodChaos", "", true},
		{"action does not match, target present", "PodChaos", "some-target", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(requiredFieldStruct{Action: tc.action, Target: tc.target})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- RequiredFieldEqualValid (int param field) — covers isEq int branch + asInt ----------

type intRequiredFieldStruct struct {
	Mode   int    `validate:""`
	Target string `validate:"RequiredFieldEqual=Mode:1"`
}

func TestRequiredFieldEqualValid_IntField(t *testing.T) {
	v := newValidate(t)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)

	cases := []struct {
		name   string
		mode   int
		target string
		valid  bool
	}{
		{"mode matches, target present", 1, "some-target", true},
		{"mode matches, target empty", 1, "", false},
		{"mode does not match, target empty", 0, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(intRequiredFieldStruct{Mode: tc.mode, Target: tc.target})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- RequiredFieldEqualValid (uint param field) — covers isEq uint branch + asUint ----------

type uintRequiredFieldStruct struct {
	Mode   uint   `validate:""`
	Target string `validate:"RequiredFieldEqual=Mode:2"`
}

func TestRequiredFieldEqualValid_UintField(t *testing.T) {
	v := newValidate(t)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)

	cases := []struct {
		name   string
		mode   uint
		target string
		valid  bool
	}{
		{"mode matches, target present", 2, "some-target", true},
		{"mode matches, target empty", 2, "", false},
		{"mode does not match, target empty", 0, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(uintRequiredFieldStruct{Mode: tc.mode, Target: tc.target})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- RequiredFieldEqualValid (float param field) — covers isEq float branch + asFloat ----------

type floatRequiredFieldStruct struct {
	Score  float64 `validate:""`
	Target string  `validate:"RequiredFieldEqual=Score:1.5"`
}

func TestRequiredFieldEqualValid_FloatField(t *testing.T) {
	v := newValidate(t)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)

	cases := []struct {
		name   string
		score  float64
		target string
		valid  bool
	}{
		{"score matches, target present", 1.5, "some-target", true},
		{"score matches, target empty", 1.5, "", false},
		{"score does not match, target empty", 0.0, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(floatRequiredFieldStruct{Score: tc.score, Target: tc.target})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- RequiredFieldEqualValid (pointer validated field) — covers requireCheckFieldKind Ptr branch ----------

type ptrRequiredFieldStruct struct {
	Action string  `validate:""`
	Target *string `validate:"RequiredFieldEqual=Action:chaos"`
}

func TestRequiredFieldEqualValid_PtrField(t *testing.T) {
	v := newValidate(t)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)

	nonNilTarget := "some-target"
	cases := []struct {
		name   string
		action string
		target *string
		valid  bool
	}{
		{"action matches, target ptr non-nil", "chaos", &nonNilTarget, true},
		{"action matches, target ptr nil", "chaos", nil, false},
		{"action does not match, target nil (allowed)", "other", nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(ptrRequiredFieldStruct{Action: tc.action, Target: tc.target})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

// ---------- RequiredFieldEqualValid (slice param field) — covers isEq slice/len branch ----------

type sliceRequiredFieldStruct struct {
	Tags   []string `validate:""`
	Target string   `validate:"RequiredFieldEqual=Tags:2"`
}

func TestRequiredFieldEqualValid_SliceField(t *testing.T) {
	v := newValidate(t)
	v.RegisterValidation("RequiredFieldEqual", RequiredFieldEqualValid, true)

	cases := []struct {
		name   string
		tags   []string
		target string
		valid  bool
	}{
		{"slice len matches, target present", []string{"a", "b"}, "some-target", true},
		{"slice len matches, target empty", []string{"a", "b"}, "", false},
		{"slice len does not match, target empty", []string{"a"}, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(sliceRequiredFieldStruct{Tags: tc.tags, Target: tc.target})
			if tc.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}
