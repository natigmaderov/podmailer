/*
Copyright 2025.

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

// SMTPConfig defines the SMTP server configuration
type SMTPConfig struct {
	// Server is the SMTP server address
	Server string `json:"server"`

	// Port is the SMTP server port
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// Username for SMTP authentication
	Username string `json:"username"`

	// Password for SMTP authentication
	Password string `json:"password"`

	// FromEmail is the sender's email address
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	FromEmail string `json:"fromEmail"`
}

// PodMailerSpec defines the desired state of PodMailer
type PodMailerSpec struct {
	// SMTPConfig contains the SMTP server configuration
	SMTP SMTPConfig `json:"smtp"`

	// Recipients is a list of email addresses to notify
	// +kubebuilder:validation:MinItems=1
	Recipients []string `json:"recipients"`

	// Namespaces is a list of namespaces to monitor
	// If empty, all namespaces will be monitored
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// CheckInterval is the interval between pod status checks in seconds
	// +kubebuilder:validation:Minimum=30
	// +kubebuilder:default=60
	CheckInterval int32 `json:"checkInterval,omitempty"`
}

// PodStatus represents the status of a monitored pod
type PodStatus struct {
	// Name of the pod
	Name string `json:"name"`

	// Namespace of the pod
	Namespace string `json:"namespace"`

	// Status of the pod
	Status string `json:"status"`

	// LastNotificationTime is the timestamp of the last notification sent for this pod
	LastNotificationTime *metav1.Time `json:"lastNotificationTime,omitempty"`
}

// PodMailerStatus defines the observed state of PodMailer
type PodMailerStatus struct {
	// LastCheckTime is the last time the pods were checked
	LastCheckTime *metav1.Time `json:"lastCheckTime,omitempty"`

	// DownPods contains the list of pods that are currently down
	DownPods []PodStatus `json:"downPods,omitempty"`

	// LastNotificationTime is the last time an email notification was sent
	LastNotificationTime *metav1.Time `json:"lastNotificationTime,omitempty"`

	// Conditions represent the latest available observations of the PodMailer's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PodMailer is the Schema for the podmailers API.
type PodMailer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodMailerSpec   `json:"spec,omitempty"`
	Status PodMailerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodMailerList contains a list of PodMailer.
type PodMailerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodMailer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodMailer{}, &PodMailerList{})
}
