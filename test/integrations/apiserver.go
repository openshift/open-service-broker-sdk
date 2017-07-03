package integrations

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	cmdserver "github.com/openshift/open-service-broker-sdk/cmd/broker/server"
	apiserver "github.com/openshift/open-service-broker-sdk/pkg/apiserver"
	clientset "github.com/openshift/open-service-broker-sdk/pkg/client/clientset_generated/internalclientset"
	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/dynamic"
)

// DefaultServerConfig sets up a config for integration tests
func DefaultServerConfig() (*apiserver.Config, error) {
	port, err := findFreeLocalPort()
	if err != nil {
		return nil, err
	}
	options := cmdserver.NewBrokerServerOptions()
	options.RecommendedOptions.SecureServing.BindPort = port
	options.RecommendedOptions.Authentication.SkipInClusterLookup = true
	options.RecommendedOptions.SecureServing.BindAddress = net.ParseIP("127.0.0.1")
	etcdURL, ok := os.LookupEnv("KUBE_INTEGRATION_ETCD_URL")
	if !ok {
		etcdURL = "http://127.0.0.1:2379"
	}
	options.RecommendedOptions.Etcd.StorageConfig.ServerList = []string{etcdURL}
	options.RecommendedOptions.Etcd.StorageConfig.Prefix = uuid.New()
	genericConfig := genericapiserver.NewConfig(apiserver.Codecs)
	genericConfig.Authenticator = nil
	genericConfig.Authorizer = authorizerfactory.NewAlwaysAllowAuthorizer()
	if err := options.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}
	if err := options.RecommendedOptions.Etcd.ApplyTo(genericConfig); err != nil {
		return nil, err
	}
	if err := options.RecommendedOptions.SecureServing.ApplyTo(genericConfig); err != nil {
		return nil, err
	}
	if err := options.RecommendedOptions.Audit.ApplyTo(genericConfig); err != nil {
		return nil, err
	}
	if err := options.RecommendedOptions.Features.ApplyTo(genericConfig); err != nil {
		return nil, err
	}
	return &apiserver.Config{
		GenericConfig: genericConfig,
	}, nil

}

func startServer(config *apiserver.Config) (chan struct{}, clientset.Interface, dynamic.ClientPool, error) {
	stopCh := make(chan struct{})
	server, err := config.Complete().New()
	if err != nil {
		return nil, nil, nil, err
	}
	go func() {
		err := server.GenericAPIServer.PrepareRun().Run(stopCh)
		if err != nil {
			close(stopCh)
			panic(err)
		}
	}()

	// wait until the server is healthy
	err = wait.PollImmediate(30*time.Millisecond, 30*time.Second, func() (bool, error) {
		healthClient, err := clientset.NewForConfig(server.GenericAPIServer.LoopbackClientConfig)
		if err != nil {
			return false, nil
		}
		healthResult := healthClient.Discovery().RESTClient().Get().AbsPath("/healthz").Do()
		if healthResult.Error() != nil {
			return false, nil
		}
		rawHealth, err := healthResult.Raw()
		if err != nil {
			return false, nil
		}
		if string(rawHealth) != "ok" {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		close(stopCh)
		return nil, nil, nil, err
	}

	brokerClient, err := clientset.NewForConfig(server.GenericAPIServer.LoopbackClientConfig)
	if err != nil {
		close(stopCh)
		return nil, nil, nil, err
	}

	bytes, _ := brokerClient.Discovery().RESTClient().Get().AbsPath("/apis/sdkbroker.broker.k8s.io/v1alpha1").DoRaw()
	fmt.Print(string(bytes))
	return stopCh, brokerClient, dynamic.NewDynamicClientPool(server.GenericAPIServer.LoopbackClientConfig), nil
}

// StartDefaultServer starts the api server outside of a kubernetes cluster and allows for integration tests to be run
func StartDefaultServer() (chan struct{}, clientset.Interface, dynamic.ClientPool, error) {
	config, err := DefaultServerConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	return startServer(config)
}

// findFreeLocalPort returns the number of an available port number on
// the loopback interface.  Useful for determining the port to launch
// a server on.  Error handling required - there is a non-zero chance
// that the returned port number will be bound by another process
// after this function returns.
func findFreeLocalPort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	_, portStr, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return port, nil
}
