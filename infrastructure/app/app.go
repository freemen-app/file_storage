package app

import (
	fileRepo "github.com/freemen-app/file_storage/adapter/repository/file"
	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/log"
	awsSession "github.com/freemen-app/file_storage/infrastructure/store/aws"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

type (
	shutdowner interface {
		Shutdown()
	}

	App struct {
		config       *config.Config
		cleanupTasks []shutdowner
		FileUseCase  fileUseCase.UseCase
	}
)

func New(config *config.Config) *App {
	log.ConfigureLogger(config.Logger.Level)

	app := &App{config: config}
	defer app.shutdownOnPanic()

	session := awsSession.New()
	fileRepo := fileRepo.New(session, config.S3.Bucket)
	app.FileUseCase = fileUseCase.New(fileRepo)

	return app
}

func (a *App) AddCleanupTask(s shutdowner) {
	a.cleanupTasks = append(a.cleanupTasks, s)
}

func (a *App) Shutdown() {
	lastIndex := len(a.cleanupTasks) - 1

	for i := range a.cleanupTasks {
		a.cleanupTasks[lastIndex-i].Shutdown()
	}
}

func (a *App) shutdownOnPanic() {
	if r := recover(); r != nil {
		a.Shutdown()
		panic(r)
	}
}
