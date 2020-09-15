package api

import (
	"github.com/kataras/iris/v12"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Ping(c iris.Context) {
	c.JSON(iris.Map{"pong": true})
}

func (h *Handler) UploadFile(c iris.Context) {
	// err := h.
}
