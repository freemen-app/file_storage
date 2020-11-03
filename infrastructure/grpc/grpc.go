package grpcApi

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/app"
	pb "github.com/freemen-app/file_storage/infrastructure/proto"
)

type api struct {
	listener net.Listener
	handler  *handler
	grpc     *grpc.Server
}

func (g api) Grpc() *grpc.Server {
	return g.grpc
}

func (g api) Handler() *handler {
	return g.handler
}

func (g api) Listener() net.Listener {
	return g.listener
}

func New(app *app.App, config *config.ApiConfig) *api {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	handler := NewHandler(app.UseCases().FileUseCase)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(handler.ErrMiddleware))
	pb.RegisterFileStorageServer(grpcServer, handler)

	return &api{
		listener: listener,
		handler:  handler,
		grpc:     grpcServer,
	}
}

func (g *api) Start() {
	log.Info().Msgf("Started grpc server on %s", g.listener.Addr())
	if err := g.grpc.Serve(g.listener); err != nil {
		panic(err)
	}
}

func (g *api) Shutdown() {
	g.grpc.Stop()
	log.Info().Msg("Server stopped")
}
