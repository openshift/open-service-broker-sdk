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

// minimum supported client version
const minAPIVersionMajor, minAPIVersionMinor = 2, 7

func Route(container *restful.Container, path string, b Broker) {
	shim := func(f func(Broker, *restful.Request) *Response) func(*restful.Request, *restful.Response) {
		return func(req *restful.Request, resp *restful.Response) {
			fmt.Printf("got request %#v\n", *req)
			response := f(b, req)
			if response.Err != nil {
				resp.WriteHeaderAndJson(response.Code, &ErrorResponse{Description: response.Err.Error()}, restful.MIME_JSON)
			} else {
				resp.WriteHeaderAndJson(response.Code, response.Body, restful.MIME_JSON)
			}
		}
	}

	ws := restful.WebService{}
	ws.Path(path + "/v2")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Filter(apiVersion)

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

func apiVersion(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	versions := strings.SplitN(req.HeaderParameter(XBrokerAPIVersion), ".", 3)
	if len(versions) != 2 || atoi(versions[0]) != minAPIVersionMajor || atoi(versions[1]) < minAPIVersionMinor {
		resp.WriteHeaderAndJson(http.StatusPreconditionFailed, &ErrorResponse{Description: fmt.Sprintf("%s header must >= %d.%d", XBrokerAPIVersion, minAPIVersionMajor, minAPIVersionMinor)}, restful.MIME_JSON)
	}
	resp.AddHeader(XBrokerAPIVersion, APIVersion)
	chain.ProcessFilter(req, resp)
}

func catalog(b Broker, req *restful.Request) *Response {
	return b.Catalog()
}

func provision(b Broker, req *restful.Request) *Response {
	instance_id := req.PathParameter("instance_id")
	glog.Infof("processing provision for %s", instance_id)

	/*
		if errors := ValidateUUID(field.NewPath("instance_id"), instance_id); errors != nil {
			return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
		}
	*/
	var preq ProvisionRequest

	err := req.ReadEntity(&preq)
	if err != nil {
		return &Response{http.StatusBadRequest, nil, err}
	}

	if !preq.AcceptsIncomplete {
		return &Response{http.StatusUnprocessableEntity, AsyncRequired, nil}
	}

	return b.Provision(instance_id, &preq)
}

func deprovision(b Broker, req *restful.Request) *Response {
	if req.QueryParameter("accepts_incomplete") != "true" {
		return &Response{http.StatusUnprocessableEntity, &AsyncRequired, nil}
	}

	instance_id := req.PathParameter("instance_id")

	if errors := ValidateUUID(field.NewPath("instance_id"), instance_id); errors != nil {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	return b.Deprovision(instance_id)
}

func lastOperation(b Broker, req *restful.Request) *Response {
	instance_id := req.PathParameter("instance_id")
	if errors := ValidateUUID(field.NewPath("instance_id"), instance_id); errors != nil {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	operation := Operation(req.QueryParameter("operation"))
	if operation != OperationProvisioning &&
		operation != OperationUpdating &&
		operation != OperationDeprovisioning {
		return &Response{http.StatusBadRequest, nil, fmt.Errorf("invalid operation")}
	}

	return b.LastOperation(instance_id, operation)
}

func bind(b Broker, req *restful.Request) *Response {
	instance_id := req.PathParameter("instance_id")
	errors := ValidateUUID(field.NewPath("instance_id"), instance_id)

	binding_id := req.PathParameter("binding_id")
	errors = append(errors, ValidateUUID(field.NewPath("binding_id"), binding_id)...)

	if len(errors) > 0 {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	var breq BindRequest
	err := req.ReadEntity(&breq)
	if err != nil {
		return &Response{http.StatusBadRequest, nil, err}
	}

	return b.Bind(instance_id, binding_id, &breq)
}

func unbind(b Broker, req *restful.Request) *Response {
	instance_id := req.PathParameter("instance_id")
	errors := ValidateUUID(field.NewPath("instance_id"), instance_id)

	binding_id := req.PathParameter("binding_id")
	errors = append(errors, ValidateUUID(field.NewPath("binding_id"), binding_id)...)

	if len(errors) > 0 {
		return &Response{http.StatusBadRequest, nil, errors.ToAggregate()}
	}

	return b.Unbind(instance_id, binding_id)
}
