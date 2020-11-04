package app

import (
	"reflect"
	"time"

	amqpStore "github.com/freemen-app/amqp-store"

	fileRepo "github.com/freemen-app/file_storage/adapter/repository/file"
	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/log"
	awsSession "github.com/freemen-app/file_storage/infrastructure/store/aws"

	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

type (
	launchedStore interface {
		Start() error
		IsRunning() bool
		Shutdown()
	}

	stores struct {
		AMQP amqpStore.Store
	}

	repos struct {
		File fileUseCase.FileRepo
	}

	useCases struct {
		FileUseCase fileUseCase.UseCase
	}

	App struct {
		config *config.Config

		stores   *stores
		repos    *repos
		useCases *useCases

		isRunning bool
	}
)

func New(config *config.Config) *App {
	if err := config.Validate(); err != nil {
		panic(err)
	}
	log.ConfigureLogger(config.Logger.Level)
	session := awsSession.New(config.S3)
	stores := &stores{
		AMQP: amqpStore.New(config.AMQP.DSN(), time.Second),
	}
	repos := &repos{File: fileRepo.New(session, config.S3.Bucket)}
	useCases := &useCases{FileUseCase: fileUseCase.New(repos.File)}

	return &App{
		config:   config,
		stores:   stores,
		repos:    repos,
		useCases: useCases,
	}
}

func (a *App) Stores() *stores {
	return a.stores
}

func (a *App) UseCases() *useCases {
	return a.useCases
}

func (a *App) Repos() *repos {
	return a.repos
}

func (a *App) IsRunning() bool {
	return a.isRunning
}

func (a *App) Start() error {
	// Iterate through all attributes of store and
	// launch stores that implements launchedStore interface
	stores := reflect.ValueOf(*a.stores)
	for i := 0; i < stores.NumField(); i++ {
		if stores.Field(i).IsNil() {
			continue
		}
		attribute := stores.Field(i).Interface()
		if store, ok := attribute.(launchedStore); ok && !store.IsRunning() {
			if err := store.Start(); err != nil {
				return err
			}
		}
	}

	a.isRunning = true
	return nil
}

func (a *App) Shutdown() {
	// Iterate through all attributes of store and
	// stop stores that implements launchedStore interface
	stores := reflect.ValueOf(*a.stores)
	for i := 0; i < stores.NumField(); i++ {
		if stores.Field(i).IsNil() {
			continue
		}
		attribute := stores.Field(i).Interface()
		if store, ok := attribute.(launchedStore); ok && store.IsRunning() {
			store.Shutdown()
		}
	}
	a.isRunning = false
}
