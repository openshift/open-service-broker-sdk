package operations

import (
	clientset "github.com/openshift/brokersdk/pkg/client/clientset_generated/clientset"
)

type BrokerOperations struct {
	Client *clientset.Clientset
}
