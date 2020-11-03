package grpcApi

import fileUseCase "github.com/freemen-app/file_storage/usecase/file"

type Handler = handler
type API = api

func (h *handler) SetFileUseCase(useCase fileUseCase.UseCase) {
	h.fileUseCase = useCase
}
