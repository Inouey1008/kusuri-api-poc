package drug

import (
	"context"
	"net/http"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/httpx"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

type service interface {
	Search(ctx context.Context, q string) ([]Drug, error)
	GetByYJCode(ctx context.Context, yjCode string) (*Drug, error)
}

type Handler struct {
	service service
}

func NewHandler(service service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Search(writer http.ResponseWriter, request *http.Request) {
	q := request.URL.Query().Get("q")

	if errs := validation.Validate(searchRequest{Q: q}); errs != nil {
		errorx.Validation.WithDetails(errs).Write(writer)
		return
	}

	items, err := handler.service.Search(request.Context(), q)
	if err != nil {
		errorx.Internal.Write(writer)
		return
	}

	responses := make([]drugResponse, len(items))
	for i, item := range items {
		responses[i] = item.toResponse()
	}

	httpx.WriteJSON(writer, http.StatusOK, searchResponse{
		Total: len(responses),
		Items: responses,
	})
}

func (handler *Handler) GetByYJCode(writer http.ResponseWriter, request *http.Request) {
	code := request.PathValue("yjCode")

	if errs := validation.Validate(getRequest{YJCode: code}); errs != nil {
		errorx.Validation.WithDetails(errs).Write(writer)
		return
	}

	item, err := handler.service.GetByYJCode(request.Context(), code)
	if err != nil {
		errorx.Internal.Write(writer)
		return
	}
	if item == nil {
		errorx.NotFound.Write(writer)
		return
	}

	httpx.WriteJSON(writer, http.StatusOK, item.toResponse())
}
