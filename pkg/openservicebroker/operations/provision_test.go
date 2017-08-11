package operations

import (
	"testing"

	"github.com/openshift/open-service-broker-sdk/pkg/apis/broker"
	clientset "github.com/openshift/open-service-broker-sdk/pkg/client/clientset_generated/internalclientset/fake"
	"github.com/openshift/open-service-broker-sdk/pkg/openservicebroker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestProvision(t *testing.T) {
	cases := []struct {
		Name        string
		InstanceID  string
		ExpectError bool
		Request     *openservicebroker.ProvisionRequest
		Code        int
	}{
		{
			Name:        "test async provision success",
			InstanceID:  "testID",
			ExpectError: false,
			Code:        202,
			Request: &openservicebroker.ProvisionRequest{
				Context: openservicebroker.KubernetesContext{
					Platform:  "kubernetes",
					Namespace: "test",
				},
				ServiceID:         "test",
				PlanID:            "test",
				Parameters:        map[string]string{},
				AcceptsIncomplete: true,
				OrganizationID:    "org",
				SpaceID:           "space",
			},
		},
		{
			Name:        "test sync provision fails with unprocessable entity",
			InstanceID:  "testID",
			ExpectError: true,
			Code:        422,
			Request: &openservicebroker.ProvisionRequest{
				Context: openservicebroker.KubernetesContext{
					Platform:  "kubernetes",
					Namespace: "test",
				},
				ServiceID:         "test",
				PlanID:            "test",
				Parameters:        map[string]string{},
				AcceptsIncomplete: false,
				OrganizationID:    "org",
				SpaceID:           "space",
			},
		},
	}

	for _, tc := range cases {
		si := &broker.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{Namespace: broker.Namespace},
		}

		broker := &BrokerOperations{
			Client: clientset.NewSimpleClientset(si),
		}
		t.Run(tc.Name, func(t *testing.T) {
			res := broker.Provision(tc.InstanceID, tc.Request)
			if tc.ExpectError && res.Err == nil {
				t.Fatal("Expected an err provisioning but got none")
			}
			if !tc.ExpectError && res.Err != nil {
				t.Fatalf("did not expect an error provisioning : %s ", res.Err)
			}
			if res.Code != tc.Code {
				t.Fatalf("Expected a status code %v but got %v", tc.Code, res.Code)
			}
		})
	}
}
