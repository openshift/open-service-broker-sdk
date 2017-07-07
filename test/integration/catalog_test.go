package integration

import (
	"net/http"
	"testing"

	"encoding/json"

	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
)

func TestCatalogEndpoint(t *testing.T) {
	stopCh, brokerClient, _, err := StartDefaultServer()
	if err != nil {
		t.Fatal(err)
	}
	defer close(stopCh)
	cases := []struct {
		Name          string
		Code          int
		ExpectError   bool
		ExpectedItems int
	}{
		{
			Name:          "catalog ok",
			Code:          http.StatusOK,
			ExpectedItems: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var catalogRes = openservicebroker.CatalogResponse{}
			req := brokerClient.Broker().RESTClient().Get().AbsPath("/broker/sdkbroker.broker.io/v2/catalog").SetHeader(openservicebroker.XBrokerAPIVersion, openservicebroker.APIVersion)
			res := req.Do()
			if tc.ExpectError && res.Error() == nil {
				t.Fatal("Expected an err from the catalog endpoint but got none")
			}
			if !tc.ExpectError && res.Error() != nil {
				t.Fatalf("did not expect an error from the catalog endpoint : %s ", res.Error())
			}
			status := new(int)
			res.StatusCode(status)
			if *status != tc.Code {
				t.Fatalf("expected %v status code but got %v ", tc.Code, *status)
			}
			data, err := res.Raw()
			if err != nil {
				t.Fatalf("failed to get raw response after catalog call %s ", err)
			}
			if err := json.Unmarshal(data, &catalogRes); err != nil {
				t.Fatalf("failed to unmarshal into catalog response %s", err)
			}
			if len(catalogRes.Services) != tc.ExpectedItems {
				t.Fatalf("expected %v items but got %v ", tc.ExpectedItems, len(catalogRes.Services))
			}
			assertHasService(t, "service-name", catalogRes.Services)
			for _, it := range catalogRes.Services {
				assertHasPlan(t, "gold-plan", it.Plans)
			}

		})
	}
}

func assertHasPlan(t *testing.T, expectedPlan string, plans []openservicebroker.Plan) {
	for _, p := range plans {
		if p.Name == expectedPlan {
			return
		}
	}
	t.Fatalf("expected to find plan %s but it was not present", expectedPlan)
}

func assertHasService(t *testing.T, expectedService string, services []*openservicebroker.Service) {
	for _, s := range services {
		if s.Name == expectedService {
			return
		}
	}
	t.Fatalf("expected to find service %s but it was not present", expectedService)
}
