package operations

import (
	"errors"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapi "k8s.io/client-go/pkg/api"

	"github.com/openshift/brokersdk/pkg/apis/broker"
	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

// LastOperation is an implementation of the service broker last operation api.
func (b *BrokerOperations) LastOperation(instance_id string, operation openservicebroker.Operation) *openservicebroker.Response {
	// The only operation we can check on is a Provision operation.
	if operation != openservicebroker.OperationProvisioning {
		return &openservicebroker.Response{http.StatusBadRequest, nil, errors.New("invalid operation")}
	}

	// Find the ServiceInstance that represents the state of this service instanceid
	si, err := b.Client.ServiceInstances(broker.Namespace).Get(instance_id, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			if operation == openservicebroker.OperationDeprovisioning {
				return &openservicebroker.Response{http.StatusGone, &struct{}{}, nil}
			} else {
				return &openservicebroker.Response{http.StatusBadRequest, nil, err}
			}
		}
		return &openservicebroker.Response{http.StatusInternalServerError, nil, err}
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

	return &openservicebroker.Response{http.StatusOK, &openservicebroker.LastOperationResponse{State: state}, nil}
}
