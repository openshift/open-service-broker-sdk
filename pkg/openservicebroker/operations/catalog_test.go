package operations

import (
	"testing"

	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
)

func TestCatalog(t *testing.T) {
	cases := []struct {
		Name         string
		ExpectError  bool
		Code         int
		CatalogItems int
	}{
		{
			Name:         "test get catalog",
			ExpectError:  false,
			Code:         200,
			CatalogItems: 1,
		},
	}

	broker := &BrokerOperations{}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			res := broker.Catalog()
			if tc.ExpectError && res.Err == nil {
				t.Fatal("expected an err getting the catalog but got none")
			}
			if !tc.ExpectError && res.Err != nil {
				t.Fatalf("did not expect an error getting the catalog : %s ", res.Err)
			}
			if nil != res.Body {
				catRes, ok := res.Body.(*openservicebroker.CatalogResponse)
				if !ok {
					t.Fatalf("casting res.Body to a slice of services failed")
				}
				if len(catRes.Services) != tc.CatalogItems {
					t.Fatalf("expected %d catalog items but instead got %d ", tc.CatalogItems, len(catRes.Services))
				}
			}
			if res.Code != tc.Code {
				t.Fatalf("expected response code %d but got %d ", tc.Code, res.Code)
			}
		})
	}
}
