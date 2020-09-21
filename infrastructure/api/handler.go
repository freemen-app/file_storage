package api

import (
	"context"
	"errors"

	"github.com/kataras/iris/v12"
	irisContext "github.com/kataras/iris/v12/context"

	"github.com/freemen-app/file_storage/domain/dto"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

type Handler struct {
	fileUseCase fileUseCase.UseCase
}

func NewHandler(fileUseCase fileUseCase.UseCase) *Handler {
	return &Handler{
		fileUseCase: fileUseCase,
	}
}

func (h *Handler) Ping(c iris.Context) {
	c.JSON(iris.Map{"pong": true})
}

func (h *Handler) UploadFile(irisCtx iris.Context) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	irisCtx.OnClose(func(c *irisContext.Context) { cancel() })

	file, headers, err := irisCtx.FormFile("file")
	if err != nil {
		return "", errors.New("file: cannot be blank")
	}
	input := &dto.UploadInput{
		File:      file,
		Filename:  headers.Filename,
		ACL:       irisCtx.FormValue("acl"),
		Directory: irisCtx.FormValue("directory"),
	}
	url, err := h.fileUseCase.Upload(ctx, input)
	return url, err
}
