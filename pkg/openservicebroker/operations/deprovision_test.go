package operations

import (
	"errors"
	"net/http"
	"testing"

	"github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	clientset "github.com/openshift/open-service-broker-sdk/pkg/client/clientset_generated/internalclientset/fake"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestDeprovision(t *testing.T) {
	cases := []struct {
		Name        string
		InstanceID  string
		ExpectError bool
		DeleteError error
		Code        int
	}{
		{
			Name:        "deprovisions ok",
			InstanceID:  "test",
			ExpectError: false,
			Code:        http.StatusAccepted,
		},
		{
			Name:        "deprovisions fails with unexpected error",
			InstanceID:  "test",
			ExpectError: true,
			DeleteError: errors.New("unexpected error deprovisioning"),
			Code:        http.StatusInternalServerError,
		},
		{
			Name:        "deprovisions fails with not found error",
			InstanceID:  "test",
			ExpectError: false,
			DeleteError: kerrors.NewNotFound(broker.Resource("serviceInstance"), "test"),
			Code:        http.StatusGone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &clientset.Clientset{}
			client.Fake = ktesting.Fake{}
			client.Fake.AddReactor("delete", "*", func(a ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, tc.DeleteError
			})
			broker := &BrokerOperations{
				Client: client,
			}
			res := broker.Deprovision(tc.InstanceID)
			if tc.ExpectError && res.Err == nil {
				t.Fatal("Expected an err deprovisioning but got none")
			}
			if !tc.ExpectError && res.Err != nil {
				t.Fatalf("did not expect an error deprovisioning : %s ", res.Err)
			}
			if tc.Code != res.Code {
				t.Fatalf("expected a response code %v but got %v ", tc.Code, res.Code)
			}

		})
	}
}
