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
func (b *BrokerOperations) Deprovision(instanceID string) *openservicebroker.Response {
	err := b.Client.Broker().ServiceInstances(broker.Namespace).Delete(instanceID, &metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &openservicebroker.Response{Code: http.StatusGone, Body: &openservicebroker.DeprovisionResponse{}, Err: nil}
		}
		return &openservicebroker.Response{Code: http.StatusInternalServerError, Body: nil, Err: err}
	}
	return &openservicebroker.Response{Code: http.StatusAccepted, Body: &openservicebroker.DeprovisionResponse{Operation: openservicebroker.OperationDeprovisioning}, Err: nil}
}
