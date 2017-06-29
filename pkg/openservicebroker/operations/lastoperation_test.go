package operations

import (
	"errors"
	"net/http"
	"testing"

	"github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	clientset "github.com/openshift/open-service-broker-sdk/pkg/client/clientset_generated/internalclientset/fake"
	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kapi "k8s.io/client-go/pkg/api"
	ktesting "k8s.io/client-go/testing"
)

func TestLastoperation(t *testing.T) {
	cases := []struct {
		Name                string
		ExpectError         bool
		Operation           openservicebroker.Operation
		ServiceInstanceCond []broker.ServiceInstanceCondition
		Code                int
		GetError            error
		InstanceID          string
		ExpectedState       openservicebroker.LastOperationState
	}{
		{
			Name:        "lastoperation provision ok",
			ExpectError: false,
			Code:        http.StatusOK,
			Operation:   openservicebroker.OperationProvisioning,
			ServiceInstanceCond: []broker.ServiceInstanceCondition{{
				Type:   broker.ServiceInstanceReady,
				Status: kapi.ConditionTrue,
			}},
			GetError:      nil,
			InstanceID:    "instance",
			ExpectedState: openservicebroker.LastOperationStateSucceeded,
		},
		{
			Name:        "lastoperation provision failed",
			ExpectError: false,
			Code:        http.StatusOK,
			Operation:   openservicebroker.OperationProvisioning,
			ServiceInstanceCond: []broker.ServiceInstanceCondition{{
				Type:   broker.ServiceInstanceFailed,
				Status: kapi.ConditionTrue,
			}},
			GetError:      nil,
			InstanceID:    "instance",
			ExpectedState: openservicebroker.LastOperationStateFailed,
		},
		{
			Name:        "lastoperation provision in progress",
			ExpectError: false,
			Code:        http.StatusOK,
			Operation:   openservicebroker.OperationProvisioning,
			ServiceInstanceCond: []broker.ServiceInstanceCondition{{
				Status: kapi.ConditionTrue,
			}},
			GetError:      nil,
			InstanceID:    "instance",
			ExpectedState: openservicebroker.LastOperationStateInProgress,
		},
		{
			Name:                "lastoperation error for provisioning when service instance not found",
			ExpectError:         true,
			Code:                http.StatusBadRequest,
			Operation:           openservicebroker.OperationProvisioning,
			ServiceInstanceCond: []broker.ServiceInstanceCondition{},
			GetError:            kerrors.NewNotFound(broker.Resource("serviceInstance"), "test"),
			InstanceID:          "instance",
		},
		{
			Name:                "lastoperation service gone for deprovisioning",
			ExpectError:         false,
			Code:                http.StatusGone,
			Operation:           openservicebroker.OperationDeprovisioning,
			ServiceInstanceCond: []broker.ServiceInstanceCondition{},
			GetError:            kerrors.NewNotFound(broker.Resource("serviceInstance"), "test"),
			InstanceID:          "instance",
		},
		{
			Name:                "lastoperation internal service error when general error finding service instance",
			ExpectError:         true,
			Code:                http.StatusInternalServerError,
			Operation:           openservicebroker.OperationProvisioning,
			ServiceInstanceCond: []broker.ServiceInstanceCondition{},
			GetError:            errors.New("something terribly wrong"),
			InstanceID:          "instance",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			client := &clientset.Clientset{}
			client.Fake = ktesting.Fake{}
			client.Fake.AddReactor("get", "*", func(a ktesting.Action) (bool, runtime.Object, error) {
				serviceInstance := &broker.ServiceInstance{
					Status: broker.ServiceInstanceStatus{
						Conditions: tc.ServiceInstanceCond,
					},
				}
				return true, serviceInstance, tc.GetError
			})
			broker := &BrokerOperations{
				Client: client,
			}
			res := broker.LastOperation(tc.InstanceID, tc.Operation)
			if tc.ExpectError && res.Err == nil {
				t.Fatal("Expected an err from lastoperation but got none")
			}
			if !tc.ExpectError && res.Err != nil {
				t.Fatalf("did not expect an error from lastoperation : %s ", res.Err)
			}
			if tc.Code != res.Code {
				t.Fatalf("expected a response code %v but got %v ", tc.Code, res.Code)
			}
			if res.Code == http.StatusOK {
				loRes, ok := res.Body.(*openservicebroker.LastOperationResponse)
				if !ok {
					t.Fatal("expected the response body to be a LastOperationResponse")
				}
				if loRes.State != tc.ExpectedState {
					t.Fatalf("expected lastoperation state to be %v but got %v ", tc.ExpectedState, loRes.State)
				}
			}
		})
	}
}
