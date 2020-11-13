package grpcApi

import (
	"context"
	"io"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	fileStorage "github.com/freemen-app/api/file_storage"

	grpcPresenter "github.com/freemen-app/file_storage/adapter/presenter/grpc"
	"github.com/freemen-app/file_storage/domain/dto"
	fileUseCase "github.com/freemen-app/file_storage/usecase/file"
)

type handler struct {
	fileUseCase   fileUseCase.UseCase
	grpcPresenter grpcPresenter.Presenter
}

func (h *handler) FileUseCase() fileUseCase.UseCase {
	return h.fileUseCase
}

func (h *handler) Presenter() grpcPresenter.Presenter {
	return h.grpcPresenter
}

func NewHandler(fileUseCase fileUseCase.UseCase) *handler {
	presenter := grpcPresenter.New()
	return &handler{
		fileUseCase:   fileUseCase,
		grpcPresenter: presenter,
	}
}

func (h *handler) Upload(stream fileStorage.FileStorage_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "cannot receive file info")
	}
	metadata := req.GetMetadata()
	log.Printf(
		"receive an upload request for filename [%s] with directory: %s",
		metadata.GetFilename(),
		metadata.GetDirectory(),
	)

	pr, pw := io.Pipe()
	go func(w *io.PipeWriter) {
		for {
			req, err := stream.Recv()
			if err != nil {
				log.Printf("Got error in pipewriter: %v", err)
				_ = w.CloseWithError(err)
				break
			}
			chunk := req.GetContent()
			log.Printf("received a chunk with size: %d", len(chunk))
			n, err := w.Write(chunk)
			if err != nil {
				_ = w.CloseWithError(err)
				break
			}
			log.Printf("wrote %d bytes to a pipe writer", n)
		}
	}(pw)

	uploadInput := &dto.UploadInput{
		File:      pr,
		Directory: metadata.GetDirectory(),
		Filename:  metadata.GetFilename(),
		ACL:       "public-read",
	}
	if url, err := h.fileUseCase.Upload(stream.Context(), uploadInput); err != nil {
		log.Printf("Got error from upload: %s", err.Error())
		return h.grpcPresenter.ConvertError(err).Err()
	} else if err := stream.SendAndClose(&fileStorage.UploadResponse{Url: url}); err != nil {
		log.Printf("Got error from send and close: %s", err.Error())
		return status.Errorf(codes.Unknown, "cannot send response: %v", err)
	}
	return nil
}

func (h *handler) Delete(ctx context.Context, request *fileStorage.DeleteRequest) (*empty.Empty, error) {
	err := h.fileUseCase.Delete(ctx, dto.DeleteInput(request.Url))
	return new(empty.Empty), err
}

func (h *handler) BatchDelete(ctx context.Context, request *fileStorage.BatchDeleteRequest) (*empty.Empty, error) {
	urls := make([]dto.DeleteInput, len(request.Urls))
	for i, url := range request.Urls {
		urls[i] = dto.DeleteInput(url)
	}
	err := h.fileUseCase.BatchDelete(ctx, urls)
	return new(empty.Empty), err
}

func (h *handler) ErrMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		// log.Error().Msgf("[%T] %v", err, err)
		err = h.grpcPresenter.ConvertError(err).Err()
	}
	return resp, err

}
