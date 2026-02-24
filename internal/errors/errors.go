package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrUnauthorized        = NewAppError("UNAUTHORIZED", "unauthorized: missing auth token", nil)
	ErrAccountNotFound     = NewAppError("ACCOUNT_NOT_FOUND", "account not found", nil)
	ErrAPIKeyNotFound      = NewAppError("API_KEY_NOT_FOUND", "API key not found", nil)
	ErrInvalidRequest      = NewAppError("INVALID_REQUEST", "email or mobile required", nil)
	ErrServiceNotAvailable = NewAppError("SERVICE_UNAVAILABLE", "service not available", nil)
	ErrAccountExists       = NewAppError("ACCOUNT_EXISTS", "email already exists", nil)
	ErrMobileExists        = NewAppError("MOBILE_EXISTS", "mobile already exists", nil)
)

func WriteErrorResponse(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	detail := "Internal server error"

	if err != nil {
		var appErr *AppError
		if stderrors.As(err, &appErr) {
			status = statusCodeForError(appErr.Code)
			detail = appErr.Message
		} else {
			slog.Error("[errors] internal error", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{"detail": detail}); encodeErr != nil {
		slog.Error("[errors] failed to encode error response", "error", encodeErr)
	}
}

func statusCodeForError(code string) int {
	switch code {
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "ACCOUNT_NOT_FOUND", "API_KEY_NOT_FOUND":
		return http.StatusNotFound
	case "INVALID_REQUEST":
		return http.StatusBadRequest
	case "API_KEY_EXPIRED":
		return http.StatusUnauthorized
	case "API_KEY_EXPIRING":
		return http.StatusConflict
	case "SERVICE_UNAVAILABLE":
		return http.StatusServiceUnavailable
	case "ACCOUNT_EXISTS", "MOBILE_EXISTS":
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
