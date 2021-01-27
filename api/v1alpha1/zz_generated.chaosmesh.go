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

package v1alpha1

import (
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const KindAwsChaos = "AwsChaos"

// IsDeleted returns whether this resource has been deleted
func (in *AwsChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *AwsChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetDuration would return the duration for chaos
func (in *AwsChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *AwsChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *AwsChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *AwsChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *AwsChaos) SetNextRecover(t time.Time) {
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
func (in *AwsChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *AwsChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindAwsChaos,
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
func (in *AwsChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// AwsChaosList contains a list of AwsChaos
type AwsChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *AwsChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *DNSChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *DNSChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *DNSChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *DNSChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *DNSChaos) SetNextRecover(t time.Time) {
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
func (in *DNSChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *DNSChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindDNSChaos,
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
func (in *DNSChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// DNSChaosList contains a list of DNSChaos
type DNSChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNSChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *DNSChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *HTTPChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *HTTPChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *HTTPChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *HTTPChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *HTTPChaos) SetNextRecover(t time.Time) {
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
func (in *HTTPChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *HTTPChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindHTTPChaos,
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
func (in *HTTPChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// HTTPChaosList contains a list of HTTPChaos
type HTTPChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *HTTPChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

const KindIoChaos = "IoChaos"

// IsDeleted returns whether this resource has been deleted
func (in *IoChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *IoChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetDuration would return the duration for chaos
func (in *IoChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *IoChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *IoChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *IoChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *IoChaos) SetNextRecover(t time.Time) {
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
func (in *IoChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *IoChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindIoChaos,
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
func (in *IoChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// IoChaosList contains a list of IoChaos
type IoChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IoChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *IoChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *JVMChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *JVMChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *JVMChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *JVMChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *JVMChaos) SetNextRecover(t time.Time) {
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
func (in *JVMChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *JVMChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindJVMChaos,
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
func (in *JVMChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// JVMChaosList contains a list of JVMChaos
type JVMChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JVMChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *JVMChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *KernelChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *KernelChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *KernelChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *KernelChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *KernelChaos) SetNextRecover(t time.Time) {
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
func (in *KernelChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *KernelChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindKernelChaos,
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
func (in *KernelChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// KernelChaosList contains a list of KernelChaos
type KernelChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KernelChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *KernelChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *NetworkChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *NetworkChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *NetworkChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *NetworkChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *NetworkChaos) SetNextRecover(t time.Time) {
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
func (in *NetworkChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *NetworkChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindNetworkChaos,
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
func (in *NetworkChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// NetworkChaosList contains a list of NetworkChaos
type NetworkChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *NetworkChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *PodChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *PodChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *PodChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *PodChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *PodChaos) SetNextRecover(t time.Time) {
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
func (in *PodChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *PodChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindPodChaos,
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
func (in *PodChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// PodChaosList contains a list of PodChaos
type PodChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *PodChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *StressChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *StressChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *StressChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *StressChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *StressChaos) SetNextRecover(t time.Time) {
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
func (in *StressChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *StressChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindStressChaos,
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
func (in *StressChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// StressChaosList contains a list of StressChaos
type StressChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StressChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *StressChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
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

// GetDuration would return the duration for chaos
func (in *TimeChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *TimeChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

func (in *TimeChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

func (in *TimeChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

func (in *TimeChaos) SetNextRecover(t time.Time) {
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
func (in *TimeChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetChaos would return the a record for chaos
func (in *TimeChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindTimeChaos,
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
func (in *TimeChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// +kubebuilder:object:root=true

// TimeChaosList contains a list of TimeChaos
type TimeChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimeChaos `json:"items"`
}

// ListChaos returns a list of chaos
func (in *TimeChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {

	SchemeBuilder.Register(&AwsChaos{}, &AwsChaosList{})
	all.register(KindAwsChaos, &ChaosKind{
		Chaos:     &AwsChaos{},
		ChaosList: &AwsChaosList{},
	})

	SchemeBuilder.Register(&DNSChaos{}, &DNSChaosList{})
	all.register(KindDNSChaos, &ChaosKind{
		Chaos:     &DNSChaos{},
		ChaosList: &DNSChaosList{},
	})

	SchemeBuilder.Register(&HTTPChaos{}, &HTTPChaosList{})
	all.register(KindHTTPChaos, &ChaosKind{
		Chaos:     &HTTPChaos{},
		ChaosList: &HTTPChaosList{},
	})

	SchemeBuilder.Register(&IoChaos{}, &IoChaosList{})
	all.register(KindIoChaos, &ChaosKind{
		Chaos:     &IoChaos{},
		ChaosList: &IoChaosList{},
	})

	SchemeBuilder.Register(&JVMChaos{}, &JVMChaosList{})
	all.register(KindJVMChaos, &ChaosKind{
		Chaos:     &JVMChaos{},
		ChaosList: &JVMChaosList{},
	})

	SchemeBuilder.Register(&KernelChaos{}, &KernelChaosList{})
	all.register(KindKernelChaos, &ChaosKind{
		Chaos:     &KernelChaos{},
		ChaosList: &KernelChaosList{},
	})

	SchemeBuilder.Register(&NetworkChaos{}, &NetworkChaosList{})
	all.register(KindNetworkChaos, &ChaosKind{
		Chaos:     &NetworkChaos{},
		ChaosList: &NetworkChaosList{},
	})

	SchemeBuilder.Register(&PodChaos{}, &PodChaosList{})
	all.register(KindPodChaos, &ChaosKind{
		Chaos:     &PodChaos{},
		ChaosList: &PodChaosList{},
	})

	SchemeBuilder.Register(&StressChaos{}, &StressChaosList{})
	all.register(KindStressChaos, &ChaosKind{
		Chaos:     &StressChaos{},
		ChaosList: &StressChaosList{},
	})

	SchemeBuilder.Register(&TimeChaos{}, &TimeChaosList{})
	all.register(KindTimeChaos, &ChaosKind{
		Chaos:     &TimeChaos{},
		ChaosList: &TimeChaosList{},
	})

}
