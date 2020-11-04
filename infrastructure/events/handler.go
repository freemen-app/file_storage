package events

import (
	"context"
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"

	"github.com/freemen-app/file_storage/domain/dto"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

type (
	handler struct {
		fileUseCase fileUseCase.UseCase
	}
)

func (h *handler) DeleteFiles(delivery amqp.Delivery) {
	ctx := context.Background()
	var input dto.BatchDeleteInput
	if err := json.Unmarshal(delivery.Body, &input); logError(err) != nil {
		logError(delivery.Reject(false))
		return
	}

	err := h.fileUseCase.BatchDelete(ctx, input)
	logError(err)

	switch err.(type) {
	case nil:
		logError(delivery.Ack(true))

	case validation.Error:
	case validation.Errors:
		logError(delivery.Reject(false))
	default:
		logError(delivery.Reject(true))
	}
}

func logError(err error) error {
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
