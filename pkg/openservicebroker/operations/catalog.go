package operations

import (
	"net/http"

	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

func (b *BrokerOperations) Catalog() *openservicebroker.Response {

	services := make([]*openservicebroker.Service, 1)

	service_metadata := make(map[string]interface{})
	service_metadata["metadata_key1"] = "metadata_value1"

	service_plans := make([]openservicebroker.Plan, 1)
	service_plans[0] = openservicebroker.Plan{
		Name:        "gold plan",
		ID:          "gold_plan_id",
		Description: "gold plan description",
		Bindable:    true,
		Free:        true,
	}
	services[0] = &openservicebroker.Service{
		Name:        "service name",
		ID:          "serviceUUID",
		Description: "service description",
		Tags:        []string{"tag1", "tag2"},
		Bindable:    true,
		Metadata:    service_metadata,
		Plans:       service_plans,
	}
	// return &openservicebroker.Response{http.StatusInternalServerError, nil, err}
	return &openservicebroker.Response{http.StatusOK, &openservicebroker.CatalogResponse{Services: services}, nil}
}
