package apierror

import "fmt"

// the single error type that travels from handlers to the global error handler
// it carries enough context to produce a structured JSON response without any ad-hoc mapping
type AppError struct {
	HTTPStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.HTTPStatus, e.Code, e.Message)
}

func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
	}
}

func NewBadRequest(code, message string) *AppError {
	return New(400, code, message)
}

func NewUnauthorized(code, message string) *AppError {
	return New(401, code, message)
}

func NewForbidden(code, message string) *AppError {
	return New(403, code, message)
}

func NewNotFound(code, message string) *AppError {
	return New(404, code, message)
}

func NewConflict(code, message string) *AppError {
	return New(409, code, message)
}

func NewUnprocessable(code, message string) *AppError {
	return New(422, code, message)
}
