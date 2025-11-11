package handlers

import (
		apperrors "gofiber-template/pkg/errors"
"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/utils"
)

type NotificationHandler struct {
	notificationService services.NotificationService
}

func NewNotificationHandler(notificationService services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications retrieves all notifications for the user
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	notifications, err := h.notificationService.GetNotifications(c.Context(), userID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve notifications").WithInternal(err))
	}

	return utils.SuccessResponse(c, notifications, "Notifications retrieved successfully")
}

// GetUnreadNotifications retrieves unread notifications
func (h *NotificationHandler) GetUnreadNotifications(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	notifications, err := h.notificationService.GetUnreadNotifications(c.Context(), userID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve unread notifications").WithInternal(err))
	}

	return utils.SuccessResponse(c, notifications, "Unread notifications retrieved successfully")
}

// GetNotification retrieves a single notification
func (h *NotificationHandler) GetNotification(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid notification ID")
	}

	notification, err := h.notificationService.GetNotification(c.Context(), notificationID, userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrNotFound.WithMessage("Notification not found").WithInternal(err))
	}

	return utils.SuccessResponse(c, notification, "Notification retrieved successfully")
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid notification ID")
	}

	err = h.notificationService.MarkAsRead(c.Context(), notificationID, userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to mark notification as read").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Notification marked as read")
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	err := h.notificationService.MarkAllAsRead(c.Context(), userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to mark all notifications as read").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "All notifications marked as read")
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	notificationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid notification ID")
	}

	err = h.notificationService.DeleteNotification(c.Context(), notificationID, userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to delete notification").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Notification deleted successfully")
}

// DeleteAllNotifications deletes all notifications
func (h *NotificationHandler) DeleteAllNotifications(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	err := h.notificationService.DeleteAllNotifications(c.Context(), userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to delete all notifications").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "All notifications deleted successfully")
}

// GetUnreadCount retrieves unread notification count
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	count, err := h.notificationService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve unread count").WithInternal(err))
	}

	return utils.SuccessResponse(c, fiber.Map{
		"count": count,
	}, "Unread count retrieved successfully")
}

// GetSettings retrieves notification settings
func (h *NotificationHandler) GetSettings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	settings, err := h.notificationService.GetSettings(c.Context(), userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve notification settings").WithInternal(err))
	}

	return utils.SuccessResponse(c, settings, "Notification settings retrieved successfully")
}

// UpdateSettings updates notification settings
func (h *NotificationHandler) UpdateSettings(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var req dto.NotificationSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		errors := utils.GetValidationErrors(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	settings, err := h.notificationService.UpdateSettings(c.Context(), userID, &req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to update notification settings").WithInternal(err))
	}

	return utils.SuccessResponse(c, settings, "Notification settings updated successfully")
}
