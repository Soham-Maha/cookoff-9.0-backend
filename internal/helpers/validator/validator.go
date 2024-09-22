package validator

import (
	"net/http"

	httphelpers "github.com/CodeChefVIT/cookoff-backend/internal/helpers/http"
	logger "github.com/CodeChefVIT/cookoff-backend/internal/helpers/logging"
	"github.com/go-playground/validator"
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}

func ValidatePayload(w http.ResponseWriter, v any) error {
	if err := Validate.Struct(v); err != nil {
		logger.Errof("Please provide values for all required fields.")
		httphelpers.WriteError(w, http.StatusNotAcceptable, err.Error())
		return err
	}
	httphelpers.WriteJSON(w, http.StatusOK, map[string]string{"message": "Payload Validated"})
	return nil
}