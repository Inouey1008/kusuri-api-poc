package drug

import "context"

type Repository interface {
	FindByName(ctx context.Context, q string) ([]Drug, error)
}
