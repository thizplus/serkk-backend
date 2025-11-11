package utils

import (
	"github.com/gofiber/fiber/v2"
	apperrors "gofiber-template/pkg/errors"
)

type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type PaginatedResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    Meta        `json:"meta"`
	Error   string      `json:"error,omitempty"`
}

type Meta struct {
	Total  int64 `json:"total"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

// SuccessResponse returns success response
func SuccessResponse(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created returns 201 response
func Created(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent returns 204 response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// ErrorResponse returns error response with AppError support
func ErrorResponse(c *fiber.Ctx, err error) error {
	// Check if it's AppError
	if appErr, ok := apperrors.IsAppError(err); ok {
		return c.Status(appErr.StatusCode).JSON(Response{
			Success: false,
			Error: &ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
				Fields:  appErr.Fields,
			},
		})
	}

	// Unknown error - return generic error
	return c.Status(fiber.StatusInternalServerError).JSON(Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "An unexpected error occurred",
		},
	})
}

// PaginatedSuccessResponse returns paginated success response
func PaginatedSuccessResponse(c *fiber.Ctx, message string, data interface{}, total int64, offset, limit int) error {
	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta: Meta{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	})
}

// ========== Backward Compatibility Functions (Deprecated) ==========
// These functions are kept for backward compatibility
// New code should use ErrorResponse with AppError instead

// ValidationErrorResponse returns validation error (deprecated)
// Use apperrors.ErrValidation instead
func ValidationErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, apperrors.ErrValidation.WithMessage(message))
}

// UnauthorizedResponse returns unauthorized error (deprecated)
// Use apperrors.ErrUnauthorized instead
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, apperrors.ErrUnauthorized.WithMessage(message))
}

// NotFoundResponse returns not found error (deprecated)
// Use apperrors.ErrNotFound instead
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, apperrors.ErrNotFound.WithMessage(message))
}

// InternalServerErrorResponse returns internal error (deprecated)
// Use apperrors.ErrInternal instead
func InternalServerErrorResponse(c *fiber.Ctx, message string, err error) error {
	return ErrorResponse(c, apperrors.ErrInternal.WithMessage(message).WithInternal(err))
}