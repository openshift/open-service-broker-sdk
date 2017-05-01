package operations

import (
	"net/http"

	"github.com/golang/glog"

	broker "github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	brokerapi "github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
)

// Provision is an implementation of the service broker provision api
func (b *BrokerOperations) Provision(instance_id string, preq *openservicebroker.ProvisionRequest) *openservicebroker.Response {
	// provision will create a new ServiceInstance resource to be processed
	// by the controller.
	si := brokerapi.ServiceInstance{}
	si.Name = instance_id
	// this credential will be returned to bind requests, in theory it is a value
	// consumers of the service instance will need to access the instance.
	si.Spec.Credential = "some_secret"
	//si.Status.Conditions = append(si.Status.Conditions, brokerapi.ServiceInstanceCondition{})

	// Create the ServiceInstance object that represents this service instance.  The
	// controller will see the request and progress it from there.
	_, err := b.Client.ServiceInstances(broker.Namespace).Create(&si)
	if err != nil {
		glog.Errorf("Failed to create a service instance\n:%v\n", err)
	}

	// Use this for async provision flows.  Technically the service instance isn't provisioned
	// until the controller sees the request, does work, and marks it ready.
	return &openservicebroker.Response{http.StatusAccepted, openservicebroker.ProvisionResponse{Operation: openservicebroker.OperationProvisioning}, nil}

	// For synchronous flows we can just return complete.
	//return &openservicebroker.Response{http.StatusOK, openservicebroker.ProvisionResponse{Operation: openservicebroker.OperationProvisioning}, nil}
}
