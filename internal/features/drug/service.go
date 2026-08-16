package drug

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Search(ctx context.Context, q string) ([]Drug, error) {
	return service.repository.FindByName(ctx, q)
}
