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

package server

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/wait"

	genericapiserver "k8s.io/apiserver/pkg/server"
	genericserveroptions "k8s.io/apiserver/pkg/server/options"

	"github.com/openshift/brokersdk/pkg/apis/broker/v1alpha1"
	"github.com/openshift/brokersdk/pkg/apiserver"
	clientset "github.com/openshift/brokersdk/pkg/client/clientset_generated/clientset"
	"github.com/openshift/brokersdk/pkg/controller"
)

// BrokerServerOptions contains the aggregation of configuration structs for
// the service-catalog server. The theory here is that any future user
// of this server will be able to use this options object as a sub
// options of its own.
type BrokerServerOptions struct {
	// the runtime configuration of our server
	GenericServerRunOptions *genericserveroptions.ServerRunOptions
	// the https configuration. certs, etc
	//ServingOptions *genericserveroptions.ServingOptions
	ServingOptions *genericserveroptions.SecureServingOptions
	// storage with etcd
	EtcdOptions *genericserveroptions.EtcdOptions
	// authn
	AuthenticationOptions *genericserveroptions.DelegatingAuthenticationOptions
	// authz
	AuthorizationOptions *genericserveroptions.DelegatingAuthorizationOptions

	RecommendedOptions *genericserveroptions.RecommendedOptions
}

const (
	// Store generated SSL certificates in a place that won't collide with the
	// k8s core API server.
	certDirectory = "/var/run/brokersdk"

	// I made this up to match some existing paths. I am not sure if there
	// are any restrictions on the format or structure beyond text
	// separated by slashes.
	etcdPathPrefix = "/k8s.io/brokersdk"

	// GroupName I made this up. Maybe we'll need it.
	GroupName = "brokersdk.k8s.io"
)

// NewCommandServer creates a new cobra command to run our server.
func NewCommandServer(out io.Writer) *cobra.Command {
	// initalize our sub options.
	recommended := genericserveroptions.NewRecommendedOptions(etcdPathPrefix, apiserver.Scheme, apiserver.Codecs.LegacyCodec(v1alpha1.SchemeGroupVersion))
	options := &BrokerServerOptions{
		RecommendedOptions:      recommended,
		GenericServerRunOptions: genericserveroptions.NewServerRunOptions(),
		//ServingOptions:          genericserveroptions.NewInsecureServingOptions(),
		ServingOptions:        genericserveroptions.NewSecureServingOptions(),
		EtcdOptions:           recommended.Etcd,
		AuthenticationOptions: genericserveroptions.NewDelegatingAuthenticationOptions(),
		AuthorizationOptions:  genericserveroptions.NewDelegatingAuthorizationOptions(),
		//InsecureServingOptions:  genericserveroptions.NewInsecureServingOptions(),
	}

	// Set generated SSL cert path correctly
	//options.SecureServingOptions.ServerCert.CertDirectory = certDirectory

	// Create the command that runs the API server
	cmd := &cobra.Command{
		Short: "run a brokersdk server",
		RunE: func(c *cobra.Command, args []string) error {
			return options.RunServer(wait.NeverStop)
		},
	}

	// We pass flags object to sub option structs to have them configure
	// themselves. Each options adds its own command line flags
	// in addition to the flags that are defined above.
	flags := cmd.Flags()
	// TODO consider an AddFlags() method on our options
	// struct. Will need to import pflag.
	//
	// repeated pattern seems like it should be refactored if all
	// options were of an interface type that specified AddFlags.
	flags.AddGoFlagSet(flag.CommandLine)
	options.RecommendedOptions.AddFlags(flags)
	return cmd
}

// RunServer is a method on the options for composition. Allows embedding in a
// higher level options as we do the etcd and serving options.
func (serverOptions BrokerServerOptions) RunServer(stopCh <-chan struct{}) error {
	glog.V(4).Infof("Preparing to run API server")

	// server configuration options
	glog.V(4).Infoln("Setting up secure serving options")
	if err := serverOptions.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", net.ParseIP("127.0.0.1")); err != nil {

		//if err := serverOptions.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(serverOptions.GenericServerRunOptions.AdvertiseAddress.String()); err != nil {
		glog.Errorf("Error creating self-signed certificates: %v", err)
		return err
	}

	// config
	glog.V(4).Infoln("Configuring generic API server")
	genericconfig := genericapiserver.NewConfig().WithSerializer(apiserver.Codecs)

	//serverOptions.GenericServerRunOptions.ApplyTo(genericconfig)
	//serverOptions.ServingOptions.ApplyTo(genericconfig)
	serverOptions.RecommendedOptions.ApplyTo(genericconfig)
	//serverOptions.EtcdOptions.ApplyTo(genericconfig)

	// audit logging
	genericconfig.AuditWriter = os.Stdout

	// these are all mutators of each specific suboption in serverOptions object.
	// this repeated pattern seems like we could refactor
	/*
		if _, err := genericconfig.ApplySecureServingOptions(serverOptions.SecureServingOptions); err != nil {
			glog.Errorln(err)
			return err
		}

		genericconfig.ApplyInsecureServingOptions(serverOptions.InsecureServingOptions)
	*/
	glog.V(4).Info("Setting up authn (disabled)")
	// need to figure out what's throwing the `missing clientCA file` err
	/*
		if _, err := genericconfig.ApplyDelegatingAuthenticationOptions(serverOptions.AuthenticationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	glog.V(4).Infoln("Setting up authz (disabled)")
	// having this enabled causes the server to crash for any call
	/*
		if _, err := genericconfig.ApplyDelegatingAuthorizationOptions(serverOptions.AuthorizationOptions); err != nil {
			glog.Infoln(err)
			return err
		}
	*/

	// Set the finalized generic and storage configs
	config := apiserver.Config{
		GenericConfig: genericconfig,
	}

	// Fill in defaults not already set in the config
	completedconfig := config.Complete()

	// make the server
	glog.V(4).Infoln("Completing API server configuration")
	server, err := completedconfig.New()
	glog.Infof("complete, error=%v", err)
	if err != nil {
		return fmt.Errorf("error completing API server configuration: %v", err)
	}

	// I don't like this. We're reaching in too far to call things.
	preparedserver := server.GenericAPIServer.PrepareRun() // post api installation setup? We should have set up the api already?

	brokerClient, err := clientset.NewForConfig(server.GenericAPIServer.LoopbackClientConfig)
	if err != nil {
		glog.Errorf("could not get broker client: %v", err)
	}

	controller, err := controller.NewController(*brokerClient)
	if err != nil {
		glog.Errorf("could not create controller: %v", err)
	}
	go func() {
		controller.Run(stopCh)
	}()

	glog.Infof("Running the API server")
	preparedserver.Run(stopCh)

	return nil
}
