package models

import (
	"fmt"
	"net/http"
)

// ErrorType representa o tipo de erro
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation_error"
	ErrorTypeAuthentication ErrorType = "authentication_error"
	ErrorTypeAuthorization  ErrorType = "authorization_error"
	ErrorTypeNotFound       ErrorType = "not_found_error"
	ErrorTypeConflict       ErrorType = "conflict_error"
	ErrorTypeInternal       ErrorType = "internal_error"
	ErrorTypeBadRequest     ErrorType = "bad_request_error"
)

// APIError representa um erro padronizado da API
type APIError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
}

// Error implementa a interface error
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// ErrorResponse representa a resposta de erro padronizada
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// NewValidationError cria um novo erro de validação
func NewValidationError(message, details string) *APIError {
	return &APIError{
		Type:       ErrorTypeValidation,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewAuthenticationError cria um novo erro de autenticação
func NewAuthenticationError(message string) *APIError {
	return &APIError{
		Type:       ErrorTypeAuthentication,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewAuthorizationError cria um novo erro de autorização
func NewAuthorizationError(message string) *APIError {
	return &APIError{
		Type:       ErrorTypeAuthorization,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewNotFoundError cria um novo erro de recurso não encontrado
func NewNotFoundError(message string) *APIError {
	return &APIError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewConflictError cria um novo erro de conflito
func NewConflictError(message string) *APIError {
	return &APIError{
		Type:       ErrorTypeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewInternalError cria um novo erro interno
func NewInternalError(message string) *APIError {
	return &APIError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// NewBadRequestError cria um novo erro de requisição inválida
func NewBadRequestError(message, details string) *APIError {
	return &APIError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}