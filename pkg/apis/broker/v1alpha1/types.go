/*
Copyright 2017 The Kubernetes Authors.

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

	kapi "k8s.io/client-go/pkg/api/v1"
)

// ServiceInstanceList is a list of ServiceInstance objects.
type ServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ServiceInstance `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient=true

// ServiceInstance represents a service instance provision request,
// possibly fullfilled.
type ServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ServiceInstanceSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ServiceInstanceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ServiceInstanceSpec defines the requested ServiceInstance
type ServiceInstanceSpec struct {
	// Credential is a sample value associatd w/ the provisioned service
	Credential string `json:"credential" protobuf:"bytes,1,opt,name=credential"`
}

// ServiceInstanceStatus defines the current state of the ServiceInstance
type ServiceInstanceStatus struct {
	Conditions []ServiceInstanceCondition `json:"conditions" protobuf:"bytes,1,rep,name=conditions"`
}

// ServiceInstanceCondition contains condition information for a
// ServiceInstance.
type ServiceInstanceCondition struct {
	// Type of the condition, currently Ready or Failure.
	Type ServiceInstanceConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ServiceInstanceConditionType"`
	// Status of the condition, one of True, False or Unknown.
	Status kapi.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
	// LastTransitionTime is the last time a condition status transitioned from
	// one state to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string `json:"reason" protobuf:"bytes,4,opt,name=reason"`
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string `json:"message" protobuf:"bytes,5,opt,name=message"`
}

// ServiceInstanceConditionType is the type of condition pertaining to a
// ServiceInstance.
type ServiceInstanceConditionType string

const (
	// ServiceInstanceReady indicates the readiness of the template
	// instantiation.
	ServiceInstanceReady ServiceInstanceConditionType = "Ready"
	// ServiceInstanceInstantiateFailed indicates the failure of the provision request
	ServiceInstanceFailed ServiceInstanceConditionType = "Failed"

	APIGroupVersion = "v1alpha1"
)
