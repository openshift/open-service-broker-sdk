package operations

import (
	"errors"
	"net/http"
	"testing"

	brokerapi "github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	clientset "github.com/openshift/open-service-broker-sdk/pkg/client/clientset_generated/internalclientset/fake"
	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestBind(t *testing.T) {
	cases := []struct {
		Name        string
		ExpectError bool
		Code        int
		GetError    error
		InstanceID  string
		BindingID   string
		BindingReq  *openservicebroker.BindRequest
	}{
		{
			Name:        "binds ok",
			ExpectError: false,
			Code:        http.StatusCreated,
			GetError:    nil,
			InstanceID:  "instid",
			BindingID:   "bindingid",
			BindingReq:  &openservicebroker.BindRequest{},
		},
		{
			Name:        "binds fails on error",
			ExpectError: true,
			Code:        http.StatusInternalServerError,
			GetError:    errors.New("something is terribly wrong"),
			InstanceID:  "instid",
			BindingID:   "bindingid",
			BindingReq:  &openservicebroker.BindRequest{},
		},
		{
			Name:        "binds fails on service instance not found",
			ExpectError: false,
			Code:        http.StatusGone,
			GetError:    kerrors.NewNotFound(brokerapi.Resource("serviceInstance"), "test"),
			InstanceID:  "instid",
			BindingID:   "bindingid",
			BindingReq:  &openservicebroker.BindRequest{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &clientset.Clientset{}
			client.Fake = ktesting.Fake{}
			client.Fake.AddReactor("get", "*", func(a ktesting.Action) (bool, runtime.Object, error) {
				serviceInstance := &brokerapi.ServiceInstance{
					Spec: brokerapi.ServiceInstanceSpec{
						Credential: "supersecret",
					},
				}
				return true, serviceInstance, tc.GetError
			})
			broker := &BrokerOperations{
				Client: client,
			}
			res := broker.Bind(tc.InstanceID, tc.BindingID, tc.BindingReq)
			if tc.ExpectError && res.Err == nil {
				t.Fatal("Expected an err binding but got none")
			}
			if !tc.ExpectError && res.Err != nil {
				t.Fatalf("did not expect an error binding : %s ", res.Err)
			}
			if tc.Code != res.Code {
				t.Fatalf("expected a response code %v but got %v ", tc.Code, res.Code)
			}
			if res.Code == http.StatusCreated {
				bindRes, ok := res.Body.(*openservicebroker.BindResponse)
				if !ok {
					t.Fatal("expected the response body to be a BindResponse")
				}
				if _, ok := bindRes.Credentials["credential"]; !ok {
					t.Fatal("expected to find a credential key in the BindResponse")
				}
			}
		})
	}
}
