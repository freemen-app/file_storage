package customErrors

import validation "github.com/go-ozzo/ozzo-validation/v4"

var (
	InvalidURL = validation.NewError("400", "url: invalid format")
)
