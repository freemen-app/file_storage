package grpcApi

import (
	"bytes"
	"context"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/rs/zerolog/log"
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
	var file bytes.Buffer

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

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
		}
		chunk := req.GetContent()
		log.Printf("received a chunk with size: %d", len(chunk))
		if _, err = file.Write(chunk); err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
		}
	}

	uploadInput := &dto.UploadInput{
		File:      bytes.NewReader(file.Bytes()),
		Directory: metadata.GetDirectory(),
		Filename:  metadata.GetFilename(),
		ACL:       "public-read",
	}
	if url, err := h.fileUseCase.Upload(stream.Context(), uploadInput); err != nil {
		return h.grpcPresenter.ConvertError(err).Err()
	} else if err := stream.SendAndClose(&fileStorage.UploadResponse{Url: url}); err != nil {
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
		log.Error().Msgf("[%T] %v", err, err)
		err = h.grpcPresenter.ConvertError(err).Err()
	}
	return resp, err

}
