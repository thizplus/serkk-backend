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

type SearchHandler struct {
	searchService services.SearchService
}

func NewSearchHandler(searchService services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search performs a search for posts (posts only, with cursor pagination)
// Supports both cursor-based (recommended) and offset-based (deprecated) pagination
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.ValidationErrorResponse(c, "Search query is required")
	}

	cursor := c.Query("cursor", "")
	limit := normalizeLimit(c.Query("limit", "20"))

	// Get userID if authenticated (optional)
	var userIDPtr *uuid.UUID
	if userID, ok := c.Locals("userID").(uuid.UUID); ok {
		userIDPtr = &userID
	}

	// Check if using cursor-based pagination (recommended)
	if cursor != "" || c.Query("offset") == "" {
		// Use cursor-based pagination (new way) - Posts only
		results, err := h.searchService.SearchWithCursor(c.Context(), userIDPtr, query, cursor, limit)
		if err != nil {
			return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Search failed").WithInternal(err))
		}
		return utils.SuccessResponse(c, results, "Search completed successfully")
	}

	// Fallback to offset-based pagination (deprecated)
	searchType := c.Query("type", "post") // Default to "post" only
	req := &dto.SearchRequest{
		Query: query,
		Type:  searchType,
		Limit: limit,
	}

	results, err := h.searchService.Search(c.Context(), userIDPtr, req)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Search failed").WithInternal(err))
	}

	return utils.SuccessResponse(c, results, "Search completed successfully (offset-based deprecated)")
}

// GetSearchHistory retrieves user's search history
func (h *SearchHandler) GetSearchHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	history, err := h.searchService.GetSearchHistory(c.Context(), userID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve search history").WithInternal(err))
	}

	return utils.SuccessResponse(c, history, "Search history retrieved successfully")
}

// GetPopularSearches retrieves popular searches
func (h *SearchHandler) GetPopularSearches(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	searches, err := h.searchService.GetPopularSearches(c.Context(), limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve popular searches").WithInternal(err))
	}

	return utils.SuccessResponse(c, searches, "Popular searches retrieved successfully")
}

// ClearSearchHistory clears user's search history
func (h *SearchHandler) ClearSearchHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	err := h.searchService.ClearSearchHistory(c.Context(), userID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to clear search history").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Search history cleared successfully")
}

// DeleteSearchHistoryItem deletes a search history item
func (h *SearchHandler) DeleteSearchHistoryItem(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	historyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid history ID")
	}

	err = h.searchService.DeleteSearchHistoryItem(c.Context(), userID, historyID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to delete search history item").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Search history item deleted successfully")
}
