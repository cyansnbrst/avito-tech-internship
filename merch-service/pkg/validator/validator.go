package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	v *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	return &CustomValidator{v: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.v.Struct(i); err != nil {
		firstError := err.(validator.ValidationErrors)[0]

		errMessage := fmt.Errorf("field '%s' failed on the '%s' rule", firstError.Field(), firstError.Tag())

		return errMessage
	}

	return nil
}
