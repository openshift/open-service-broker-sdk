package openservicebroker

import (
	"regexp"

	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var parameterNameExp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ValidateUUID(path *field.Path, u string) field.ErrorList {
	if uuid.Parse(u) == nil {
		return field.ErrorList{field.Invalid(path, u, "must be a valid UUID")}
	}
	return nil
}
