package operations

import (
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
)

// Deprovision is an implementation of the service broker deprovision api.
// It deletes the ServiceInstance associatied with the instanceid being deprovisioned.
// This will trigger the controller to see the service instance as deleted and
// further cleanup can be done there.
func (b *BrokerOperations) Deprovision(instance_id string) *openservicebroker.Response {
	err := b.Client.ServiceInstances(broker.Namespace).Delete(instance_id, &metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &openservicebroker.Response{http.StatusGone, &openservicebroker.DeprovisionResponse{}, nil}
		}
		return &openservicebroker.Response{http.StatusInternalServerError, nil, err}
	}
	return &openservicebroker.Response{http.StatusAccepted, &openservicebroker.DeprovisionResponse{Operation: openservicebroker.OperationDeprovisioning}, nil}
}
