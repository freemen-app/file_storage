package grpcPresenter

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	customErrors "github.com/freemen-app/file_storage/domain/errors"
)

type (
	userPresenter struct{}

	Presenter interface {
		ConvertError(err error) *status.Status
	}
)

func New() Presenter {
	return new(userPresenter)
}

func (p *userPresenter) ConvertError(err error) *status.Status {
	// Check if err is nil or value
	switch err {
	case nil:
		return nil
	case context.DeadlineExceeded:
		return status.New(codes.DeadlineExceeded, err.Error())
	case context.Canceled:
		return status.New(codes.Canceled, err.Error())
	}

	// Check if err has defined type
	var grpcErr *status.Status

	switch errObj := err.(type) {
	case validation.Errors:
		grpcErr = status.New(codes.InvalidArgument, errObj.Error())
		grpcErr, _ = grpcErr.WithDetails(customErrors.BadRequestDetails(&errObj))
	default:
		grpcErr = status.New(codes.Internal, err.Error())
	}
	return grpcErr
}
