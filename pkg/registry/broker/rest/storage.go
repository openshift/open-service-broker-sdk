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

package rest

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the servicecatalog API group.
type StorageProvider struct{}

/*
// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// servicecatalog API group.
func (p StorageProvider) NewRESTStorage(apiResourceConfigSource genericapiserverstorage.APIResourceConfigSource, restOptionsGetter genericapiserveroptions.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(broker.GroupName, registry, Scheme, metav1.ParameterCodec, Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.SchemeGroupVersion
	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		v1alpha1.SchemeGroupVersion.Version: p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter),
	}

	return apiGroupInfo, true
}

func (p StorageProvider) v1alpha1Storage(apiResourceConfigSource genericapiserverstorage.APIResourceConfigSource, restOptionsGetter genericapiserveroptions.RESTOptionsGetter) map[string]rest.Storage {
	getter, _ := restOptionsGetter.GetRESTOptions(broker.Resource("serviceinstances"))
	instances, instancesStatus := serviceinstance.NewStorage(getter)
	return map[string]rest.Storage{
		"serviceinstances":        instances,
		"serviceinstances/status": instancesStatus,
	}
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return broker.GroupName
}
*/
