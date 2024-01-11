/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StreamdataSpec defines the desired state of Streamdata
type StreamdataSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Streamdata. Edit streamdata_types.go to remove/update
	StubIp   string `json:"stubip"`
	ServerIp string `json:"serverip"`
	ClientIp string `json:"clientip"`
	// The list of server ports associated with this stream
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +kubebuilder:validation:MinItems=1
	// TODO confirm omitempty
	ServerPorts []*Port `json:"serverports,omitempty"`
	// The list of client ports associated with this stream
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +kubebuilder:validation:MinItems=1
	ClientPorts []*Port `json:"clientports,omitempty"`
	StreamState int     `json:"streamstate,omitempty"`
}

// Port contains information about a port.
type Port struct {
	// The name of this port. It is optional.  In the future it may
	// have DNS semantics.
	// +optional
	Name string `json:"name,omitempty"`
	// The IP protocol for this port.
	// +optional
	// +kubebuilder:default=UDP
	Protocol string `json:"protocol,omitempty"`
	// The port that used by the client or server.
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=0
	Port int32 `json:"port"`
}

// StreamdataStatus defines the observed state of Streamdata
type StreamdataStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Enum=PENDING;SUCCESS;ERROR
	Status string              `json:"status,omitempty"`
	Reason string              `json:"reason,omitempty"`
	Items  map[string][]string `json:"items,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Streamdata is the Schema for the streamdata API
type Streamdata struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StreamdataSpec   `json:"spec,omitempty"`
	Status StreamdataStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StreamdataList contains a list of Streamdata
type StreamdataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Streamdata `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Streamdata{}, &StreamdataList{})
}
