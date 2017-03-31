package operations

import (
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

const uuidMapGracePeriod = 30

func (b *BrokerOperations) Deprovision(instance_id string) *openservicebroker.Response {
	err := b.Client.ServiceInstances("brokersdk").Delete(instance_id, &metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &openservicebroker.Response{http.StatusGone, &openservicebroker.DeprovisionResponse{}, nil}
		}
		return &openservicebroker.Response{http.StatusInternalServerError, nil, err}
	}
	return &openservicebroker.Response{http.StatusAccepted, &openservicebroker.DeprovisionResponse{Operation: openservicebroker.OperationDeprovisioning}, nil}
}
