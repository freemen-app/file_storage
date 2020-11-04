package events

import (
	amqpStore "github.com/freemen-app/amqp-store"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"

	"github.com/freemen-app/file_storage/infrastructure/app"
)

type (
	consumes struct {
		DeleteFiles *amqpStore.ConsumeConfig `mapstructure:"delete_files"`
	}

	consumer struct {
		store    amqpStore.Store
		handler  *handler
		consumes *consumes
	}
)

func (c *consumes) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.DeleteFiles, validation.Required),
	)
}

func New(app *app.App, conf *amqpStore.Config) *consumer {
	consumes := new(consumes)
	if err := mapstructure.Decode(conf.Consumes, consumes); err != nil {
		panic(err)
	} else if err := consumes.Validate(); err != nil {
		panic(err)
	}

	handler := &handler{fileUseCase: app.UseCases().FileUseCase}
	return &consumer{
		store:    app.Stores().AMQP,
		handler:  handler,
		consumes: consumes,
	}
}

func (c *consumer) Start() error {
	confHandlerMapping := map[*amqpStore.ConsumeConfig]func(delivery amqp.Delivery){
		c.consumes.DeleteFiles: c.handler.DeleteFiles,
	}
	for conf, handler := range confHandlerMapping {
		if err := c.store.Subscribe(conf, handler); err != nil {
			return err
		}
	}
	log.Info().Msg("Started AMQP server")
	return nil
}
