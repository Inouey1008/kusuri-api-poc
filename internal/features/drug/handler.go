package drug

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

type service interface {
	Search(ctx context.Context, q string) ([]Drug, error)
}

type Handler struct {
	service service
}

func NewHandler(service service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) Search(c echo.Context) error {
	q := c.QueryParam("q")

	if errs := validation.Validate(searchRequest{Q: q}); errs != nil {
		return errorx.Validation.WithDetails(errs)
	}

	items, err := handler.service.Search(c.Request().Context(), q)
	if err != nil {
		return err
	}

	responses := make([]drugResponse, len(items))
	for i, item := range items {
		responses[i] = item.toResponse()
	}

	return c.JSON(http.StatusOK, searchResponse{
		Total: len(responses),
		Items: responses,
	})
}
