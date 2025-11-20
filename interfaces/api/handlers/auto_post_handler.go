package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
)

type AutoPostHandler struct {
	service   services.AutoPostService
	validator *validator.Validate
}

func NewAutoPostHandler(service services.AutoPostService) *AutoPostHandler {
	return &AutoPostHandler{
		service:   service,
		validator: validator.New(),
	}
}

// CreateSetting godoc
// @Summary Create auto-post setting
// @Description Create a new auto-post setting for AI-powered automatic posting
// @Tags auto-post
// @Accept json
// @Produce json
// @Param request body dto.CreateAutoPostSettingRequest true "Create auto-post setting request"
// @Success 201 {object} dto.AutoPostSettingResponse
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings [post]
func (h *AutoPostHandler) CreateSetting(c *fiber.Ctx) error {
	var req dto.CreateAutoPostSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Invalid request body"))
	}

	if err := h.validator.Struct(&req); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrValidation.WithMessage(err.Error()))
	}

	setting, err := h.service.CreateSetting(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(setting)
}

// GetSetting godoc
// @Summary Get auto-post setting
// @Description Get auto-post setting by ID
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} dto.AutoPostSettingResponse
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings/{id} [get]
func (h *AutoPostHandler) GetSetting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	setting, err := h.service.GetSetting(c.Context(), id)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrNotFound.WithMessage(err.Error()))
	}

	return c.JSON(setting)
}

// UpdateSetting godoc
// @Summary Update auto-post setting
// @Description Update auto-post setting by ID
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Param request body dto.UpdateAutoPostSettingRequest true "Update auto-post setting request"
// @Success 200 {object} dto.AutoPostSettingResponse
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings/{id} [put]
func (h *AutoPostHandler) UpdateSetting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	var req dto.UpdateAutoPostSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	if err := h.validator.Struct(&req); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrValidation.WithMessage(err.Error()))
	}

	setting, err := h.service.UpdateSetting(c.Context(), id, &req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.JSON(setting)
}

// DeleteSetting godoc
// @Summary Delete auto-post setting
// @Description Delete auto-post setting by ID
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Success 204
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings/{id} [delete]
func (h *AutoPostHandler) DeleteSetting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	if err := h.service.DeleteSetting(c.Context(), id); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListSettings godoc
// @Summary List auto-post settings
// @Description List all auto-post settings
// @Tags auto-post
// @Accept json
// @Produce json
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings [get]
func (h *AutoPostHandler) ListSettings(c *fiber.Ctx) error {
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 20)

	settings, total, err := h.service.ListSettings(c.Context(), offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.JSON(fiber.Map{
		"settings": settings,
		"meta": fiber.Map{
			"offset":     offset,
			"limit":      limit,
			"totalCount": total,
		},
	})
}

// EnableSetting godoc
// @Summary Enable auto-post setting
// @Description Enable auto-post setting by ID
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings/{id}/enable [post]
func (h *AutoPostHandler) EnableSetting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	if err := h.service.EnableSetting(c.Context(), id); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.JSON(fiber.Map{"message": "Setting enabled successfully"})
}

// DisableSetting godoc
// @Summary Disable auto-post setting
// @Description Disable auto-post setting by ID
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings/{id}/disable [post]
func (h *AutoPostHandler) DisableSetting(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	if err := h.service.DisableSetting(c.Context(), id); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.JSON(fiber.Map{"message": "Setting disabled successfully"})
}

// TriggerAutoPost godoc
// @Summary Trigger auto-post manually
// @Description Manually trigger an auto-post generation for a specific setting
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Setting ID"
// @Param request body dto.TriggerAutoPostRequest false "Trigger auto-post request"
// @Success 200 {object} dto.TriggerAutoPostResponse
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/settings/{id}/trigger [post]
func (h *AutoPostHandler) TriggerAutoPost(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	var req dto.TriggerAutoPostRequest
	if err := c.BodyParser(&req); err != nil {
		// If body parsing fails, just use nil topic (will randomly select)
		req.Topic = nil
	}

	response, err := h.service.GenerateAndPost(c.Context(), id, req.Topic)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.JSON(response)
}

// ListLogs godoc
// @Summary List auto-post logs
// @Description List auto-post generation logs
// @Tags auto-post
// @Accept json
// @Produce json
// @Param settingId query string false "Filter by Setting ID"
// @Param status query string false "Filter by status (pending, success, failed)"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.AutoPostLogListResponse
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/logs [get]
func (h *AutoPostHandler) ListLogs(c *fiber.Ctx) error {
	req := dto.ListAutoPostLogsRequest{
		Offset: c.QueryInt("offset", 0),
		Limit:  c.QueryInt("limit", 20),
	}

	if settingID := c.Query("settingId"); settingID != "" {
		id, err := uuid.Parse(settingID)
		if err != nil {
			return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
		}
		req.SettingID = &id
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	response, err := h.service.ListLogs(c.Context(), &req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage(err.Error()))
	}

	return c.JSON(response)
}

// GetLog godoc
// @Summary Get auto-post log
// @Description Get auto-post log by ID
// @Tags auto-post
// @Accept json
// @Produce json
// @Param id path string true "Log ID"
// @Success 200 {object} dto.AutoPostLogResponse
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /auto-post/logs/{id} [get]
func (h *AutoPostHandler) GetLog(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage(""))
	}

	log, err := h.service.GetLog(c.Context(), id)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrNotFound.WithMessage(err.Error()))
	}

	return c.JSON(log)
}
