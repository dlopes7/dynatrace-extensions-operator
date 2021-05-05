/*
Copyright 2021.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExtensionsSpec defines the desired state of Extension
type ExtensionsSpec struct {

	// ExtensionLinks is a list of extensions that should be installed
	Extensions []ExtensionSpec `json:"extensions,omitempty"`
}

type ExtensionSpec struct {
	Name         string `json:"name,omitempty"`
	DownloadLink string `json:"downloadLink,omitempty"`
}

// ExtensionStatus defines the observed state of Extension
type ExtensionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Instances map[string]ExtensionInstance `json:"instances,omitempty"`
}

type ExtensionInstance struct {
	PodName   string `json:"podName,omitempty"`
	IPAddress string `json:"ipAddress,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Extension is the Schema for the extensions API
type Extension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtensionsSpec  `json:"spec,omitempty"`
	Status ExtensionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExtensionList contains a list of Extension
type ExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Extension `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Extension{}, &ExtensionList{})
}
