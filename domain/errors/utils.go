package customErrors

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func BadRequestDetails(errors *validation.Errors) *errdetails.BadRequest {
	if errors == nil {
		return nil
	}
	details := &errdetails.BadRequest{}
	for key, err := range *errors {
		detail := &errdetails.BadRequest_FieldViolation{
			Field:       key,
			Description: err.Error(),
		}
		details.FieldViolations = append(details.FieldViolations, detail)
	}
	return details
}
