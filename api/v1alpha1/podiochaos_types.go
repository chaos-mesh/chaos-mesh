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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PodIoChaosSpec defines the desired state of IoChaos
type PodIoChaosSpec struct {
	// VolumeMountPath represents the target mount path
	// It must be a root of mount path now.
	// TODO: search the mount parent of any path automatically.
	// TODO: support multiple different volume mount path in one pod
	VolumeMountPath string `json:"volumeMountPath"`

	// TODO: support multiple different container to inject in one pod
	// +optional
	Container *string `json:"container,omitempty"`

	// Pid represents a running toda process id
	// +optional
	Pid int64 `json:"pid,omitempty"`

	// StartTime represents the start time of a toda process
	// +optional
	StartTime int64 `json:"startTime,omitempty"`

	// Actions are a list of IoChaos actions
	// +optional
	Actions []IoChaosAction `json:"actions,omitempty"`
}

// IoChaosAction defines an possible action of IoChaos
type IoChaosAction struct {
	Type IoChaosType `json:"type"`

	Filter `json:",inline"`

	// Faults represents the fault to inject
	// +optional
	Faults []IoFault `json:"faults,omitempty"`

	// Latency represents the latency to inject
	// +optional
	Latency string `json:"latency,omitempty"`

	// AttrOverride represents the attribution to override
	// +optional
	*AttrOverrideSpec `json:",inline"`

	// Source represents the source of current rules
	Source string `json:"source,omitempty"`
}

// IoChaosType represents the type of an IoChaos Action
type IoChaosType string

const (
	// IoLatency represents injecting latency for io operation
	IoLatency IoChaosType = "latency"

	// IoFaults represents injecting faults for io operation
	IoFaults IoChaosType = "fault"

	// IoAttrOverride represents replacing attribution for io operation
	IoAttrOverride IoChaosType = "attrOverride"
)

// Filter represents a filter of IoChaos action, which will define the
// scope of an IoChaosAction
type Filter struct {
	// Path represents a glob of injecting path
	Path string `json:"path"`

	// Methods represents the method that the action will inject in
	// +optional
	Methods []IoMethod `json:"methods,omitempty"`

	// Percent represents the percent probability of injecting this action
	Percent int `json:"percent"`
}

// IoFault represents the fault to inject and their weight
type IoFault struct {
	Errno  uint32 `json:"errno"`
	Weight int32  `json:"weight"`
}

// AttrOverrideSpec represents an override of attribution
type AttrOverrideSpec struct {
	//+optional
	Ino *uint64 `json:"ino,omitempty"`
	//+optional
	Size *uint64 `json:"size,omitempty"`
	//+optional
	Blocks *uint64 `json:"blocks,omitempty"`
	//+optional
	Atime *Timespec `json:"atime,omitempty"`
	//+optional
	Mtime *Timespec `json:"mtime,omitempty"`
	//+optional
	Ctime *Timespec `json:"ctime,omitempty"`
	//+optional
	Kind *FileType `json:"kind,omitempty"`
	//+optional
	Perm *uint16 `json:"perm,omitempty"`
	//+optional
	Nlink *uint32 `json:"nlink,omitempty"`
	//+optional
	UID *uint32 `json:"uid,omitempty"`
	//+optional
	GID *uint32 `json:"gid,omitempty"`
	//+optional
	Rdev *uint32 `json:"rdev,omitempty"`
}

// Timespec represents a time
type Timespec struct {
	Sec  int64 `json:"sec"`
	Nsec int64 `json:"nsec"`
}

// FileType represents type of a file
type FileType string

const (
	NamedPipe   FileType = "namedPipe"
	CharDevice  FileType = "charDevice"
	BlockDevice FileType = "blockDevice"
	Directory   FileType = "directory"
	RegularFile FileType = "regularFile"
	TSymlink    FileType = "symlink"
	Socket      FileType = "socket"
)

type IoMethod string

const (
	LookUp      IoMethod = "lookup"
	Forget      IoMethod = "forget"
	GetAttr     IoMethod = "getattr"
	SetAttr     IoMethod = "setattr"
	ReadLink    IoMethod = "readlink"
	Mknod       IoMethod = "mknod"
	Mkdir       IoMethod = "mkdir"
	UnLink      IoMethod = "unlink"
	Rmdir       IoMethod = "rmdir"
	MSymlink    IoMethod = "symlink"
	Rename      IoMethod = "rename"
	Link        IoMethod = "link"
	Open        IoMethod = "open"
	Read        IoMethod = "read"
	Write       IoMethod = "write"
	Flush       IoMethod = "flush"
	Release     IoMethod = "release"
	Fsync       IoMethod = "fsync"
	Opendir     IoMethod = "opendir"
	Readdir     IoMethod = "readdir"
	Releasedir  IoMethod = "releasedir"
	Fsyncdir    IoMethod = "fsyncdir"
	Statfs      IoMethod = "statfs"
	SetXAttr    IoMethod = "setxattr"
	GetXAttr    IoMethod = "getxattr"
	ListXAttr   IoMethod = "listxattr"
	RemoveXAttr IoMethod = "removexattr"
	Access      IoMethod = "access"
	Create      IoMethod = "create"
	GetLk       IoMethod = "getlk"
	SetLk       IoMethod = "setlk"
	Bmap        IoMethod = "bmap"
)

// +kubebuilder:object:root=true

// PodIoChaos is the Schema for the podiochaos API
type PodIoChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PodIoChaosSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// PodIoChaosList contains a list of PodIoChaos
type PodIoChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodIoChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodIoChaos{}, &PodIoChaosList{})
}
