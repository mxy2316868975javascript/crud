package httpapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/mxy2316868975javascript/crud/internal/domain"
)

// writeFailure 把领域错误映射成 HTTP 状态码和统一错误体。
func writeFailure(response http.ResponseWriter, err error) {
	var validation *domain.ValidationError
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(response, http.StatusNotFound, "not found")
	case errors.Is(err, domain.ErrWordExists):
		writeError(response, http.StatusConflict, "word already exists")
	case errors.Is(err, domain.ErrWordInUse):
		writeError(response, http.StatusConflict, "word is used by a test")
	case errors.Is(err, domain.ErrTestCompleted):
		writeError(response, http.StatusConflict, "conflict")
	case errors.Is(err, domain.ErrNotEnoughWords):
		writeError(response, http.StatusBadRequest, "not enough words in the bank")
	case errors.Is(err, domain.ErrInvalidAnswers):
		writeError(response, http.StatusBadRequest, "answers must cover every question exactly once")
	case errors.As(err, &validation):
		writeError(response, http.StatusBadRequest, validation.Message)
	default:
		log.Printf("server error: %v", err)
		writeError(response, http.StatusInternalServerError, "internal server error")
	}
}
