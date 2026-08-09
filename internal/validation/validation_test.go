package validation_test

import (
	"testing"

	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

func TestValidate_Success(t *testing.T) {
	type input struct {
		Name string `json:"name" validate:"required,max=10"`
	}

	errs := validation.Validate(input{Name: "hello"})
	if errs != nil {
		t.Fatalf("want nil, got %v", errs)
	}
}

func TestValidate_Required(t *testing.T) {
	type input struct {
		Name string `json:"name" validate:"required"`
	}

	errs := validation.Validate(input{Name: ""})
	if len(errs) == 0 {
		t.Fatal("want errors, got nil")
	}
	if errs[0].Field != "name" {
		t.Errorf("field = %q, want \"name\"", errs[0].Field)
	}
	if errs[0].Message != "is required" {
		t.Errorf("message = %q, want \"is required\"", errs[0].Message)
	}
}

func TestValidate_Max(t *testing.T) {
	type input struct {
		Q string `json:"q" validate:"omitempty,max=5"`
	}

	errs := validation.Validate(input{Q: "toolong"})
	if len(errs) == 0 {
		t.Fatal("want errors, got nil")
	}
	if errs[0].Field != "q" {
		t.Errorf("field = %q, want \"q\"", errs[0].Field)
	}
}

func TestValidate_Len(t *testing.T) {
	type input struct {
		Code string `json:"code" validate:"required,len=12"`
	}

	errs := validation.Validate(input{Code: "short"})
	if len(errs) == 0 {
		t.Fatal("want errors, got nil")
	}
	if errs[0].Field != "code" {
		t.Errorf("field = %q, want \"code\"", errs[0].Field)
	}
	if errs[0].Message != "must be exactly 12 characters" {
		t.Errorf("message = %q, want \"must be exactly 12 characters\"", errs[0].Message)
	}
}

func TestValidate_Alphanum(t *testing.T) {
	type input struct {
		Code string `json:"code" validate:"required,alphanum"`
	}

	errs := validation.Validate(input{Code: "abc!def"})
	if len(errs) == 0 {
		t.Fatal("want errors, got nil")
	}
	if errs[0].Field != "code" {
		t.Errorf("field = %q, want \"code\"", errs[0].Field)
	}
	if errs[0].Message != "must be alphanumeric" {
		t.Errorf("message = %q, want \"must be alphanumeric\"", errs[0].Message)
	}
}
