package operations

import (
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapi "k8s.io/client-go/pkg/api"

	"github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
)

// LastOperation is an implementation of the service broker last operation api.
func (b *BrokerOperations) LastOperation(instanceID string, operation openservicebroker.Operation) *openservicebroker.Response {
	// Find the ServiceInstance that represents the state of this service instanceid
	si, err := b.Client.Broker().ServiceInstances(broker.Namespace).Get(instanceID, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			if operation == openservicebroker.OperationDeprovisioning {
				return &openservicebroker.Response{Code: http.StatusGone, Body: &struct{}{}, Err: nil}
			}
			return &openservicebroker.Response{Code: http.StatusBadRequest, Body: nil, Err: err}

		}
		return &openservicebroker.Response{Code: http.StatusInternalServerError, Body: nil, Err: err}
	}

	// Check the conditions on the ServiceInstance to determine the operation state.
	// If there are no conditions, the controller has not processes the instance yet,
	// so it's in progress.  Otherwise there will be a ready or failed condition present.
	state := openservicebroker.LastOperationStateInProgress
	for _, condition := range si.Status.Conditions {
		if condition.Type == broker.ServiceInstanceReady && condition.Status == kapi.ConditionTrue {
			state = openservicebroker.LastOperationStateSucceeded
		}
		if condition.Type == broker.ServiceInstanceFailed && condition.Status == kapi.ConditionTrue {
			state = openservicebroker.LastOperationStateFailed
		}
	}

	return &openservicebroker.Response{Code: http.StatusOK, Body: &openservicebroker.LastOperationResponse{State: state}, Err: nil}
}
