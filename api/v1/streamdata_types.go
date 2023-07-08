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

// StreamdataSpec defines the desired state of Streamdata
type StreamdataSpec struct {
	// +kubebuilder:validation:Required
	StubIp string `json:"stubip,omitempty"`
	// +kubebuilder:validation:Required
	ServerIp string `json:"serverip,omitempty"`
	// +kubebuilder:validation:Required
	ClientIp string `json:"clientip,omitempty"`
	// The port that used by the server.
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	ServerPort int `json:"serverport,omitempty"`
	// The port that used by the client
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	ClientPort int `json:"clientport,omitempty"`
	// The stream state set my the controlplane
	// +kubebuilder:validation:Enum=create;play;teardown
	// +kubebuilder:validation:Required
	StreamState string `json:"streamstate,omitempty"`
	// The node that the stub is on.  Could be derived
	// by calling kube API based on other values in CRD
	NodeID string `json:"nodeid,omitempty"`
}

// StreamdataStatus defines the observed state of Streamdata
type StreamdataStatus struct {
	// The current state of this stream
	// +kubebuilder:validation:Enum=PENDING;SUCCESS;ERROR
	Status string `json:"status,omitempty"`
	// If in error the most recent error will be shown
	Reason string `json:"reason,omitempty"`
	// The current dataplane status as determined by this module
	// +kubebuilder:validation:Enum=PENDING;SUCCESS;ERROR
	StreamStatus string `json:"streamstatus,omitempty"`
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
