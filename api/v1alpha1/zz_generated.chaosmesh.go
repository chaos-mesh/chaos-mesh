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
	duration, err := time.ParseDuration(string(*in.Duration))
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
	
	if in.Spec.Action==Ec2Restart {
		return true
	}

	return false
	
}

var AWSChaosWebhookLog = logf.Log.WithName("AWSChaos-resource")

func (in *AWSChaos) ValidateCreate() error {
	AWSChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *AWSChaos) ValidateUpdate(old runtime.Object) error {
	AWSChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*AWSChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *AWSChaos) ValidateDelete() error {
	AWSChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &AWSChaos{}

func (in *AWSChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &AWSChaos{}

func (in *AWSChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var DNSChaosWebhookLog = logf.Log.WithName("DNSChaos-resource")

func (in *DNSChaos) ValidateCreate() error {
	DNSChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *DNSChaos) ValidateUpdate(old runtime.Object) error {
	DNSChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*DNSChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *DNSChaos) ValidateDelete() error {
	DNSChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &DNSChaos{}

func (in *DNSChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &DNSChaos{}

func (in *DNSChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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
	
	if in.Spec.Action==NodeReset {
		return true
	}

	return false
	
}

var GCPChaosWebhookLog = logf.Log.WithName("GCPChaos-resource")

func (in *GCPChaos) ValidateCreate() error {
	GCPChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *GCPChaos) ValidateUpdate(old runtime.Object) error {
	GCPChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*GCPChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *GCPChaos) ValidateDelete() error {
	GCPChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &GCPChaos{}

func (in *GCPChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &GCPChaos{}

func (in *GCPChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var HTTPChaosWebhookLog = logf.Log.WithName("HTTPChaos-resource")

func (in *HTTPChaos) ValidateCreate() error {
	HTTPChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateUpdate(old runtime.Object) error {
	HTTPChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*HTTPChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateDelete() error {
	HTTPChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &HTTPChaos{}

func (in *HTTPChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &HTTPChaos{}

func (in *HTTPChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var IOChaosWebhookLog = logf.Log.WithName("IOChaos-resource")

func (in *IOChaos) ValidateCreate() error {
	IOChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *IOChaos) ValidateUpdate(old runtime.Object) error {
	IOChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*IOChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *IOChaos) ValidateDelete() error {
	IOChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &IOChaos{}

func (in *IOChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &IOChaos{}

func (in *IOChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var JVMChaosWebhookLog = logf.Log.WithName("JVMChaos-resource")

func (in *JVMChaos) ValidateCreate() error {
	JVMChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *JVMChaos) ValidateUpdate(old runtime.Object) error {
	JVMChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*JVMChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *JVMChaos) ValidateDelete() error {
	JVMChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &JVMChaos{}

func (in *JVMChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &JVMChaos{}

func (in *JVMChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var KernelChaosWebhookLog = logf.Log.WithName("KernelChaos-resource")

func (in *KernelChaos) ValidateCreate() error {
	KernelChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *KernelChaos) ValidateUpdate(old runtime.Object) error {
	KernelChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*KernelChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *KernelChaos) ValidateDelete() error {
	KernelChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &KernelChaos{}

func (in *KernelChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &KernelChaos{}

func (in *KernelChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var NetworkChaosWebhookLog = logf.Log.WithName("NetworkChaos-resource")

func (in *NetworkChaos) ValidateCreate() error {
	NetworkChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *NetworkChaos) ValidateUpdate(old runtime.Object) error {
	NetworkChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*NetworkChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *NetworkChaos) ValidateDelete() error {
	NetworkChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &NetworkChaos{}

func (in *NetworkChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &NetworkChaos{}

func (in *NetworkChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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
	
	if in.Spec.Action==PodKillAction || in.Spec.Action==ContainerKillAction {
		return true
	}

	return false
	
}

var PodChaosWebhookLog = logf.Log.WithName("PodChaos-resource")

func (in *PodChaos) ValidateCreate() error {
	PodChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PodChaos) ValidateUpdate(old runtime.Object) error {
	PodChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*PodChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *PodChaos) ValidateDelete() error {
	PodChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &PodChaos{}

func (in *PodChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &PodChaos{}

func (in *PodChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var StressChaosWebhookLog = logf.Log.WithName("StressChaos-resource")

func (in *StressChaos) ValidateCreate() error {
	StressChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateUpdate(old runtime.Object) error {
	StressChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*StressChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateDelete() error {
	StressChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &StressChaos{}

func (in *StressChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &StressChaos{}

func (in *StressChaos) Default() {
	gw.Default(in)
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
	duration, err := time.ParseDuration(string(*in.Duration))
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

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

var TimeChaosWebhookLog = logf.Log.WithName("TimeChaos-resource")

func (in *TimeChaos) ValidateCreate() error {
	TimeChaosWebhookLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *TimeChaos) ValidateUpdate(old runtime.Object) error {
	TimeChaosWebhookLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*TimeChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *TimeChaos) ValidateDelete() error {
	TimeChaosWebhookLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

var _ webhook.Validator = &TimeChaos{}

func (in *TimeChaos) Validate() error {
	errs := gw.Validate(in)
	return gw.Aggregate(errs)
}

var _ webhook.Defaulter = &TimeChaos{}

func (in *TimeChaos) Default() {
	gw.Default(in)
}

func init() {

	SchemeBuilder.Register(&AWSChaos{}, &AWSChaosList{})
	all.register(KindAWSChaos, &ChaosKind{
		Chaos:     &AWSChaos{},
		GenericChaosList: &AWSChaosList{},
	})

	SchemeBuilder.Register(&DNSChaos{}, &DNSChaosList{})
	all.register(KindDNSChaos, &ChaosKind{
		Chaos:     &DNSChaos{},
		GenericChaosList: &DNSChaosList{},
	})

	SchemeBuilder.Register(&GCPChaos{}, &GCPChaosList{})
	all.register(KindGCPChaos, &ChaosKind{
		Chaos:     &GCPChaos{},
		GenericChaosList: &GCPChaosList{},
	})

	SchemeBuilder.Register(&HTTPChaos{}, &HTTPChaosList{})
	all.register(KindHTTPChaos, &ChaosKind{
		Chaos:     &HTTPChaos{},
		GenericChaosList: &HTTPChaosList{},
	})

	SchemeBuilder.Register(&IOChaos{}, &IOChaosList{})
	all.register(KindIOChaos, &ChaosKind{
		Chaos:     &IOChaos{},
		GenericChaosList: &IOChaosList{},
	})

	SchemeBuilder.Register(&JVMChaos{}, &JVMChaosList{})
	all.register(KindJVMChaos, &ChaosKind{
		Chaos:     &JVMChaos{},
		GenericChaosList: &JVMChaosList{},
	})

	SchemeBuilder.Register(&KernelChaos{}, &KernelChaosList{})
	all.register(KindKernelChaos, &ChaosKind{
		Chaos:     &KernelChaos{},
		GenericChaosList: &KernelChaosList{},
	})

	SchemeBuilder.Register(&NetworkChaos{}, &NetworkChaosList{})
	all.register(KindNetworkChaos, &ChaosKind{
		Chaos:     &NetworkChaos{},
		GenericChaosList: &NetworkChaosList{},
	})

	SchemeBuilder.Register(&PodChaos{}, &PodChaosList{})
	all.register(KindPodChaos, &ChaosKind{
		Chaos:     &PodChaos{},
		GenericChaosList: &PodChaosList{},
	})

	SchemeBuilder.Register(&StressChaos{}, &StressChaosList{})
	all.register(KindStressChaos, &ChaosKind{
		Chaos:     &StressChaos{},
		GenericChaosList: &StressChaosList{},
	})

	SchemeBuilder.Register(&TimeChaos{}, &TimeChaosList{})
	all.register(KindTimeChaos, &ChaosKind{
		Chaos:     &TimeChaos{},
		GenericChaosList: &TimeChaosList{},
	})


	allScheduleItem.register(KindAWSChaos, &ChaosKind{
		Chaos:     &AWSChaos{},
		GenericChaosList: &AWSChaosList{},
	})

	allScheduleItem.register(KindDNSChaos, &ChaosKind{
		Chaos:     &DNSChaos{},
		GenericChaosList: &DNSChaosList{},
	})

	allScheduleItem.register(KindGCPChaos, &ChaosKind{
		Chaos:     &GCPChaos{},
		GenericChaosList: &GCPChaosList{},
	})

	allScheduleItem.register(KindHTTPChaos, &ChaosKind{
		Chaos:     &HTTPChaos{},
		GenericChaosList: &HTTPChaosList{},
	})

	allScheduleItem.register(KindIOChaos, &ChaosKind{
		Chaos:     &IOChaos{},
		GenericChaosList: &IOChaosList{},
	})

	allScheduleItem.register(KindJVMChaos, &ChaosKind{
		Chaos:     &JVMChaos{},
		GenericChaosList: &JVMChaosList{},
	})

	allScheduleItem.register(KindKernelChaos, &ChaosKind{
		Chaos:     &KernelChaos{},
		GenericChaosList: &KernelChaosList{},
	})

	allScheduleItem.register(KindNetworkChaos, &ChaosKind{
		Chaos:     &NetworkChaos{},
		GenericChaosList: &NetworkChaosList{},
	})

	allScheduleItem.register(KindPodChaos, &ChaosKind{
		Chaos:     &PodChaos{},
		GenericChaosList: &PodChaosList{},
	})

	allScheduleItem.register(KindStressChaos, &ChaosKind{
		Chaos:     &StressChaos{},
		GenericChaosList: &StressChaosList{},
	})

	allScheduleItem.register(KindTimeChaos, &ChaosKind{
		Chaos:     &TimeChaos{},
		GenericChaosList: &TimeChaosList{},
	})

	allScheduleItem.register(KindWorkflow, &ChaosKind{
		Chaos:     &Workflow{},
		GenericChaosList: &WorkflowList{},
	})

}
