package validatorhelper

import "github.com/go-playground/validator/v10"

type Validator struct {
	v *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{v: validator.New()}
}

func (v *Validator) Validate(models ...interface{}) error {
	for _, m := range models {
		err := v.v.Struct(m)
		if err != nil {
			return err
		}
	}
	return nil
}
