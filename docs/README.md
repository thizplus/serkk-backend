# API Documentation

This directory contains auto-generated Swagger/OpenAPI documentation for the GoFiber Social Media API.

## 📚 Accessing the Documentation

Once the server is running, you can access the interactive API documentation at:

```
http://localhost:3000/swagger/index.html
```

## 🔄 Regenerating Documentation

If you make changes to API handlers or add new endpoints, regenerate the documentation:

```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

## 📝 Adding Documentation to New Endpoints

Add Swagger annotations above your handler functions:

```go
// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post with title, content, and media
// @Tags Posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param post body dto.CreatePostRequest true "Post data"
// @Success 200 {object} map[string]interface{} "Post created successfully"
// @Failure 400 {object} map[string]interface{} "Validation failed"
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
    // ...
}
```

## 🔐 Authentication

Protected endpoints require a JWT token. In Swagger UI:

1. Click the "Authorize" button (top right)
2. Enter: `Bearer <your-jwt-token>`
3. Click "Authorize"

## 📋 Available Tags

- **Authentication** - User registration and login
- **Users** - User profile management
- **Posts** - Post CRUD operations
- **Health** - Health check endpoints
- **Metrics** - Application metrics

## 📄 Generated Files

- `docs.go` - Go package for swagger docs
- `swagger.json` - OpenAPI specification (JSON)
- `swagger.yaml` - OpenAPI specification (YAML)

## 🔗 Resources

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Fiber Swagger Middleware](https://github.com/gofiber/swagger)
