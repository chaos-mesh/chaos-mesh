// Copyright 2021 Chaos Mesh Authors.
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

package v1alpha1

import (
	"encoding/json"
	// "reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// var _ = reflect.String

const KindAWSChaos = "AWSChaos"

// IsDeleted returns whether this resource has been deleted
func (in *AWSChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *AWSChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *AWSChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *AWSChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetStatus returns the status
func (in *AWSChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *AWSChaos) GetSpecAndMetaString() (string, error) {
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

// AWSChaosList contains a list of AWSChaos
type AWSChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *AWSChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *AWSChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *AWSChaos) IsOneShot() bool {

	if in.Spec.Action == Ec2Restart {
		return true
	}

	return false

}

const KindDNSChaos = "DNSChaos"

// IsDeleted returns whether this resource has been deleted
func (in *DNSChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *DNSChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *DNSChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *DNSChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *DNSChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *DNSChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *DNSChaos) GetSpecAndMetaString() (string, error) {
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

// DNSChaosList contains a list of DNSChaos
type DNSChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *DNSChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *DNSChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *DNSChaos) IsOneShot() bool {

	return false

}

const KindGCPChaos = "GCPChaos"

// IsDeleted returns whether this resource has been deleted
func (in *GCPChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *GCPChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *GCPChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *GCPChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *GCPChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *GCPChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *GCPChaos) GetSpecAndMetaString() (string, error) {
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

// GCPChaosList contains a list of GCPChaos
type GCPChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *GCPChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *GCPChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *GCPChaos) IsOneShot() bool {

	if in.Spec.Action == NodeReset {
		return true
	}

	return false

}

const KindHTTPChaos = "HTTPChaos"

// IsDeleted returns whether this resource has been deleted
func (in *HTTPChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *HTTPChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *HTTPChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *HTTPChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *HTTPChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *HTTPChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *HTTPChaos) GetSpecAndMetaString() (string, error) {
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

// HTTPChaosList contains a list of HTTPChaos
type HTTPChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *HTTPChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *HTTPChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *HTTPChaos) IsOneShot() bool {

	return false

}

const KindIOChaos = "IOChaos"

// IsDeleted returns whether this resource has been deleted
func (in *IOChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *IOChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *IOChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *IOChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *IOChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *IOChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *IOChaos) GetSpecAndMetaString() (string, error) {
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

// IOChaosList contains a list of IOChaos
type IOChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IOChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *IOChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *IOChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *IOChaos) IsOneShot() bool {

	return false

}

const KindJVMChaos = "JVMChaos"

// IsDeleted returns whether this resource has been deleted
func (in *JVMChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *JVMChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *JVMChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *JVMChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *JVMChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *JVMChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *JVMChaos) GetSpecAndMetaString() (string, error) {
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

// JVMChaosList contains a list of JVMChaos
type JVMChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JVMChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *JVMChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *JVMChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *JVMChaos) IsOneShot() bool {

	return false

}

const KindKernelChaos = "KernelChaos"

// IsDeleted returns whether this resource has been deleted
func (in *KernelChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *KernelChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *KernelChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *KernelChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *KernelChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *KernelChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *KernelChaos) GetSpecAndMetaString() (string, error) {
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

// KernelChaosList contains a list of KernelChaos
type KernelChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KernelChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *KernelChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *KernelChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *KernelChaos) IsOneShot() bool {

	return false

}

const KindNetworkChaos = "NetworkChaos"

// IsDeleted returns whether this resource has been deleted
func (in *NetworkChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *NetworkChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *NetworkChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *NetworkChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *NetworkChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *NetworkChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *NetworkChaos) GetSpecAndMetaString() (string, error) {
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

// NetworkChaosList contains a list of NetworkChaos
type NetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *NetworkChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *NetworkChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *NetworkChaos) IsOneShot() bool {

	return false

}

const KindPodChaos = "PodChaos"

// IsDeleted returns whether this resource has been deleted
func (in *PodChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *PodChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *PodChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *PodChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *PodChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *PodChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *PodChaos) GetSpecAndMetaString() (string, error) {
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

// PodChaosList contains a list of PodChaos
type PodChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *PodChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *PodChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *PodChaos) IsOneShot() bool {

	if in.Spec.Action == PodKillAction || in.Spec.Action == ContainerKillAction {
		return true
	}

	return false

}

const KindStressChaos = "StressChaos"

// IsDeleted returns whether this resource has been deleted
func (in *StressChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *StressChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *StressChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *StressChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *StressChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *StressChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *StressChaos) GetSpecAndMetaString() (string, error) {
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

// StressChaosList contains a list of StressChaos
type StressChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StressChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *StressChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *StressChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *StressChaos) IsOneShot() bool {

	return false

}

const KindTimeChaos = "TimeChaos"

// IsDeleted returns whether this resource has been deleted
func (in *TimeChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *TimeChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetObjectMeta would return the ObjectMeta for chaos
func (in *TimeChaos) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

// GetDuration would return the duration for chaos
func (in *TimeChaosSpec) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetChaos would return the a record for chaos
// func (in *TimeChaos) GetChaos() GenericChaos{
// 	instance := &ChaosInstance{
// 		Name:      in.Name,
// 		Namespace: in.Namespace,
// 		StartTime: in.CreationTimestamp.Time,
// 		Action:    "",
// 		UID:       string(in.UID),
// 		Status:    in.Status.ChaosStatus,
// 	}

// 	action := reflect.ValueOf(in).Elem().FieldByName("Spec").FieldByName("Action")
// 	if action.IsValid() {
// 		instance.Action = action.String()
// 	}
// 	if in.Spec.Duration != nil {
// 		instance.Duration = *in.Spec.Duration
// 	}
// 	if in.DeletionTimestamp != nil {
// 		instance.EndTime = in.DeletionTimestamp.Time
// 	}
// 	return instance
// }

// GetStatus returns the status
func (in *TimeChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// GetSpecAndMetaString returns a string including the meta and spec field of this chaos object.
func (in *TimeChaos) GetSpecAndMetaString() (string, error) {
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

// TimeChaosList contains a list of TimeChaos
type TimeChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimeChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *TimeChaosList) ListChaos() []GenericChaos {
	var result []GenericChaos
	for _, item := range in.Items {
		item := item
		result = append(result, &item)
	}
	return result
}

func (in *TimeChaos) DurationExceeded(now time.Time) (bool, time.Duration, error) {
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

func (in *TimeChaos) IsOneShot() bool {

	return false

}

func init() {

	SchemeBuilder.Register(&AWSChaos{}, &AWSChaosList{})
	all.register(KindAWSChaos, &ChaosKind{
		Chaos:            &AWSChaos{},
		GenericChaosList: &AWSChaosList{},
	})

	SchemeBuilder.Register(&DNSChaos{}, &DNSChaosList{})
	all.register(KindDNSChaos, &ChaosKind{
		Chaos:            &DNSChaos{},
		GenericChaosList: &DNSChaosList{},
	})

	SchemeBuilder.Register(&GCPChaos{}, &GCPChaosList{})
	all.register(KindGCPChaos, &ChaosKind{
		Chaos:            &GCPChaos{},
		GenericChaosList: &GCPChaosList{},
	})

	SchemeBuilder.Register(&HTTPChaos{}, &HTTPChaosList{})
	all.register(KindHTTPChaos, &ChaosKind{
		Chaos:            &HTTPChaos{},
		GenericChaosList: &HTTPChaosList{},
	})

	SchemeBuilder.Register(&IOChaos{}, &IOChaosList{})
	all.register(KindIOChaos, &ChaosKind{
		Chaos:            &IOChaos{},
		GenericChaosList: &IOChaosList{},
	})

	SchemeBuilder.Register(&JVMChaos{}, &JVMChaosList{})
	all.register(KindJVMChaos, &ChaosKind{
		Chaos:            &JVMChaos{},
		GenericChaosList: &JVMChaosList{},
	})

	SchemeBuilder.Register(&KernelChaos{}, &KernelChaosList{})
	all.register(KindKernelChaos, &ChaosKind{
		Chaos:            &KernelChaos{},
		GenericChaosList: &KernelChaosList{},
	})

	SchemeBuilder.Register(&NetworkChaos{}, &NetworkChaosList{})
	all.register(KindNetworkChaos, &ChaosKind{
		Chaos:            &NetworkChaos{},
		GenericChaosList: &NetworkChaosList{},
	})

	SchemeBuilder.Register(&PodChaos{}, &PodChaosList{})
	all.register(KindPodChaos, &ChaosKind{
		Chaos:            &PodChaos{},
		GenericChaosList: &PodChaosList{},
	})

	SchemeBuilder.Register(&StressChaos{}, &StressChaosList{})
	all.register(KindStressChaos, &ChaosKind{
		Chaos:            &StressChaos{},
		GenericChaosList: &StressChaosList{},
	})

	SchemeBuilder.Register(&TimeChaos{}, &TimeChaosList{})
	all.register(KindTimeChaos, &ChaosKind{
		Chaos:            &TimeChaos{},
		GenericChaosList: &TimeChaosList{},
	})

	allScheduleItem.register(KindAWSChaos, &ChaosKind{
		Chaos:            &AWSChaos{},
		GenericChaosList: &AWSChaosList{},
	})

	allScheduleItem.register(KindDNSChaos, &ChaosKind{
		Chaos:            &DNSChaos{},
		GenericChaosList: &DNSChaosList{},
	})

	allScheduleItem.register(KindGCPChaos, &ChaosKind{
		Chaos:            &GCPChaos{},
		GenericChaosList: &GCPChaosList{},
	})

	allScheduleItem.register(KindHTTPChaos, &ChaosKind{
		Chaos:            &HTTPChaos{},
		GenericChaosList: &HTTPChaosList{},
	})

	allScheduleItem.register(KindIOChaos, &ChaosKind{
		Chaos:            &IOChaos{},
		GenericChaosList: &IOChaosList{},
	})

	allScheduleItem.register(KindJVMChaos, &ChaosKind{
		Chaos:            &JVMChaos{},
		GenericChaosList: &JVMChaosList{},
	})

	allScheduleItem.register(KindKernelChaos, &ChaosKind{
		Chaos:            &KernelChaos{},
		GenericChaosList: &KernelChaosList{},
	})

	allScheduleItem.register(KindNetworkChaos, &ChaosKind{
		Chaos:            &NetworkChaos{},
		GenericChaosList: &NetworkChaosList{},
	})

	allScheduleItem.register(KindPodChaos, &ChaosKind{
		Chaos:            &PodChaos{},
		GenericChaosList: &PodChaosList{},
	})

	allScheduleItem.register(KindStressChaos, &ChaosKind{
		Chaos:            &StressChaos{},
		GenericChaosList: &StressChaosList{},
	})

	allScheduleItem.register(KindTimeChaos, &ChaosKind{
		Chaos:            &TimeChaos{},
		GenericChaosList: &TimeChaosList{},
	})

	allScheduleItem.register(KindWorkflow, &ChaosKind{
		Chaos:            &Workflow{},
		GenericChaosList: &WorkflowList{},
	})

}
