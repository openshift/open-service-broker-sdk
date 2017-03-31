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

package controller

import (
	"time"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	kapi "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"

	"k8s.io/kubernetes/pkg/api/unversioned"

	brokerapi "github.com/openshift/brokersdk/pkg/apis/broker/v1alpha1"
	brokerclientset "github.com/openshift/brokersdk/pkg/client/clientset_generated/clientset"
)

// Controller describes a controller that backs the service catalog API for
// Open Service Broker compliant Brokers.
type Controller interface {
	// Run runs the controller until the given stop channel can be read from.
	Run(stopCh <-chan struct{})
}

// controller is a concrete Controller.
type controller struct {
	brokerClient brokerclientset.Clientset
	informer     cache.Controller
}

// NewController returns a new Open Service Broker catalog
// controller.
func NewController(
	brokerClient brokerclientset.Clientset,
) (Controller, error) {

	controller := &controller{
		brokerClient: brokerClient,
	}

	_, controller.informer = cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return brokerClient.ServiceInstances("brokersdk").List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return brokerClient.ServiceInstances("brokersdk").Watch(options)
			},
		},
		&brokerapi.ServiceInstance{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.serviceInstanceAdd,
			DeleteFunc: controller.serviceInstanceDelete,
		},
	)
	return controller, nil
}

// Run runs the controller until the given stop channel can be read from.
func (c *controller) Run(stopCh <-chan struct{}) {
	glog.Info("Starting broker controller")
	c.informer.Run(stopCh)
}

// ServiceInstance handlers
func (c *controller) serviceInstanceAdd(obj interface{}) {
	instance, ok := obj.(*brokerapi.ServiceInstance)
	if instance == nil || !ok {
		return
	}
	glog.Infof("controller sees instance %s", instance.Name)
	condition := brokerapi.ServiceInstanceCondition{
		Type:               brokerapi.ServiceInstanceReady,
		Status:             kapi.ConditionTrue,
		LastTransitionTime: unversioned.Time{time.Now()},
		Reason:             "ServiceProvisioned",
		Message:            "This service has been provisioned",
	}
	instance.Status.Conditions = append(instance.Status.Conditions, condition)
	_, err := c.brokerClient.Broker().ServiceInstances("brokersdk").Update(instance)
	if err != nil {
		glog.Errorf("Error updating service instance %s to ready: %v", instance.Name, err)
	}

	//	c.reconcileInstance(instance)
}

func (c *controller) serviceInstanceDelete(obj interface{}) {
	instance, ok := obj.(*brokerapi.ServiceInstance)
	if instance == nil || !ok {
		return
	}

	//	c.reconcileInstance(instance)
}
