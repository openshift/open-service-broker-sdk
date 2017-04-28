package openservicebroker

import (
	"github.com/golang/glog"

	"fmt"
	"net/http"
	"strconv"
	"strings"

	restful "github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// minimum supported client version for the service broker api version
const minAPIVersionMajor, minAPIVersionMinor = 2, 7

// Route sets up the service broker api endpoints.
func Route(container *restful.Container, path string, b Broker) {

	// shim does some basic boiler plate handling of all requests before handing the request
	// off to the api specific function.
	shim := func(f func(Broker, *restful.Request) *Response) func(*restful.Request, *restful.Response) {
		return func(req *restful.Request, resp *restful.Response) {
			response := f(b, req)
			if response.Err != nil {
				resp.WriteHeaderAndJson(response.Code, &ErrorResponse{Description: response.Err.Error()}, restful.MIME_JSON)
			} else {
				resp.WriteHeaderAndJson(response.Code, response.Body, restful.MIME_JSON)
			}
		}
	}

	ws := restful.WebService{}
	// v2 is a required part of the service broker api request path.
	ws.Path(path + "/v2")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(apiVersion)

	// register the various service broker api paths
	ws.Route(ws.GET("/catalog").To(shim(catalog)))
	ws.Route(ws.PUT("/service_instances/{instance_id}").To(shim(provision)))
	ws.Route(ws.DELETE("/service_instances/{instance_id}").To(shim(deprovision)))
	ws.Route(ws.GET("/service_instances/{instance_id}/last_operation").To(shim(lastOperation)))
	ws.Route(ws.PUT("/service_instances/{instance_id}/service_bindings/{binding_id}").To(shim(bind)))
	ws.Route(ws.DELETE("/service_instances/{instance_id}/service_bindings/{binding_id}").To(shim(unbind)))
	container.Add(&ws)
}

func atoi(s string) int {
	rv, err := strconv.Atoi(s)
	if err != nil {
		rv = 0
	}
	return rv
}

// apiVersion ensures that the request is using a supported service broker api version.
func apiVersion(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	versions := strings.SplitN(req.HeaderParameter(XBrokerAPIVersion), ".", 3)
	if len(versions) != 2 || atoi(versions[0]) != minAPIVersionMajor || atoi(versions[1]) < minAPIVersionMinor {
		resp.WriteHeaderAndJson(http.StatusPreconditionFailed, &ErrorResponse{Description: fmt.Sprintf("%s header must >= %d.%d", XBrokerAPIVersion, minAPIVersionMajor, minAPIVersionMinor)}, restful.MIME_JSON)
	}
	resp.AddHeader(XBrokerAPIVersion, APIVersion)
	chain.ProcessFilter(req, resp)
}

// catalog hands the request off to the BrokerOperations catalog implementation
func catalog(b Broker, req *restful.Request) *Response {
	return b.Catalog()
}

// provision hands the request off to the BrokerOperations provision implementation
func provision(b Broker, req *restful.Request) *Response {
	// grab the instance id of the provision request
	instanceID := req.PathParameter("instance_id")

	glog.Infof("processing provision request for %s", instanceID)

	// make sure it's a valid uuid
	if errors := ValidateUUID(field.NewPath("instance_id"), instanceID); errors != nil {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	var preq ProvisionRequest
	err := req.ReadEntity(&preq)
	if err != nil {
		return &Response{http.StatusBadRequest, nil, err}
	}

	// this broker performs asynchronous provisioning, so the client must
	// indicate that it will accept incomplete(async) provision responses.
	if !preq.AcceptsIncomplete {
		return &Response{http.StatusUnprocessableEntity, AsyncRequired, nil}
	}

	return b.Provision(instanceID, &preq)
}

// deprovision hands the request off to the BrokerOperations deprovision implementation
func deprovision(b Broker, req *restful.Request) *Response {
	// this broker performs asynchronous deprovisioning, so the client must
	// indicate it supports an async response.
	if req.QueryParameter("accepts_incomplete") != "true" {
		return &Response{http.StatusUnprocessableEntity, &AsyncRequired, nil}
	}

	// grab the service instance id we are deprovisioning
	instanceID := req.PathParameter("instance_id")

	glog.Infof("processing deprovision request for %s", instanceID)

	// make sure it's a valid uuid
	if errors := ValidateUUID(field.NewPath("instance_id"), instanceID); errors != nil {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	return b.Deprovision(instanceID)
}

// lastOperation hands the request off to the BrokerOperations lastoperation implementation
func lastOperation(b Broker, req *restful.Request) *Response {

	// get the service instance id who's state is being requested
	instanceID := req.PathParameter("instance_id")

	glog.Infof("processing lastoperation request for %s", instanceID)

	if errors := ValidateUUID(field.NewPath("instance_id"), instanceID); errors != nil {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	operation := Operation(req.QueryParameter("operation"))
	if operation != OperationProvisioning &&
		operation != OperationUpdating &&
		operation != OperationDeprovisioning {
		return &Response{http.StatusBadRequest, nil, fmt.Errorf("invalid operation")}
	}

	return b.LastOperation(instanceID, operation)
}

// bind hands the request off to the BrokerOperations bind implementation
func bind(b Broker, req *restful.Request) *Response {

	// get the service instance id we're binding to
	instanceID := req.PathParameter("instance_id")

	glog.Infof("processing bind request for %s", instanceID)

	errors := ValidateUUID(field.NewPath("instance_id"), instanceID)

	// get the id for the binding that will be created
	bindingID := req.PathParameter("binding_id")
	glog.Infof("with binding id %s", bindingID)
	errors = append(errors, ValidateUUID(field.NewPath("binding_id"), bindingID)...)

	if len(errors) > 0 {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	var breq BindRequest
	err := req.ReadEntity(&breq)
	if err != nil {
		return &Response{http.StatusBadRequest, nil, err}
	}

	return b.Bind(instanceID, bindingID, &breq)
}

// unbind hands the request off to the BrokerOperations unbind implementation
func unbind(b Broker, req *restful.Request) *Response {

	// get the service instance id we are unbinding from
	instanceID := req.PathParameter("instance_id")

	glog.Infof("processing unbind request for %s", instanceID)

	errors := ValidateUUID(field.NewPath("instance_id"), instanceID)

	// get the id of the binding we are removing
	bindingID := req.PathParameter("binding_id")
	glog.Infof("with binding id %s", bindingID)
	errors = append(errors, ValidateUUID(field.NewPath("binding_id"), bindingID)...)

	if len(errors) > 0 {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	return b.Unbind(instanceID, bindingID)
}
