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

package broker

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapi "k8s.io/client-go/pkg/api"

	"k8s.io/kubernetes/pkg/api/unversioned"
)

// ServiceInstanceList is a list of ServiceInstance objects.
type ServiceInstanceList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []ServiceInstance
}

// +genclient=true

// ServiceInstance represents a service instance provision request,
// possibly fullfilled.
type ServiceInstance struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   ServiceInstanceSpec
	Status ServiceInstanceStatus
}

// ServiceInstanceSpec defines the requested ServiceInstance
type ServiceInstanceSpec struct {
	Credential string
}

// ServiceInstanceStatus defines the current state of the ServiceInstance
type ServiceInstanceStatus struct {
	Conditions []ServiceInstanceCondition
}

// ServiceInstanceCondition contains condition information for a
// ServiceInstance.
type ServiceInstanceCondition struct {
	// Type of the condition, currently Ready or InstantiateFailure.
	Type ServiceInstanceConditionType
	// Status of the condition, one of True, False or Unknown.
	Status kapi.ConditionStatus
	// LastTransitionTime is the last time a condition status transitioned from
	// one state to another.
	LastTransitionTime unversioned.Time
	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	Reason string
	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	Message string
}

// ServiceInstanceConditionType is the type of condition pertaining to a
// ServiceInstance.
type ServiceInstanceConditionType string

const (
	// ServiceInstanceReady indicates the readiness of the template
	// instantiation.
	ServiceInstanceReady ServiceInstanceConditionType = "Ready"
	// ServiceInstanceInstantiateFailed indicates the failure of the provision request
	ServiceInstanceFailed ServiceInstanceConditionType = "Failure"
)
