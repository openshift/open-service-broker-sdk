/*
Copyright 2016 The Kubernetes Authors.

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

package serviceinstance

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/openshift/brokersdk/pkg/apis/broker"
)

var (
	errNotAnInstance = errors.New("not an instance")
)

/*
// NewStorage creates a new rest.Storage for each of Instances and
// Status of Instances
//func NewStorage(opts generic.RESTOptions) (rest.Storage, rest.Storage) {
func NewStorage(opts generic.RESTOptions) rest.Storage {
	prefix := "/" + opts.ResourcePrefix

	newListFunc := func() runtime.Object { return &broker.ServiceInstanceList{} }
	newKeyFunc := func(obj runtime.Object) (string, error) { return obj.(*broker.ServiceInstance).Name, nil }
	storageInterface, dFunc := opts.Decorator(
		runtime.NewScheme(),
		opts.StorageConfig,
		1000,
		&broker.ServiceInstance{},
		prefix,
		newKeyFunc,
		newListFunc,
		nil,
		storage.NoTriggerPublisher,
	)

	store := registry.Store{
		NewFunc: func() runtime.Object {
			return &broker.ServiceInstance{}
		},
		// NewListFunc returns an object capable of storing results of an etcd list.
		NewListFunc: newListFunc,
		// Produces a path that etcd understands, to the root of the resource
		// by combining the namespace in the context with the given prefix
		KeyRootFunc: func(ctx request.Context) string {
			return registry.NamespaceKeyRootFunc(ctx, prefix)
		},
		// Produces a path that etcd understands, to the resource by combining
		// the namespace in the context with the given prefix
		KeyFunc: func(ctx request.Context, name string) (string, error) {
			return registry.NamespaceKeyFunc(ctx, prefix, name)
		},
		// Retrieve the name field of the resource.
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*broker.ServiceInstance).Name, nil
		},
		// Used to match objects based on labels/fields for list.
		PredicateFunc: Match,
		// QualifiedResource should always be plural
		QualifiedResource: api.Resource("instances"),

		CreateStrategy: instanceRESTStrategies,
		UpdateStrategy: instanceRESTStrategies,
		DeleteStrategy: instanceRESTStrategies,

		Storage:     storageInterface,
		DestroyFunc: dFunc,
	}

	//statusStore := store
	//statusStore.UpdateStrategy = instanceStatusUpdateStrategy

	//return &store, &statusStore
	return &store
}
*/

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) rest.Storage {
	strategy := NewStrategy(scheme)

	store := &registry.Store{
		Copier:      scheme,
		NewFunc:     func() runtime.Object { return &broker.ServiceInstance{} },
		NewListFunc: func() runtime.Object { return &broker.ServiceInstanceList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*broker.ServiceInstance).Name, nil
		},
		PredicateFunc:     MatchServiceInstance,
		QualifiedResource: broker.Resource("serviceinstance"),

		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}
	return store
}
