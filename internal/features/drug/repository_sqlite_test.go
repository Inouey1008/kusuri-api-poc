package drug_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/test/testdb"
)

func TestSQLiteRepository_FindByName(t *testing.T) {
	repository := drug.NewRepository(testdb.Connect(t))

	testCases := []struct {
		name         string
		query        string
		expectedLen  int
		expectedCode string // 先頭件の yj_code。空文字なら検証しない
	}{
		{
			name:         `部分一致で複数件`,
			query:        "エゼチミブ",
			expectedLen:  2,
			expectedCode: "2189018F1043",
		},
		{
			name:         `部分一致で 1 件`,
			query:        "セレコキシブ",
			expectedLen:  1,
			expectedCode: "1149037F2093",
		},
		{
			name:         `正規化された名前で一致 (小文字)`,
			query:        "jg",
			expectedLen:  1,
			expectedCode: "2189018F1043",
		},
		{
			name:        `空クエリは全件`,
			query:       "",
			expectedLen: 3,
		},
		{
			name:        `一致なしは 0 件`,
			query:       "存在しない薬",
			expectedLen: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := repository.FindByName(context.Background(), testCase.query)

			require.NoError(t, err)
			assert.NotNil(t, got) // 0 件でも nil ではなく空配列
			require.Len(t, got, testCase.expectedLen)

			if testCase.expectedCode != "" {
				assert.Equal(t, testCase.expectedCode, got[0].YJCode)
			}
		})
	}
}
