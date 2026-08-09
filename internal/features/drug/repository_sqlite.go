package drug

import (
	"context"
	"database/sql"
	"errors"
)

// sqliteRepository は Repository の SQLite 実装。
// ドライバの登録と接続の生成は platform/sqlitedb が担い、ここはクエリだけを持つ。
type sqliteRepository struct {
	db *sql.DB
}

// NewRepository は *sql.DB を受け取り Repository を返す。
// 実装を差し替えるときは、このファイルごと入れ替える。
func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

// Search は name_normalized の部分一致で最大 20 件を返す。0件は空スライス (nil でない)。
func (r *sqliteRepository) Search(ctx context.Context, q string) ([]Drug, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT yj_code, name FROM drug WHERE name_normalized LIKE '%' || ? || '%' LIMIT 20",
		q,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Drug{}
	for rows.Next() {
		var d Drug
		if err := rows.Scan(&d.YJCode, &d.Name); err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	return results, rows.Err()
}

// FindByYJCode は YJ コードで1件を返す。該当なしは (nil, nil)。
func (r *sqliteRepository) FindByYJCode(ctx context.Context, yjCode string) (*Drug, error) {
	var d Drug
	err := r.db.QueryRowContext(
		ctx,
		"SELECT yj_code, name FROM drug WHERE yj_code = ?",
		yjCode,
	).Scan(&d.YJCode, &d.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}
