package api

import (
	"context"
	"time"

	"github.com/kataras/iris/v12"
	irisContext "github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/rs/zerolog/log"

	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/app"
)

type (
	Api interface {
		Start()
		Shutdown(ctx context.Context)
	}

	api struct {
		config *config.Config
		iris   *iris.Application
		app    *app.App
	}
)

func AccessLogHandler(ctx *irisContext.Context, _ time.Duration) {
	status := ctx.GetStatusCode()
	logger := log.Info()

	if irisContext.StatusCodeNotSuccessful(status) {
		logger = log.Error()
	}

	logger.
		Str("path", ctx.Path()).
		Str("method", ctx.Method()).
		Int("status", ctx.GetStatusCode())
}

func New(config *config.Config, app *app.App) Api {
	i := iris.New()

	loggerConfig := logger.Config{
		Status:     true,
		Method:     true,
		Path:       true,
		Query:      true,
		LogFuncCtx: AccessLogHandler,
	}
	i.Use(logger.New(loggerConfig))
	setupMiddlewares(i)

	handler := NewHandler(app.FileUseCase)
	handler.Register(i)

	return &api{
		config: config,
		iris:   i,
		app:    app,
	}
}

func (a *api) Start() {
	err := a.iris.Listen(
		a.config.Api.Addr(),
		iris.WithoutServerError(iris.ErrServerClosed),
	)
	if err != nil {
		log.Error().Err(err)
	}
}

func (a *api) Shutdown(ctx context.Context) {
	if err := a.iris.Shutdown(ctx); err != nil {
		log.Error().Err(err)
	}

	a.app.Shutdown()
}

func setupMiddlewares(e *iris.Application) {
	// e.DoneGlobal(middleware.ErrorMiddlewacre)
}
