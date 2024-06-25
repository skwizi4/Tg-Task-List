package Config

import "github.com/go-playground/validator/v10"

func (config Config) ValidateConfig(validator *validator.Validate) error {
	return validator.Struct(config)
}
