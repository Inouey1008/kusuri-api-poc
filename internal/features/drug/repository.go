package drug

import "context"

// Repository は永続化層の境界。将来 sqlc 生成コードや別DB実装に差し替えられる。
type Repository interface {
	// Search は name_normalized の部分一致で医薬品を検索する。件数上限は実装依存。
	Search(ctx context.Context, q string) ([]Drug, error)
}
