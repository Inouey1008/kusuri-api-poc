// Package validation はアプリ全体で再利用できる入力バリデーション機能を提供する。
// handler は FieldError と Validate のみを使い、validator の詳細はこのパッケージに隠蔽する。
package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate はパッケージ起動時に一度だけ生成する (キャッシュ効率・コールドスタート配慮)。
var validate = validator.New()

func init() {
	// json タグの先頭要素をフィールド名として使う (例: `json:"yjCode,omitempty"` → "yjCode")。
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// FieldError はバリデーション違反の詳細を表す。
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate は構造体 s をバリデートし、違反があれば []FieldError を返す。
// 成功時は nil を返す。
func Validate(s any) []FieldError {
	if err := validate.Struct(s); err != nil {
		if errors, ok := err.(validator.ValidationErrors); ok {
			errs := make([]FieldError, 0, len(errors))
			for _, fe := range errors {
				errs = append(errs, FieldError{
					Field:   fe.Field(),
					Message: buildMessage(fe),
				})
			}
			return errs
		}
		// 予期しないエラー型はそのまま返す
		return []FieldError{{Field: "", Message: err.Error()}}
	}
	return nil
}

// buildMessage はバリデーションタグに応じた人間可読なメッセージを返す。
func buildMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "alphanum":
		return "must be alphanumeric"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
