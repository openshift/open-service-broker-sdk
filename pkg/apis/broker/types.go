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
	LastTransitionTime metav1.Time
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
	// ServiceInstanceReady indicates the service instance is Ready for use
	// (provision was successful)
	ServiceInstanceReady ServiceInstanceConditionType = "Ready"

	// ServiceInstanceInstantiateFailed indicates the provision request failed.
	ServiceInstanceFailed ServiceInstanceConditionType = "Failure"

	// TypePackage is the name of the package that defines the resource types
	// used by this broker.
	TypePackage = "github.com/openshift/open-service-broker-sdk/pkg/apis/broker"

	// GroupName is the name of the api group used for resources created/managed
	// by this broker.
	GroupName = "sdkbroker.broker.k8s.io"

	// ServiceInstancesResource is the name of the resource used to represent
	// provision requests(possibly fulfilled) for service instances
	ServiceInstancesResource = "serviceinstances"

	// ServiceInstanceResource is the name of the resource used to represent
	// provision requests(possibly fulfilled) for service instances
	ServiceInstanceResource = "serviceinstance"

	// BrokerAPIPrefix is the route prefix for the open service broker api
	// endpoints (e.g. https://yourhost.com/broker/sdkbroker.broker.io/v2/catalog)
	BrokerAPIPrefix = "/broker/sdkbroker.broker.io"

	// Namespace is the namespace the broker will be deployed in and
	// under which it will create any resources
	Namespace = "brokersdk"
)
