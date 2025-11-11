package middleware

import (
	"log"
	"github.com/gofiber/fiber/v2"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Printf("Error: %v", err)

		// Map fiber error code to appropriate AppError
		var appErr *apperrors.AppError
		switch code {
		case fiber.StatusBadRequest:
			appErr = apperrors.ErrBadRequest.WithInternal(err)
		case fiber.StatusUnauthorized:
			appErr = apperrors.ErrUnauthorized.WithInternal(err)
		case fiber.StatusForbidden:
			appErr = apperrors.ErrForbidden.WithInternal(err)
		case fiber.StatusNotFound:
			appErr = apperrors.ErrNotFound.WithInternal(err)
		case fiber.StatusConflict:
			appErr = apperrors.ErrConflict.WithInternal(err)
		default:
			appErr = apperrors.ErrInternal.WithInternal(err)
		}

		return utils.ErrorResponse(c, appErr)
	}
}