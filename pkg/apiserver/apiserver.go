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

package apiserver

import (
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"k8s.io/client-go/pkg/version"

	"github.com/openshift/brokersdk/pkg/apis/broker"
	"github.com/openshift/brokersdk/pkg/apis/broker/install"
	"github.com/openshift/brokersdk/pkg/apis/broker/v1alpha1"
	clientset "github.com/openshift/brokersdk/pkg/client/clientset_generated/clientset"
	"github.com/openshift/brokersdk/pkg/openservicebroker"
	"github.com/openshift/brokersdk/pkg/openservicebroker/operations"
	"github.com/openshift/brokersdk/pkg/registry/broker/serviceinstance"
)

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie("")
	Scheme               = runtime.NewScheme()
	Codecs               = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(groupFactoryRegistry, registry, Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

// ServiceCatalogAPIServer contains the base GenericAPIServer along with other
// configured runtime configuration
type ServiceCatalogAPIServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// Config contains a generic API server Config along with config specific to
// the service catalog API server.
type Config struct {
	GenericConfig *genericapiserver.Config

	// BABYNETES: cargo culted from master.go
	/*
		APIResourceConfigSource genericapiserverstorage.APIResourceConfigSource
		DeleteCollectionWorkers int
		StorageFactory          genericapiserverstorage.StorageFactory
	*/
}

// CompletedConfig is an internal type to take advantage of typechecking in
// the type system. mhb does not like it.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have valid data
// and can be derived from other fields.
func (c *Config) Complete() CompletedConfig {
	c.GenericConfig.Complete()

	version := version.Get()
	// Setting this var enables the version resource. We should populate the
	// fields of the object from above if we wish to have our own output. Or
	// establish our own version object somewhere else.
	c.GenericConfig.Version = &version

	return CompletedConfig{c}
}

// New creates the server to run.
func (c CompletedConfig) New() (*ServiceCatalogAPIServer, error) {
	// we need to call new on a "completed" config, which we
	// should already have, as this is a 'CompletedConfig' and the
	// only way to get here from there is by Complete()'ing. Thus
	// we skip the complete on the underlying config and go
	// straight to running it's New() method.
	genericServer, err := c.Config.GenericConfig.SkipComplete().New()
	if err != nil {
		return nil, err
	}

	glog.V(4).Infoln("Creating API server")

	s := &ServiceCatalogAPIServer{
		GenericAPIServer: genericServer,
	}

	// Not every API group compiled in is necessarily enabled by the operator
	// at runtime.
	//
	// Install the API resource config source, which describes versions of
	// which API groups are enabled.
	//c.APIResourceConfigSource = DefaultAPIResourceConfigSource()

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(broker.GroupName, registry, Scheme, metav1.ParameterCodec, Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = v1alpha1.SchemeGroupVersion
	v1alpha1storage := map[string]rest.Storage{}
	v1alpha1storage["serviceinstances"] = serviceinstance.NewREST(Scheme, c.GenericConfig.RESTOptionsGetter)
	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage
	/*
		storageConfig, _ := c.StorageFactory.NewConfig(broker.Resource("serviceinstance"))
		opts := generic.RESTOptions{
			StorageConfig:           storageConfig,
			Decorator:               generic.UndecoratedStorage,
			DeleteCollectionWorkers: c.DeleteCollectionWorkers,
			//EnableGarbageCollection: c.GenericConfig.EnableGarbageCollection,
			ResourcePrefix: c.StorageFactory.ResourcePrefix(broker.Resource("serviceinstance")),
		}
		v1alpha1storage["serviceinstance"] = serviceinstance.NewStorage(opts)

		apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1storage
	*/

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	/*
		kubeconfig, err := clientcmd.BuildConfigFromFlags("https://127.0.0.1:443", "")
		if err != nil {
			glog.Errorf("Failed to create a kube config\n:%v\n", err)

		}

		kubeconfig.Insecure = true
	*/

	//brokerClient, err := clientset.NewForConfig(kubeconfig)
	brokerClient, err := clientset.NewForConfig(s.GenericAPIServer.LoopbackClientConfig)

	// install the broker spec apis
	//s.GenericAPIServer.HandlerContainer.Add
	broker := &operations.BrokerOperations{Client: brokerClient}
	openservicebroker.Route(s.GenericAPIServer.HandlerContainer.Container, "/broker/my.broker.io", broker)

	glog.Infoln("Finished installing API groups")

	return s, nil
}
