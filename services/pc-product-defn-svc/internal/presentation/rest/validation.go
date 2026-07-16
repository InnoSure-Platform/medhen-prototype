package rest

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ProblemDetails implements RFC 7807 Problem Details for HTTP APIs
type ProblemDetails struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	
	// ValidationErrors is an extension field for bad requests
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	
	// Register custom tag name extractor to use the JSON struct tag
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct and returns formatted ValidationErrors
func ValidateStruct(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(), // This will now use the JSON tag due to our RegisterTagNameFunc
			Message: getErrorMessage(err),
		})
	}
	return errors
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Must be at least " + err.Param()
	case "max":
		return "Must be at most " + err.Param()
	default:
		return "Invalid value"
	}
}

// WriteProblem responds with a standardized Problem Details JSON
func WriteProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	
	prob := ProblemDetails{
		Status: status,
		Title:  title,
		Detail: detail,
	}
	json.NewEncoder(w).Encode(prob)
}

// WriteValidationProblem responds with a 400 Bad Request and validation details
func WriteValidationProblem(w http.ResponseWriter, errors []ValidationError) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusBadRequest)
	
	prob := ProblemDetails{
		Status:           http.StatusBadRequest,
		Title:            "Validation Error",
		Detail:           "One or more validation errors occurred.",
		ValidationErrors: errors,
	}
	json.NewEncoder(w).Encode(prob)
}
