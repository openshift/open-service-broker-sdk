package operations

import (
	"net/http"
	"testing"

	clientset "github.com/openshift/open-service-broker-sdk/pkg/client/clientset_generated/internalclientset/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestUnbind(t *testing.T) {
	cases := []struct {
		Name        string
		ExpectError bool
		Code        int
		instanceID  string
		bindingID   string
	}{
		{
			Name:        "unbinds ok",
			ExpectError: false,
			Code:        http.StatusOK,
			instanceID:  "instance",
			bindingID:   "binding",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &clientset.Clientset{}
			client.Fake = ktesting.Fake{}
			broker := &BrokerOperations{
				Client: client,
			}
			res := broker.Unbind(tc.instanceID, tc.bindingID)
			if tc.ExpectError && res.Err == nil {
				t.Fatal("Expected an err unbinding but got none")
			}
			if !tc.ExpectError && res.Err != nil {
				t.Fatalf("did not expect an error unbinding : %s ", res.Err)
			}
			if tc.Code != res.Code {
				t.Fatalf("expected a response code %v but got %v ", tc.Code, res.Code)
			}
		})
	}
}
