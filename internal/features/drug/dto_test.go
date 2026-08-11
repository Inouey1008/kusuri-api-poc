package drug

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

func TestDrug_ToResponse(t *testing.T) {
	testCases := []struct {
		name     string
		input    Drug
		expected drugResponse
	}{
		{
			name:     `通常の値`,
			input:    Drug{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
			expected: drugResponse{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
		},
		{
			name:     `ゼロ値`,
			input:    Drug{},
			expected: drugResponse{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, testCase.input.toResponse())
		})
	}
}

func TestSearchRequest_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		q             string
		expectedField string // 空文字なら成功 (nil) を期待
	}{
		{
			name: `空文字は許容 ("omitempty")`,
			q:    "",
		},
		{
			name: `通常の値`,
			q:    "エゼチミブ",
		},
		{
			name: `境界値: 100 文字 ("max=100")`,
			q:    strings.Repeat("a", 100),
		},
		{
			name: `境界値: 100 文字 (マルチバイト)`,
			q:    strings.Repeat("あ", 100),
		},
		{
			name:          `上限超過: 101 文字 ("max=100")`,
			q:             strings.Repeat("a", 101),
			expectedField: "q",
		},
		{
			name:          `上限超過: 101 文字 (マルチバイト)`,
			q:             strings.Repeat("あ", 101),
			expectedField: "q",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			errs := validation.Validate(searchRequest{Q: testCase.q})

			if testCase.expectedField == "" {
				assert.Nil(t, errs)
				return
			}

			if assert.NotEmpty(t, errs) {
				assert.Equal(t, testCase.expectedField, errs[0].Field)
			}
		})
	}
}
