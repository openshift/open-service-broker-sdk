package openservicebroker

import (
	"github.com/pborman/uuid"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateUUID checks if a provided value is a valid UUID
func ValidateUUID(path *field.Path, u string) field.ErrorList {
	if uuid.Parse(u) == nil {
		return field.ErrorList{field.Invalid(path, u, "must be a valid UUID")}
	}
	return nil
}
