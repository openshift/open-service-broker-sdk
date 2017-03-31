package operations

import (
	"errors"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kapi "k8s.io/client-go/pkg/api/v1"

	brokerapi "github.com/openshift/brokersdk/pkg/apis/broker/v1alpha1"
	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

func (b *BrokerOperations) LastOperation(instance_id string, operation openservicebroker.Operation) *openservicebroker.Response {
	if operation != openservicebroker.OperationProvisioning {
		return &openservicebroker.Response{http.StatusBadRequest, nil, errors.New("invalid operation")}
	}

	si, err := b.Client.ServiceInstances("brokersdk").Get(instance_id, metav1.GetOptions{})
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

	state := openservicebroker.LastOperationStateInProgress
	for _, condition := range si.Status.Conditions {
		if condition.Type == brokerapi.ServiceInstanceReady && condition.Status == kapi.ConditionTrue {
			state = openservicebroker.LastOperationStateSucceeded
		}
		if condition.Type == brokerapi.ServiceInstanceFailed && condition.Status == kapi.ConditionTrue {
			state = openservicebroker.LastOperationStateFailed
		}
	}

	/*
		for _, condition := range tp.Status.Conditions {
			if condition.Type == templateapi.TemplateProvisionCreated && condition.Status == api.ConditionTrue {
				state = openservicebroker.LastOperationStateSucceeded
				break
			}
			if condition.Type == templateapi.TemplateProvisionFailed && condition.Status == api.ConditionTrue {
				state = openservicebroker.LastOperationStateFailed
				break
			}
		}
	*/
	return &openservicebroker.Response{http.StatusOK, &openservicebroker.LastOperationResponse{State: state}, nil}
}
