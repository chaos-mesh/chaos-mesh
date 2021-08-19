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

package main

import (
	"bytes"
	"text/template"
)

const implImport = `
import (
	"encoding/json"
	"reflect"
	"time"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"

	gw "github.com/chaos-mesh/chaos-mesh/api/v1alpha1/genericwebhook"
)

// updating spec of a chaos will have no effect, we'd better reject it
var ErrCanNotUpdateChaos = fmt.Errorf("Cannot update chaos spec")
`

const implTemplate = `
const Kind{{.Type}} = "{{.Type}}"

// IsDeleted returns whether this resource has been deleted
func (in *{{.Type}}) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *{{.Type}}) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *{{.Type}}) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *{{.Type}}Spec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetStatus returns the status
func (in *{{.Type}}) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *{{.Type}}) GetSpecAndMetaString() (string, error) {
	spec, err := json.Marshal(in.Spec)
	if err != nil {
		return "", err
	}

	meta := in.ObjectMeta.DeepCopy()
	meta.SetResourceVersion("")
	meta.SetGeneration(0)

	return string(spec) + meta.String(), nil
}

// +kubebuilder:object:root=true

// {{.Type}}List contains a list of {{.Type}}
type {{.Type}}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{.Type}} ` + "`" + `json:"items"` + "`" + `
}

// ListChaos returns a list of chaos
func (in *{{.Type}}List) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *{{.Type}}) DurationExceeded(now time.Time) (bool, time.Duration, error) {
	duration, err := in.Spec.GetDuration()
	if err != nil {
		return false, 0, err
	}

	if duration != nil {
		stopTime := in.GetCreationTimestamp().Add(*duration)
		if stopTime.Before(now) {
			return true, 0, nil
		}

		return false, stopTime.Sub(now), nil
	}

	return false, 0, nil
}

func (in *{{.Type}}) IsOneShot() bool {
	{{if .OneShotExp}}
	if {{.OneShotExp}} {
		return true
	}

	return false
	{{else}}
	return false
	{{end}}
}

var {{.Type}}WebhookLog = logf.Log.WithName("{{.Type}}-resource")

func (in *{{.Type}}) ValidateCreate() error {
	{{.Type}}WebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *{{.Type}}) ValidateUpdate(old runtime.Object) error {
	{{.Type}}WebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*{{.Type}}).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *{{.Type}}) ValidateDelete() error {
	{{.Type}}WebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &{{.Type}}{}

func (in *{{.Type}}) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &{{.Type}}{}

func (in *{{.Type}}) Default() {
	gw.Default(in)
}
`

func generateImpl(name string, oneShotExp string) string {
	tmpl, err := template.New("impl").Parse(implTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &metadata{
		Type:       name,
		OneShotExp: oneShotExp,
	})
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}
