package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Validate(v interface{}) error {
	if v == nil {
		return fmt.Errorf("cannot validate nil")
	}
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var msgs []string
	for _, fe := range ve {
		field := fe.StructField()
		ns := fe.Namespace()
		tag := fe.Tag()
		param := fe.Param()

		switch tag {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", field))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s", ns, param))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s", ns, param))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of: %s", field, param))
		default:
			msgs = append(msgs, fmt.Sprintf("%s failed validation: %s", field, tag))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
}
