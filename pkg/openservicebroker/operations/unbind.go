package operations

import (
	"net/http"

	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

func (b *BrokerOperations) Unbind(instance_id, binding_id string) *openservicebroker.Response {
	// TODO: in principle, unbind should alter state somewhere
	return &openservicebroker.Response{http.StatusOK, &openservicebroker.UnbindResponse{}, nil}
}
