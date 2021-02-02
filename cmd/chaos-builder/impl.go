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
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
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

// GetDuration would return the duration for chaos
func (in *{{.Type}}) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetName would return the name for chaos
func (in *{{.Type}}) GetName() string {
	return in.Name
}

// SetName would set the name for chaos
func (in *{{.Type}}) SetName(name string) {
	in.Name = name
}

// GetActiveJob would return the active job of chaos
func (in *{{.Type}}) GetActiveJob() *types.NamespacedName {
	activeJob := in.Status.ActiveJob
	if len(activeJob) == 0 {
		return nil
	}

	parts := strings.Split(activeJob, "/")
	return &types.NamespacedName {parts[0], parts[1]}
}

// SetActiveJob would set the active job of chaos
func (in *{{.Type}}) SetActiveJob(namespacedName *types.NamespacedName)  {
	if namespacedName == nil {
		in.Status.ActiveJob = ""
	} else {
		in.Status.ActiveJob = namespacedName.String()
	}
}

func (in *{{.Type}}) GetJobObject() Job {
	return &{{.Type}} {}
}

func (in *{{.Type}}) IntoJobWithoutName() Job {
	job := in.DeepCopyObject().(*{{.Type}})
	job.Spec.Scheduler = nil
	job.Spec.Duration = nil
	job.ObjectMeta = metav1.ObjectMeta {
		Namespace: job.Namespace,
		Name: "",
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: job.APIVersion,
				Kind: job.Kind,
				Name: job.Name,
				UID: job.UID,
			},
		},
	}

	return job
}

func (in *{{.Type}}) UpdateJob(j Job) bool {
	chaos := j.(*{{.Type}})
	newChaos := in.IntoJobWithoutName().(*{{.Type}})

	if reflect.DeepEqual(newChaos.Spec, chaos.Spec) &&
		reflect.DeepEqual(newChaos.Labels, chaos.Labels) &&
		reflect.DeepEqual(newChaos.Annotations, chaos.Annotations) &&
		reflect.DeepEqual(newChaos.OwnerReferences, chaos.OwnerReferences) {
		return false
	}

	newChaos.Spec.DeepCopyInto(&chaos.Spec)

	if newChaos.Labels != nil {
		in, out := &newChaos.Labels, &chaos.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if newChaos.Annotations != nil {
		in, out := &newChaos.Annotations, &chaos.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if newChaos.OwnerReferences != nil {
		in, out := &newChaos.OwnerReferences, &chaos.OwnerReferences
		*out = make([]metav1.OwnerReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	return true
}

func (in *{{.Type}}) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *{{.Type}}) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *{{.Type}}) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *{{.Type}}) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextRecover = nil
		return
	}

	if in.Status.Scheduler.NextRecover == nil {
		in.Status.Scheduler.NextRecover = &metav1.Time{}
	}
	in.Status.Scheduler.NextRecover.Time = t
}

// GetScheduler would return the scheduler for chaos
func (in *{{.Type}}) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *{{.Type}}) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      Kind{{.Type}},
		StartTime: in.CreationTimestamp.Time,
		Action:    "",
		Status:    string(in.Status.Experiment.Phase),
		UID:       string(in.UID),
	}

	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
	if action.IsValid() {
		instance.Action = action.String()
	}
	if in.Spec.Duration != nil {
		instance.Duration = *in.Spec.Duration
	}
	if in.DeletionTimestamp != nil {
		instance.EndTime = in.DeletionTimestamp.Time
	}
	return instance
}

// GetStatus returns the status
func (in *{{.Type}}) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// {{.Type}}List contains a list of {{.Type}}
type {{.Type}}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{.Type}} ` + "`" + `json:"items"` + "`" + `
}

// ListChaos returns a list of chaos
func (in *{{.Type}}List) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}
`

func generateImpl(name string) string {
	tmpl, err := template.New("impl").Parse(implTemplate)
	if err != nil {
		log.Error(err, "fail to build template")
		return ""
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, &metadata{
		Type: name,
	})
	if err != nil {
		log.Error(err, "fail to execute template")
		return ""
	}

	return buf.String()
}
