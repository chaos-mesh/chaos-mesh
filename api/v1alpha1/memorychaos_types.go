package v1alpha1

import (
	"time"

	"github.com/docker/go-units"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// MemoryChaos is the Schema for the memorychaos API
type MemoryChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MemoryChaosSpec   `json:"spec"`
	Status            MemoryChaosStatus `json:"status"`
}

func (in *MemoryChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

func (in *MemoryChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

func (in *MemoryChaos) GetQuota() (int64, error) {
	return units.FromHumanSize(in.Spec.Quota)
}

// GetDuration gets the duration of TimeChaos
func (in *MemoryChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of TimeChaos
func (in *MemoryChaos) GetNextStart() time.Time {
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

// SetNextStart sets NextStart field of TimeChaos
func (in *MemoryChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

// Validate describe the memorychaos validation logic
func (in *MemoryChaos) Validate() (bool, string, error) {
	return true, "", nil
}

// GetNextRecover get NextRecover field of TimeChaos
func (in *MemoryChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

// SetNextRecover sets NextRecover field of TimeChaos
func (in *MemoryChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

// GetScheduler returns the scheduler of TimeChaos
func (in *MemoryChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

type MemoryChaosSpec struct {
	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// Quota specifies the strict memory quota a pod should have.
	// Support units: KiB
	Quota string `json:"quota"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *SchedulerSpec `json:"scheduler"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *MemoryChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *MemoryChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *MemoryChaosSpec) GetValue() string {
	return in.Value
}

type MemoryChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// MemoryChaosList contains a list of MemoryChaos
type MemoryChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MemoryChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MemoryChaos{}, &MemoryChaosList{})
}
