package validator

import (
	"net/http"

	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/go-playground/validator"
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}

func ValidatePayload(w http.ResponseWriter, v any) error {
	if err := Validate.Struct(v); err != nil {
		logger.Warnf("%s", err.Error())
		return err
	}
	return nil
}
