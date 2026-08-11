package drug

import "context"

// Service は医薬品に関するビジネスロジックを担う。
// 現状は Repository への委譲のみ。将来ここにクエリ正規化・ページング等のロジックが入る想定。
type Service struct {
	repo Repository
}

// NewService は Repository を受け取り Service を返す。
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Search は薬品名の部分一致検索を Repository に委譲する。
func (s *Service) Search(ctx context.Context, q string) ([]Drug, error) {
	return s.repo.Search(ctx, q)
}
