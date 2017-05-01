package operations

import (
	"net/http"

	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
)

// Unbind is an implementation of the service broker unbind api
func (b *BrokerOperations) Unbind(instance_id, binding_id string) *openservicebroker.Response {
	// in principle, unbind should alter state somewhere (e.g. invalidating the credentials
	// associated with this binding, for brokers that provide unique credentials for each binding)
	return &openservicebroker.Response{http.StatusOK, &openservicebroker.UnbindResponse{}, nil}
}
