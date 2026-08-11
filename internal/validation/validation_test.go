package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name            string
		input           any
		expectedField   string // 空文字なら成功 (nil) を期待
		expectedMessage string // 空文字ならメッセージは検証しない
	}{
		{
			name: `成功`,
			input: struct {
				Name string `json:"name" validate:"required,max=10"`
			}{Name: "hello"},
		},
		{
			name: `必須項目が空 ("required")`,
			input: struct {
				Name string `json:"name" validate:"required"`
			}{Name: ""},
			expectedField:   "name",
			expectedMessage: "is required",
		},
		{
			name: `上限文字数を超過 ("max")`,
			input: struct {
				Q string `json:"q" validate:"omitempty,max=5"`
			}{Q: "toolong"},
			expectedField: "q",
		},
		{
			name: `文字数が一致しない ("len")`,
			input: struct {
				Code string `json:"code" validate:"required,len=12"`
			}{Code: "short"},
			expectedField:   "code",
			expectedMessage: "must be exactly 12 characters",
		},
		{
			name: `英数字以外の文字を含む ("alphanum")`,
			input: struct {
				Code string `json:"code" validate:"required,alphanum"`
			}{Code: "abc!def"},
			expectedField:   "code",
			expectedMessage: "must be alphanumeric",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			errs := validation.Validate(testCase.input)

			if testCase.expectedField == "" {
				assert.Nil(t, errs)
				return
			}

			if assert.NotEmpty(t, errs) {
				assert.Equal(t, testCase.expectedField, errs[0].Field)
				if testCase.expectedMessage != "" {
					assert.Equal(t, testCase.expectedMessage, errs[0].Message)
				}
			}
		})
	}
}
