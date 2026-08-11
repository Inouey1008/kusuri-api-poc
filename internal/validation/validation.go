package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	// json タグの先頭要素をフィールド名として使う
	// (例: `json:"firstName,omitempty"` → "firstName")
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// 構造体 s を validate タグに基づいて検証する。
// 違反があれば []FieldError、なければ nil を返す。
func Validate(s any) []FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 予期しないエラー型はそのまま返す
		return []FieldError{{Field: "", Message: err.Error()}}
	}

	errs := make([]FieldError, 0, len(errors))
	for _, e := range errors {
		errs = append(errs, FieldError{
			Field:   e.Field(),
			Message: buildMessage(e),
		})
	}
	return errs
}

func buildMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s characters", fieldError.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", fieldError.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fieldError.Param())
	case "alphanum":
		return "must be alphanumeric"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fieldError.Param(), " ", ", "))
	default:
		return fmt.Sprintf("failed validation: %s", fieldError.Tag())
	}
}
