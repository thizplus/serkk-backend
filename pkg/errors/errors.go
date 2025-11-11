package errors

import (
	"fmt"
	"net/http"
)

// AppError represents application-specific errors
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	StatusCode int               `json:"-"`
	Internal   error             `json:"-"`
	Fields     map[string]string `json:"fields,omitempty"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// WithMessage adds a custom message
func (e *AppError) WithMessage(msg string) *AppError {
	newErr := *e
	newErr.Message = msg
	return &newErr
}

// WithInternal adds internal error
func (e *AppError) WithInternal(err error) *AppError {
	newErr := *e
	newErr.Internal = err
	return &newErr
}

// WithField adds a field error
func (e *AppError) WithField(key, value string) *AppError {
	newErr := *e
	if newErr.Fields == nil {
		newErr.Fields = make(map[string]string)
	}
	newErr.Fields[key] = value
	return &newErr
}

// Predefined errors - 400 Bad Request
var (
	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "Invalid request",
		StatusCode: http.StatusBadRequest,
	}

	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
	}
)

// 401 Unauthorized
var (
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid email or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		StatusCode: http.StatusUnauthorized,
	}
)

// 403 Forbidden
var (
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "Access denied",
		StatusCode: http.StatusForbidden,
	}
)

// 404 Not Found
var (
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "User not found",
		StatusCode: http.StatusNotFound,
	}

	ErrPostNotFound = &AppError{
		Code:       "POST_NOT_FOUND",
		Message:    "Post not found",
		StatusCode: http.StatusNotFound,
	}

	ErrCommentNotFound = &AppError{
		Code:       "COMMENT_NOT_FOUND",
		Message:    "Comment not found",
		StatusCode: http.StatusNotFound,
	}

	ErrConversationNotFound = &AppError{
		Code:       "CONVERSATION_NOT_FOUND",
		Message:    "Conversation not found",
		StatusCode: http.StatusNotFound,
	}

	ErrMessageNotFound = &AppError{
		Code:       "MESSAGE_NOT_FOUND",
		Message:    "Message not found",
		StatusCode: http.StatusNotFound,
	}
)

// 409 Conflict
var (
	ErrConflict = &AppError{
		Code:       "CONFLICT",
		Message:    "Resource already exists",
		StatusCode: http.StatusConflict,
	}

	ErrEmailExists = &AppError{
		Code:       "EMAIL_EXISTS",
		Message:    "Email already registered",
		StatusCode: http.StatusConflict,
	}

	ErrUsernameExists = &AppError{
		Code:       "USERNAME_EXISTS",
		Message:    "Username already taken",
		StatusCode: http.StatusConflict,
	}

	ErrAlreadyVoted = &AppError{
		Code:       "ALREADY_VOTED",
		Message:    "You have already voted on this item",
		StatusCode: http.StatusConflict,
	}

	ErrAlreadyFollowing = &AppError{
		Code:       "ALREADY_FOLLOWING",
		Message:    "You are already following this user",
		StatusCode: http.StatusConflict,
	}
)

// 500 Internal Server Error
var (
	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrDatabase = &AppError{
		Code:       "DATABASE_ERROR",
		Message:    "Database operation failed",
		StatusCode: http.StatusInternalServerError,
	}

	ErrFileUpload = &AppError{
		Code:       "FILE_UPLOAD_ERROR",
		Message:    "File upload failed",
		StatusCode: http.StatusInternalServerError,
	}

	ErrFileProcess = &AppError{
		Code:       "FILE_PROCESS_ERROR",
		Message:    "File processing failed",
		StatusCode: http.StatusInternalServerError,
	}

	ErrStorageError = &AppError{
		Code:       "STORAGE_ERROR",
		Message:    "Storage operation failed",
		StatusCode: http.StatusInternalServerError,
	}
)

// IsAppError checks if error is AppError
func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
