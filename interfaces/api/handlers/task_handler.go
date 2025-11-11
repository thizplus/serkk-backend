package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
	"strconv"
)

type TaskHandler struct {
	taskService services.TaskService
}

func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "User not authenticated")
	}

	var req dto.CreateTaskRequest
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

	task, err := h.taskService.CreateTask(c.Context(), user.ID, &req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Task creation failed").WithInternal(err))
	}

	taskResponse := dto.TaskToTaskResponse(task, nil)
	return utils.SuccessResponse(c, taskResponse, "Task created successfully")
}

func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid task ID")
	}

	task, err := h.taskService.GetTask(c.Context(), taskID)
	if err != nil {
		return utils.NotFoundResponse(c, "Task not found")
	}

	taskResponse := dto.TaskToTaskResponse(task, &task.User)
	return utils.SuccessResponse(c, taskResponse, "Task retrieved successfully")
}

func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid task ID")
	}

	var req dto.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ValidationErrorResponse(c, "Invalid request body")
	}

	task, err := h.taskService.UpdateTask(c.Context(), taskID, &req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Task update failed").WithInternal(err))
	}

	taskResponse := dto.TaskToTaskResponse(task, &task.User)
	return utils.SuccessResponse(c, taskResponse, "Task updated successfully")
}

func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	taskIDStr := c.Params("id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid task ID")
	}

	err = h.taskService.DeleteTask(c.Context(), taskID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Task deletion failed").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Task deleted successfully")
}

func (h *TaskHandler) GetUserTasks(c *fiber.Ctx) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.UnauthorizedResponse(c, "User not authenticated")
	}

	offsetStr := c.Query("offset", "0")
	limitStr := c.Query("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid offset parameter")
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid limit parameter")
	}

	tasks, total, err := h.taskService.GetUserTasks(c.Context(), user.ID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve tasks").WithInternal(err))
	}

	taskResponses := make([]dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = *dto.TaskToTaskResponse(task, &task.User)
	}

	response := &dto.TaskListResponse{
		Tasks: taskResponses,
		Meta: dto.PaginationMeta{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return utils.SuccessResponse(c, response, "Tasks retrieved successfully")
}

func (h *TaskHandler) ListTasks(c *fiber.Ctx) error {
	offsetStr := c.Query("offset", "0")
	limitStr := c.Query("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid offset parameter")
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid limit parameter")
	}

	tasks, total, err := h.taskService.ListTasks(c.Context(), offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve tasks").WithInternal(err))
	}

	taskResponses := make([]dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = *dto.TaskToTaskResponse(task, &task.User)
	}

	response := &dto.TaskListResponse{
		Tasks: taskResponses,
		Meta: dto.PaginationMeta{
			Total:  total,
			Offset: offset,
			Limit:  limit,
		},
	}

	return utils.SuccessResponse(c, response, "Tasks retrieved successfully")
}
