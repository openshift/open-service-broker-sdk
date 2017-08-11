package integration

import (
	"encoding/json"
	"testing"

	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
	"github.com/pborman/uuid"
)

func TestProvisionEndpoint(t *testing.T) {
	stopCh, brokerClient, _, err := StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)
	sid := uuid.NewUUID().String()
	cases := []struct {
		Name         string
		ExpectError  bool
		ProvisionReq *openservicebroker.ProvisionRequest
		Code         int
		ServiceID    string
		Operation    openservicebroker.Operation
	}{
		{
			Name: "it should provision ok",
			ProvisionReq: &openservicebroker.ProvisionRequest{
				Context: openservicebroker.KubernetesContext{
					Platform:  "kubernetes",
					Namespace: "test",
				},
				AcceptsIncomplete: true,
				PlanID:            "thePlan",
				OrganizationID:    "theOrg",
				ServiceID:         sid,
			},
			Code:      202,
			ServiceID: sid,
			Operation: openservicebroker.OperationProvisioning,
		},
		{
			Name:         "it should fail with bad UID",
			ExpectError:  true,
			ProvisionReq: &openservicebroker.ProvisionRequest{},
			Code:         400,
			ServiceID:    "serviceGuid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var provisionRes = openservicebroker.ProvisionResponse{}
			req := brokerClient.Broker().RESTClient().Put().AbsPath("/broker/sdkbroker.broker.io/v2/service_instances/" + tc.ServiceID)
			req = req.SetHeader(openservicebroker.XBrokerAPIVersion, openservicebroker.APIVersion)
			req = req.SetHeader("Content-Type", "application/json")
			body, err := json.Marshal(tc.ProvisionReq)
			if err != nil {
				t.Fatalf("unexpected error marshalling ProvisionRequest %s", err)
			}
			req = req.Body(body)
			res := req.Do()

			if tc.ExpectError && res.Error() == nil {
				t.Fatal("expected an err from the provision endpoint but got none")
			}
			if !tc.ExpectError && res.Error() != nil {
				t.Fatalf("did not expect an error from the provision endpoint : %s body %s ", res.Error(), string(body))
			}
			status := new(int)
			res.StatusCode(status)
			if *status != tc.Code {
				t.Fatalf("expected %v status code but got %v ", tc.Code, *status)
			}
			if tc.ExpectError {
				return
			}
			data, err := res.Raw()
			if err != nil {
				t.Fatalf("failed to get raw response after provision call %s ", err)
			}
			if err := json.Unmarshal(data, &provisionRes); err != nil {
				t.Fatalf("failed to unmarshal into provision response %s", err)
			}
			if provisionRes.Operation != tc.Operation {
				t.Fatalf("expected the operation to be %s but got %s ", tc.Operation, provisionRes.Operation)
			}
		})
	}

}
