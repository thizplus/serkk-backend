package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
)

func SetupUploadRoutes(api fiber.Router, h *handlers.Handlers) {
	upload := api.Group("/upload")

	// Protected routes (require authentication)
	upload.Use(middleware.Protected())

	// Legacy file upload route (backend proxy to Bunny)
	upload.Post("/file", h.FileUploadHandler.UploadFile)

	// R2 presigned URL routes (direct frontend upload)
	if h.PresignedUploadHandler != nil {
		upload.Post("/presigned-url", h.PresignedUploadHandler.GeneratePresignedUploadURL)
		upload.Post("/presigned-url-batch", h.PresignedUploadHandler.GenerateBatchPresignedUploadURLs)
		upload.Post("/confirm", h.PresignedUploadHandler.ConfirmUpload)
		upload.Post("/confirm-batch", h.PresignedUploadHandler.BatchConfirmUpload)
	}
}
