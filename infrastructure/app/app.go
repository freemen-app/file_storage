package app

import (
	fileRepo "github.com/freemen-app/file_storage/adapter/repository/file"
	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/log"
	awsSession "github.com/freemen-app/file_storage/infrastructure/store/aws"

	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

type (
	repos struct {
		File fileUseCase.FileRepo
	}

	useCases struct {
		FileUseCase fileUseCase.UseCase
	}

	App struct {
		config *config.Config

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
	repos := &repos{File: fileRepo.New(session, config.S3.Bucket)}
	useCases := &useCases{FileUseCase: fileUseCase.New(repos.File)}

	return &App{
		config:   config,
		repos:    repos,
		useCases: useCases,
	}
}

func (a *App) UseCases() *useCases {
	return a.useCases
}

func (a *App) Repos() *repos {
	return a.repos
}
