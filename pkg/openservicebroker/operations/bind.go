package operations

import (
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

func (b *BrokerOperations) Bind(instance_id, binding_id string, breq *openservicebroker.BindRequest) *openservicebroker.Response {
	si, err := b.Client.ServiceInstances("brokersdk").Get(instance_id, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &openservicebroker.Response{http.StatusGone, &openservicebroker.BindResponse{}, nil}
		}
		return &openservicebroker.Response{http.StatusInternalServerError, nil, err}
	}

	// TODO: in principle, bind should alter state somewhere

	credentials := map[string]interface{}{}
	// TODO: confirm this API
	// TODO: we're somewhat overloading 'credentials' here
	credentials["credential"] = si.Spec.Credential

	return &openservicebroker.Response{
		Code: http.StatusCreated,
		Body: &openservicebroker.BindResponse{Credentials: credentials},
		Err:  nil,
	}
}
