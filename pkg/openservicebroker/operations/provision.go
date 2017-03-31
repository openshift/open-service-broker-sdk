package operations

import (
	"net/http"

	"github.com/golang/glog"

	brokerapi "github.com/openshift/brokersdk/pkg/apis/broker/v1alpha1"
	"github.com/openshift/brokersdk/pkg/openservicebroker"
)

/*
func (b *BrokerOperations) ensureUUIDMap(instance_id string, tp *api.TemplateProvision, didWork *bool) (*api.UUIDMap, *openservicebroker.Response) {
	uuidMap := &api.UUIDMap{
		ObjectMeta: kapi.ObjectMeta{Name: instance_id},
		Spec: api.UUIDMapSpec{
			TemplateProvisionRef: kapi.ObjectReference{
				Kind:      tp.Kind,
				Namespace: tp.Namespace,
				Name:      tp.Name,
			},
		},
	}

	newUUIDMap, err := b.oc.UUIDMaps().Create(uuidMap)
	if err == nil {
		*didWork = true
		return newUUIDMap, nil
	}

	if errors.IsAlreadyExists(err) {
		var existingUM *api.UUIDMap
		existingUM, err = b.oc.UUIDMaps().Get(uuidMap.Name)
		if err == nil && reflect.DeepEqual(uuidMap.Spec, existingUM.Spec) {
			return existingUM, nil
		}

		return nil, &openservicebroker.Response{http.StatusConflict, openservicebroker.ProvisionResponse{}, nil}
	}

	return nil, &openservicebroker.Response{http.StatusInternalServerError, nil, err}
}
*/

func (b *BrokerOperations) Provision(instance_id string, preq *openservicebroker.ProvisionRequest) *openservicebroker.Response {
	/*
		kubeconfig, err := clientcmd.BuildConfigFromFlags("https://127.0.0.1:443", "")
		if err != nil {
			glog.Errorf("Failed to create a kube config\n:%v\n", err)

		}

		kubeconfig.Insecure = true
		brokerClient, err := clientset.NewForConfig(kubeconfig)

		if err != nil {
			glog.Errorf("Failed to create a broker client\n:%v\n", err)

		}
	*/
	si := brokerapi.ServiceInstance{}
	si.Name = instance_id
	si.Spec.Credential = "some_secret"
	si.Status.Conditions = append(si.Status.Conditions, brokerapi.ServiceInstanceCondition{})

	_, err := b.Client.ServiceInstances("brokersdk").Create(&si)
	if err != nil {
		glog.Errorf("Failed to create a service instance\n:%v\n", err)
	}
	/*
		siList, err := b.Client.ServiceInstances("brokersdk").List(metav1.ListOptions{})
		if err != nil {
			glog.Errorf("Failed to list instances\n:%v\n", err)

		}
		glog.Infof("got siList: %#v", siList)
	*/

	/* Use this for async provision flows
	if didWork {
		return &openservicebroker.Response{http.StatusAccepted, openservicebroker.ProvisionResponse{Operation: openservicebroker.OperationProvisioning}, nil}
	}
	*/
	return &openservicebroker.Response{http.StatusOK, openservicebroker.ProvisionResponse{Operation: openservicebroker.OperationProvisioning}, nil}
}
