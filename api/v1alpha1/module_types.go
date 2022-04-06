/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PullOptions struct {
	// +optional

	// When Insecure is true, images can be pulled from an insecure (plain HTTP) registry.
	Insecure bool `json:"insecure"`
}

type PushOptions struct {
	// +optional

	// When Insecure is true, built images can be pushed to an insecure (plain HTTP) registry.
	Insecure bool `json:"insecure"`
}

type Build struct {
	Dockerfile string `json:"dockerfile"`

	// +optional

	// Pull contains settings determining how to check if the DriverContainer image already exists.
	Pull PullOptions `json:"pull"`

	// +optional

	// Push contains settings determining how to push a built DriverContainer image.
	Push PushOptions `json:"push"`
}

// KernelMapping pairs kernel versions with a DriverContainer image.
// Kernel versions can be matched literally or using a regular expression.
type KernelMapping struct {
	// +optional

	// Build enables in-cluster builds for this mapping and allows overriding the Module's build settings.
	Build *Build `json:"build"`

	// ContainerImage is the name of the DriverContainer image that should be used to deploy the module.
	ContainerImage string `json:"containerImage"`

	// +optional

	// Literal defines a literal target kernel version to be matched exactly against node kernels.
	Literal string `json:"literal"`

	// +optional

	// Regexp is a regular expression to be match against node kernels.
	Regexp string `json:"regexp"`
}

// ModuleSpec describes how the OOT operator should deploy a Module on those nodes that need it.
type ModuleSpec struct {
	// +optional
	Build Build `json:"build"`

	// +optional

	// AdditionalVolumes is a list of volumes that will be attached to the DriverContainer / DevicePlugin pod,
	// in addition to the default ones.
	AdditionalVolumes []v1.Volume `json:"additionalVolumes"`

	// +optional

	// DevicePlugin allows overriding some properties of the container that deploys the device plugin on the node.
	// Name is ignored and is set automatically by the OOT Operator.
	DevicePlugin *v1.Container `json:"devicePlugin"`

	// DriverContainer allows overriding some properties of the container that deploys the driver on the node.
	// Name and image are ignored and are set automatically by the OOT Operator.
	DriverContainer v1.Container `json:"driverContainer"`

	// KernelMappings is a list of kernel mappings.
	// When a node's labels match Selector, then the OOT Operator will look for the first mapping that matches its
	// kernel version, and use the corresponding container image to run the DriverContainer.
	KernelMappings []KernelMapping `json:"kernelMappings"`

	// Selector describes on which nodes the Module should be loaded.
	Selector map[string]string `json:"selector"`
}

// ModuleStatus defines the observed state of Module.
type ModuleStatus struct {
	// Conditions is a list of conditions representing the Module's current state.
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// Module is the Schema for the modules API
type Module struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModuleSpec   `json:"spec,omitempty"`
	Status ModuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModuleList contains a list of Module
type ModuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Module `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Module{}, &ModuleList{})
}