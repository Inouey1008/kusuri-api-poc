package drug

import (
	"context"
	"database/sql"
)

type sqliteRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

// name_normalized の部分一致で最大 20 件を返す
// 0 件の場合は、空配列を返す
func (repository *sqliteRepository) FindByName(ctx context.Context, q string) ([]Drug, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		"SELECT yj_code, name FROM drug WHERE name_normalized LIKE '%' || ? || '%' LIMIT 20",
		q,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
