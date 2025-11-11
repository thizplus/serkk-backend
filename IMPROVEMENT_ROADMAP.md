# 🚀 IMPROVEMENT ROADMAP - แผนปรับปรุงระบบทีละขั้น

**โปรเจกต์:** GoFiber Backend - Social Media Platform
**เป้าหมาย:** เตรียมพร้อม Production Deployment
**ระยะเวลาโดยประมาณ:** 6-8 สัปดาห์
**วันที่สร้าง:** 2025-11-11

---

## 📋 สารบัญ

- [Overview](#overview)
- [Step 1: Foundation - Testing & Error Handling (Week 1-2)](#step-1-foundation---testing--error-handling-week-1-2)
- [Step 2: Security & Validation (Week 2-3)](#step-2-security--validation-week-2-3)
- [Step 3: Database & Performance (Week 3-4)](#step-3-database--performance-week-3-4)
- [Step 4: Monitoring & Logging (Week 4-5)](#step-4-monitoring--logging-week-4-5)
- [Step 5: Production Readiness (Week 5-6)](#step-5-production-readiness-week-5-6)
- [Step 6: Optimization & Polish (Week 6-8)](#step-6-optimization--polish-week-6-8)
- [Progress Tracking](#progress-tracking)

---

## 🎯 Overview

### ลำดับความสำคัญ

```
STEP 1 (Week 1-2) → STEP 2 (Week 2-3) → STEP 3 (Week 3-4)
       ↓                    ↓                    ↓
   Foundation          Security            Performance
       ↓                    ↓                    ↓
STEP 4 (Week 4-5) → STEP 5 (Week 5-6) → STEP 6 (Week 6-8)
       ↓                    ↓                    ↓
   Monitoring          Production           Optimization
```

### กฎสำคัญ
1. ✅ **ห้ามข้าม Step** - ทำตามลำดับเท่านั้น
2. ✅ **ทำให้เสร็จ 100% แต่ละ Step** - ค่อยไป Step ถัดไป
3. ✅ **Test ทุกอย่างที่เปลี่ยน** - ก่อนไป Step ถัดไป
4. ✅ **Commit บ่อยๆ** - ทุก sub-step ที่เสร็จ

---

# STEP 1: Foundation - Testing & Error Handling (Week 1-2)

**เป้าหมาย:** สร้างรากฐานที่แข็งแรง - Testing Infrastructure และ Error Handling
**ระยะเวลา:** 10-14 วัน
**ความสำคัญ:** 🔴 CRITICAL

## ทำไมต้องทำ Step นี้ก่อน?
- ไม่มี tests = ไม่กล้าแก้โค้ด = ไม่สามารถ refactor ได้
- Error handling ดี = debug ง่าย = แก้ bug เร็ว
- มี tests แล้ว = มั่นใจทุก step ถัดไป

---

## 1.1 สร้าง Error Handling System (Day 1-2)

### Checklist
- [ ] สร้างไฟล์ `pkg/errors/errors.go`
- [ ] สร้าง custom error types
- [ ] สร้าง error codes
- [ ] Update error response utilities
- [ ] Test error handling

### Implementation

#### 📝 สร้างไฟล์ `pkg/errors/errors.go`

```go
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

// Predefined errors
var (
    // 400 - Bad Request
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

    // 401 - Unauthorized
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

    // 403 - Forbidden
    ErrForbidden = &AppError{
        Code:       "FORBIDDEN",
        Message:    "Access denied",
        StatusCode: http.StatusForbidden,
    }

    // 404 - Not Found
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

    // 409 - Conflict
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

    // 500 - Internal Server Error
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
)

// IsAppError checks if error is AppError
func IsAppError(err error) (*AppError, bool) {
    appErr, ok := err.(*AppError)
    return appErr, ok
}
```

#### 📝 Update `pkg/utils/response.go`

```go
package utils

import (
    "github.com/gofiber/fiber/v2"
    apperrors "gofiber-backend/pkg/errors"
    "github.com/rs/zerolog/log"
)

type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
    Code    string            `json:"code"`
    Message string            `json:"message"`
    Fields  map[string]string `json:"fields,omitempty"`
}

// SuccessResponse returns success response
func SuccessResponse(c *fiber.Ctx, data interface{}, message string) error {
    return c.JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

// ErrorResponse returns error response
func ErrorResponse(c *fiber.Ctx, err error) error {
    // Check if it's AppError
    if appErr, ok := apperrors.IsAppError(err); ok {
        // Log internal error (don't send to client)
        if appErr.Internal != nil {
            log.Error().
                Err(appErr.Internal).
                Str("code", appErr.Code).
                Str("path", c.Path()).
                Str("method", c.Method()).
                Msg("Application error")
        }

        return c.Status(appErr.StatusCode).JSON(Response{
            Success: false,
            Error: &ErrorDetail{
                Code:    appErr.Code,
                Message: appErr.Message,
                Fields:  appErr.Fields,
            },
        })
    }

    // Unknown error - log and return generic error
    log.Error().
        Err(err).
        Str("path", c.Path()).
        Str("method", c.Method()).
        Msg("Unknown error")

    return c.Status(500).JSON(Response{
        Success: false,
        Error: &ErrorDetail{
            Code:    "INTERNAL_ERROR",
            Message: "An unexpected error occurred",
        },
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
```

#### 📝 ตัวอย่างการใช้งานใน Service

```go
// application/serviceimpl/auth_service_impl.go
package serviceimpl

import (
    "context"
    apperrors "gofiber-backend/pkg/errors"
    "golang.org/x/crypto/bcrypt"
)

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
    // Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        // User not found
        return nil, apperrors.ErrInvalidCredentials
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return nil, apperrors.ErrInvalidCredentials
    }

    // Generate token
    token, err := s.jwtUtil.GenerateToken(user.ID)
    if err != nil {
        return nil, apperrors.ErrInternal.WithInternal(err)
    }

    return &LoginResponse{Token: token}, nil
}

func (s *AuthServiceImpl) Register(ctx context.Context, dto RegisterDTO) (*User, error) {
    // Check email exists
    exists, err := s.userRepo.EmailExists(ctx, dto.Email)
    if err != nil {
        return nil, apperrors.ErrDatabase.WithInternal(err)
    }
    if exists {
        return nil, apperrors.ErrEmailExists
    }

    // Check username exists
    exists, err = s.userRepo.UsernameExists(ctx, dto.Username)
    if err != nil {
        return nil, apperrors.ErrDatabase.WithInternal(err)
    }
    if exists {
        return nil, apperrors.ErrUsernameExists
    }

    // Create user
    user, err := s.userRepo.Create(ctx, dto)
    if err != nil {
        return nil, apperrors.ErrDatabase.WithInternal(err)
    }

    return user, nil
}
```

#### 📝 Update Handlers

```go
// interfaces/api/handlers/auth_handler.go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "gofiber-backend/pkg/utils"
)

func (h *AuthHandler) Login(c *fiber.Ctx) error {
    var dto LoginDTO
    if err := c.BodyParser(&dto); err != nil {
        return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithInternal(err))
    }

    // Validate
    if err := h.validator.Struct(dto); err != nil {
        return utils.ErrorResponse(c, apperrors.ErrValidation.WithInternal(err))
    }

    // Call service
    result, err := h.authService.Login(c.Context(), dto.Email, dto.Password)
    if err != nil {
        return utils.ErrorResponse(c, err)
    }

    return utils.SuccessResponse(c, result, "Login successful")
}
```

### ✅ ตรวจสอบ
- [ ] Error types ครบทุก use case
- [ ] ทดสอบ error response format
- [ ] Internal errors ไม่ถูกส่งไปที่ client
- [ ] Error logging ทำงาน

---

## 1.2 Setup Testing Infrastructure (Day 3-4)

### Checklist
- [ ] ติดตั้ง testing libraries
- [ ] สร้างโครงสร้าง test files
- [ ] Setup test database
- [ ] สร้าง test helpers/fixtures
- [ ] Configure CI for tests

### Implementation

#### 📝 ติดตั้ง Dependencies

```bash
# Testing libraries
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/suite
go get github.com/DATA-DOG/go-sqlmock
go get github.com/go-faker/faker/v4

# HTTP testing
go get github.com/valyala/fasthttp
```

#### 📝 สร้าง Test Database Configuration

```go
// pkg/config/test_config.go
package config

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// GetTestDB returns test database connection
func GetTestDB() (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
        "localhost",
        "postgres",
        "postgres",
        "gofiber_test", // Separate test database
        "5432",
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent), // Silent in tests
    })

    if err != nil {
        return nil, err
    }

    return db, nil
}

// CleanTestDB cleans all tables
func CleanTestDB(db *gorm.DB) error {
    // Disable foreign key checks
    db.Exec("SET CONSTRAINTS ALL DEFERRED")

    // Get all tables
    tables := []string{
        "messages", "conversations", "conversation_participants",
        "notifications", "push_subscriptions",
        "comments", "votes", "saved_posts",
        "post_tags", "tags", "post_media", "media",
        "posts", "follows", "users", "search_histories",
    }

    // Truncate all tables
    for _, table := range tables {
        db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
    }

    return nil
}
```

#### 📝 สร้าง Test Helpers

```go
// pkg/testutil/helpers.go
package testutil

import (
    "context"
    "github.com/go-faker/faker/v4"
    "github.com/google/uuid"
    "gofiber-backend/domain/models"
    "golang.org/x/crypto/bcrypt"
    "time"
)

// CreateTestUser creates a test user
func CreateTestUser() *models.User {
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

    return &models.User{
        ID:           uuid.New(),
        Email:        faker.Email(),
        Username:     faker.Username(),
        PasswordHash: string(hashedPassword),
        FullName:     faker.Name(),
        Bio:          faker.Sentence(),
        ProfileImage: "https://example.com/avatar.jpg",
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
}

// CreateTestPost creates a test post
func CreateTestPost(authorID uuid.UUID) *models.Post {
    return &models.Post{
        ID:        uuid.New(),
        AuthorID:  authorID,
        Title:     faker.Sentence(),
        Content:   faker.Paragraph(),
        Votes:     0,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

// CreateTestContext creates test context with timeout
func CreateTestContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 5*time.Second)
}
```

#### 📝 Mock Repository

```go
// domain/repositories/mocks/user_repository_mock.go
package mocks

import (
    "context"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    "gofiber-backend/domain/models"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
    args := m.Called(ctx, email)
    return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
    args := m.Called(ctx, username)
    return args.Bool(0), args.Error(1)
}
```

### ✅ ตรวจสอบ
- [ ] Test database สร้างได้
- [ ] Test helpers ใช้งานได้
- [ ] Mock repositories ทำงาน
- [ ] สามารถรัน test ได้ (`go test ./...`)

---

## 1.3 Write Unit Tests - Service Layer (Day 5-8)

### Checklist
- [ ] Test AuthService (Login, Register, Logout)
- [ ] Test UserService (CRUD, Follow)
- [ ] Test PostService (Create, Update, Delete, Vote)
- [ ] Test CommentService
- [ ] Test NotificationService
- [ ] Coverage >= 70% for services

### Implementation

#### 📝 ตัวอย่าง: Test AuthService

```go
// application/serviceimpl/auth_service_impl_test.go
package serviceimpl

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gofiber-backend/domain/models"
    "gofiber-backend/domain/repositories/mocks"
    apperrors "gofiber-backend/pkg/errors"
    "gofiber-backend/pkg/testutil"
    "golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login_Success(t *testing.T) {
    // Arrange
    mockUserRepo := new(mocks.MockUserRepository)
    authService := NewAuthService(mockUserRepo)

    ctx := context.Background()
    email := "test@example.com"
    password := "password123"

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    user := &models.User{
        ID:           uuid.New(),
        Email:        email,
        PasswordHash: string(hashedPassword),
    }

    mockUserRepo.On("FindByEmail", ctx, email).Return(user, nil)

    // Act
    result, err := authService.Login(ctx, email, password)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.Token)
    mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
    // Arrange
    mockUserRepo := new(mocks.MockUserRepository)
    authService := NewAuthService(mockUserRepo)

    ctx := context.Background()
    email := "notfound@example.com"
    password := "password123"

    mockUserRepo.On("FindByEmail", ctx, email).Return(nil, apperrors.ErrUserNotFound)

    // Act
    result, err := authService.Login(ctx, email, password)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Equal(t, apperrors.ErrInvalidCredentials, err)
    mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
    // Arrange
    mockUserRepo := new(mocks.MockUserRepository)
    authService := NewAuthService(mockUserRepo)

    ctx := context.Background()
    email := "test@example.com"
    correctPassword := "password123"
    wrongPassword := "wrongpassword"

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
    user := &models.User{
        ID:           uuid.New(),
        Email:        email,
        PasswordHash: string(hashedPassword),
    }

    mockUserRepo.On("FindByEmail", ctx, email).Return(user, nil)

    // Act
    result, err := authService.Login(ctx, email, wrongPassword)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Equal(t, apperrors.ErrInvalidCredentials, err)
}

func TestAuthService_Register_Success(t *testing.T) {
    // Arrange
    mockUserRepo := new(mocks.MockUserRepository)
    authService := NewAuthService(mockUserRepo)

    ctx := context.Background()
    dto := RegisterDTO{
        Email:    "new@example.com",
        Username: "newuser",
        Password: "Password123!",
        FullName: "New User",
    }

    mockUserRepo.On("EmailExists", ctx, dto.Email).Return(false, nil)
    mockUserRepo.On("UsernameExists", ctx, dto.Username).Return(false, nil)
    mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

    // Act
    user, err := authService.Register(ctx, dto)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, dto.Email, user.Email)
    assert.Equal(t, dto.Username, user.Username)
    mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
    // Arrange
    mockUserRepo := new(mocks.MockUserRepository)
    authService := NewAuthService(mockUserRepo)

    ctx := context.Background()
    dto := RegisterDTO{
        Email:    "existing@example.com",
        Username: "newuser",
        Password: "Password123!",
    }

    mockUserRepo.On("EmailExists", ctx, dto.Email).Return(true, nil)

    // Act
    user, err := authService.Register(ctx, dto)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, apperrors.ErrEmailExists, err)
    mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_UsernameExists(t *testing.T) {
    // Arrange
    mockUserRepo := new(mocks.MockUserRepository)
    authService := NewAuthService(mockUserRepo)

    ctx := context.Background()
    dto := RegisterDTO{
        Email:    "new@example.com",
        Username: "existinguser",
        Password: "Password123!",
    }

    mockUserRepo.On("EmailExists", ctx, dto.Email).Return(false, nil)
    mockUserRepo.On("UsernameExists", ctx, dto.Username).Return(true, nil)

    // Act
    user, err := authService.Register(ctx, dto)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Equal(t, apperrors.ErrUsernameExists, err)
    mockUserRepo.AssertExpectations(t)
}
```

#### 📝 ตัวอย่าง: Test PostService

```go
// application/serviceimpl/post_service_impl_test.go
package serviceimpl

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gofiber-backend/domain/dto"
    "gofiber-backend/domain/models"
    "gofiber-backend/domain/repositories/mocks"
    apperrors "gofiber-backend/pkg/errors"
)

func TestPostService_CreatePost_Success(t *testing.T) {
    // Arrange
    mockPostRepo := new(mocks.MockPostRepository)
    mockUserRepo := new(mocks.MockUserRepository)
    postService := NewPostService(mockPostRepo, mockUserRepo)

    ctx := context.Background()
    userID := uuid.New()
    createDTO := dto.CreatePostDTO{
        Title:   "Test Post",
        Content: "This is a test post",
        TagIDs:  []uuid.UUID{uuid.New()},
    }

    mockUserRepo.On("FindByID", ctx, userID).Return(&models.User{ID: userID}, nil)
    mockPostRepo.On("Create", ctx, mock.AnythingOfType("*models.Post")).Return(nil)
    mockPostRepo.On("AttachTags", ctx, mock.AnythingOfType("uuid.UUID"), createDTO.TagIDs).Return(nil)

    // Act
    post, err := postService.CreatePost(ctx, userID, createDTO)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, post)
    assert.Equal(t, createDTO.Title, post.Title)
    assert.Equal(t, createDTO.Content, post.Content)
    mockPostRepo.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
}

func TestPostService_CreatePost_UserNotFound(t *testing.T) {
    // Arrange
    mockPostRepo := new(mocks.MockPostRepository)
    mockUserRepo := new(mocks.MockUserRepository)
    postService := NewPostService(mockPostRepo, mockUserRepo)

    ctx := context.Background()
    userID := uuid.New()
    createDTO := dto.CreatePostDTO{Title: "Test", Content: "Content"}

    mockUserRepo.On("FindByID", ctx, userID).Return(nil, apperrors.ErrUserNotFound)

    // Act
    post, err := postService.CreatePost(ctx, userID, createDTO)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, post)
    assert.Equal(t, apperrors.ErrUserNotFound, err)
}

func TestPostService_DeletePost_Success(t *testing.T) {
    // Arrange
    mockPostRepo := new(mocks.MockPostRepository)
    mockUserRepo := new(mocks.MockUserRepository)
    postService := NewPostService(mockPostRepo, mockUserRepo)

    ctx := context.Background()
    postID := uuid.New()
    userID := uuid.New()

    post := &models.Post{
        ID:       postID,
        AuthorID: userID,
    }

    mockPostRepo.On("FindByID", ctx, postID).Return(post, nil)
    mockPostRepo.On("Delete", ctx, postID).Return(nil)

    // Act
    err := postService.DeletePost(ctx, postID, userID)

    // Assert
    assert.NoError(t, err)
    mockPostRepo.AssertExpectations(t)
}

func TestPostService_DeletePost_Forbidden(t *testing.T) {
    // Arrange
    mockPostRepo := new(mocks.MockPostRepository)
    mockUserRepo := new(mocks.MockUserRepository)
    postService := NewPostService(mockPostRepo, mockUserRepo)

    ctx := context.Background()
    postID := uuid.New()
    authorID := uuid.New()
    otherUserID := uuid.New()

    post := &models.Post{
        ID:       postID,
        AuthorID: authorID,
    }

    mockPostRepo.On("FindByID", ctx, postID).Return(post, nil)

    // Act
    err := postService.DeletePost(ctx, postID, otherUserID)

    // Assert
    assert.Error(t, err)
    assert.Equal(t, apperrors.ErrForbidden, err)
}
```

#### 📝 Run Tests และดู Coverage

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test ./application/serviceimpl/... -v

# Run specific test
go test ./application/serviceimpl/... -run TestAuthService_Login_Success -v
```

#### 📝 สร้าง Makefile

```makefile
# Makefile
.PHONY: test test-coverage test-unit test-integration

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run unit tests only
test-unit:
	go test ./application/serviceimpl/... -v
	go test ./pkg/... -v

# Run integration tests
test-integration:
	go test ./infrastructure/... -v -tags=integration

# Check coverage percentage
coverage-check:
	go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'
```

### ✅ ตรวจสอบ
- [ ] Tests ทั้งหมดผ่าน (สีเขียว)
- [ ] Code coverage >= 70%
- [ ] Test ทั้ง success และ error cases
- [ ] Mock ทำงานถูกต้อง

---

## 1.4 Integration Tests - Repository Layer (Day 9-10)

### Checklist
- [ ] Setup test database migration
- [ ] Test UserRepository
- [ ] Test PostRepository
- [ ] Test CommentRepository
- [ ] Test with real database queries

### Implementation

#### 📝 Integration Test Example

```go
// infrastructure/postgres/user_repository_impl_test.go
//go:build integration
// +build integration

package postgres

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "gofiber-backend/domain/models"
    "gofiber-backend/pkg/config"
    "gofiber-backend/pkg/testutil"
    "gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
    suite.Suite
    db   *gorm.DB
    repo *UserRepositoryImpl
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
    // Setup test database
    db, err := config.GetTestDB()
    assert.NoError(suite.T(), err)

    // Run migrations
    db.AutoMigrate(&models.User{})

    suite.db = db
    suite.repo = NewUserRepository(db)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
    sqlDB, _ := suite.db.DB()
    sqlDB.Close()
}

func (suite *UserRepositoryTestSuite) SetupTest() {
    // Clean database before each test
    config.CleanTestDB(suite.db)
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
    ctx := context.Background()
    user := testutil.CreateTestUser()

    err := suite.repo.Create(ctx, user)

    assert.NoError(suite.T(), err)
    assert.NotEqual(suite.T(), uuid.Nil, user.ID)
}

func (suite *UserRepositoryTestSuite) TestFindByEmail_Success() {
    ctx := context.Background()
    user := testutil.CreateTestUser()
    suite.repo.Create(ctx, user)

    found, err := suite.repo.FindByEmail(ctx, user.Email)

    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), found)
    assert.Equal(suite.T(), user.Email, found.Email)
}

func (suite *UserRepositoryTestSuite) TestFindByEmail_NotFound() {
    ctx := context.Background()

    found, err := suite.repo.FindByEmail(ctx, "nonexistent@example.com")

    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), found)
}

func (suite *UserRepositoryTestSuite) TestEmailExists_True() {
    ctx := context.Background()
    user := testutil.CreateTestUser()
    suite.repo.Create(ctx, user)

    exists, err := suite.repo.EmailExists(ctx, user.Email)

    assert.NoError(suite.T(), err)
    assert.True(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestEmailExists_False() {
    ctx := context.Background()

    exists, err := suite.repo.EmailExists(ctx, "nonexistent@example.com")

    assert.NoError(suite.T(), err)
    assert.False(suite.T(), exists)
}

func TestUserRepositoryTestSuite(t *testing.T) {
    suite.Run(t, new(UserRepositoryTestSuite))
}
```

#### 📝 Run Integration Tests

```bash
# Run integration tests only
go test ./infrastructure/... -v -tags=integration

# Run with test database
GOFIBER_ENV=test go test ./infrastructure/... -v -tags=integration
```

### ✅ ตรวจสอบ
- [ ] Integration tests ผ่านหมด
- [ ] Test database isolated
- [ ] Tests ไม่ทิ้ง dirty data
- [ ] Coverage repository layer >= 80%

---

## 📊 Step 1 Completion Checklist

### ก่อนไป Step 2 ต้องเช็คว่า:

- [ ] ✅ Error handling system พร้อมใช้งาน
- [ ] ✅ Error codes ครบถ้วน
- [ ] ✅ Testing infrastructure พร้อม
- [ ] ✅ Unit tests coverage >= 70%
- [ ] ✅ Integration tests ทำงาน
- [ ] ✅ All tests passing (สีเขียวทั้งหมด)
- [ ] ✅ Mock repositories ครบ
- [ ] ✅ Test helpers พร้อมใช้งาน
- [ ] ✅ CI configuration พร้อม

### Commit Message

```bash
git add .
git commit -m "feat: implement error handling and testing infrastructure

- Add structured error types and codes
- Update error response utilities
- Create comprehensive test infrastructure
- Add unit tests for service layer (coverage: 70%+)
- Add integration tests for repository layer
- Setup test database and helpers
- Create mock repositories
- Configure test coverage reporting

BREAKING CHANGE: Error response format changed to structured format"
```

---

# STEP 2: Security & Validation (Week 2-3)

**เป้าหมาย:** เพิ่มความปลอดภัยและ validation ที่แข็งแรง
**ระยะเวลา:** 7-10 วัน
**ความสำคัญ:** 🔴 CRITICAL

## ทำไมต้องทำ Step นี้?
- ป้องกัน security vulnerabilities
- ป้องกัน API abuse
- Validate input ก่อนถึง database
- เพิ่มความน่าเชื่อถือของระบบ

---

## 2.1 Implement Rate Limiting (Day 1-2)

### Checklist
- [ ] ติดตั้ง rate limiter middleware
- [ ] Global rate limiting
- [ ] Per-endpoint rate limiting
- [ ] Login/Register rate limiting
- [ ] Test rate limiting

### Implementation

#### 📝 ติดตั้ง Dependencies

```bash
go get github.com/gofiber/fiber/v2/middleware/limiter
```

#### 📝 สร้าง Rate Limiter Configuration

```go
// pkg/middleware/rate_limiter.go
package middleware

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/limiter"
    "github.com/gofiber/storage/redis/v2"
)

// GlobalRateLimiter returns global rate limiter (100 req/min)
func GlobalRateLimiter(redisStore *redis.Storage) fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        100,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code":    "RATE_LIMIT_EXCEEDED",
                    "message": "Too many requests, please try again later",
                },
            })
        },
        Storage: redisStore, // Use Redis for distributed rate limiting
    })
}

// AuthRateLimiter for login/register endpoints (5 req/min)
func AuthRateLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        5,
        Expiration: 1 * time.Minute,
        KeyGenerator: func(c *fiber.Ctx) string {
            // Limit by IP
            return c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code":    "AUTH_RATE_LIMIT_EXCEEDED",
                    "message": "Too many authentication attempts, please try again in 1 minute",
                },
            })
        },
    })
}

// APIKeyRateLimiter for API endpoints (1000 req/hour per user)
func APIKeyRateLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        1000,
        Expiration: 1 * time.Hour,
        KeyGenerator: func(c *fiber.Ctx) string {
            // Get user ID from context (set by auth middleware)
            userID := c.Locals("userID")
            if userID != nil {
                return userID.(string)
            }
            return c.IP()
        },
    })
}

// UploadRateLimiter for file upload endpoints (10 uploads/hour)
func UploadRateLimiter() fiber.Handler {
    return limiter.New(limiter.Config{
        Max:        10,
        Expiration: 1 * time.Hour,
        KeyGenerator: func(c *fiber.Ctx) string {
            userID := c.Locals("userID")
            if userID != nil {
                return "upload:" + userID.(string)
            }
            return "upload:" + c.IP()
        },
        LimitReached: func(c *fiber.Ctx) error {
            return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
                "success": false,
                "error": fiber.Map{
                    "code":    "UPLOAD_RATE_LIMIT_EXCEEDED",
                    "message": "Upload limit exceeded, please try again later",
                },
            })
        },
    })
}
```

#### 📝 Apply Rate Limiters ใน Routes

```go
// interfaces/api/routes/routes.go
package routes

import (
    "github.com/gofiber/fiber/v2"
    "gofiber-backend/pkg/middleware"
)

func SetupRoutes(app *fiber.App, handlers *handlers.Handlers, redisStore *redis.Storage) {
    api := app.Group("/api/v1")

    // Global rate limiter
    api.Use(middleware.GlobalRateLimiter(redisStore))

    // Auth routes with stricter rate limiting
    authRoutes := api.Group("/auth")
    authRoutes.Use(middleware.AuthRateLimiter())
    {
        authRoutes.Post("/register", handlers.Auth.Register)
        authRoutes.Post("/login", handlers.Auth.Login)
        authRoutes.Post("/forgot-password", handlers.Auth.ForgotPassword)
    }

    // Protected routes with API rate limiting
    protected := api.Group("")
    protected.Use(middleware.JWTMiddleware())
    protected.Use(middleware.APIKeyRateLimiter())
    {
        // User routes
        protected.Get("/users/:id", handlers.User.GetProfile)
        protected.Put("/users/me", handlers.User.UpdateProfile)

        // Post routes
        protected.Post("/posts", handlers.Post.CreatePost)
        protected.Put("/posts/:id", handlers.Post.UpdatePost)
        protected.Delete("/posts/:id", handlers.Post.DeletePost)
    }

    // Upload routes with upload-specific rate limiting
    uploadRoutes := protected.Group("/upload")
    uploadRoutes.Use(middleware.UploadRateLimiter())
    {
        uploadRoutes.Post("/image", handlers.Media.UploadImage)
        uploadRoutes.Post("/video", handlers.Media.UploadVideo)
    }
}
```

#### 📝 Test Rate Limiting

```go
// pkg/middleware/rate_limiter_test.go
package middleware

import (
    "net/http/httptest"
    "testing"
    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
)

func TestAuthRateLimiter_ExceedsLimit(t *testing.T) {
    app := fiber.New()

    app.Use(AuthRateLimiter())
    app.Post("/login", func(c *fiber.Ctx) error {
        return c.SendString("OK")
    })

    // Make 5 requests (should succeed)
    for i := 0; i < 5; i++ {
        req := httptest.NewRequest("POST", "/login", nil)
        resp, _ := app.Test(req)
        assert.Equal(t, 200, resp.StatusCode)
    }

    // 6th request should be rate limited
    req := httptest.NewRequest("POST", "/login", nil)
    resp, _ := app.Test(req)
    assert.Equal(t, 429, resp.StatusCode)
}
```

### ✅ ตรวจสอบ
- [ ] Global rate limiter ทำงาน
- [ ] Auth endpoints มี rate limit
- [ ] Upload endpoints มี rate limit
- [ ] Rate limiting tests ผ่าน
- [ ] ใช้ Redis สำหรับ distributed limiting

---

## 2.2 Enhanced Input Validation (Day 3-4)

### Checklist
- [ ] เพิ่ม custom validators
- [ ] Password strength validation
- [ ] Email validation
- [ ] Username validation
- [ ] File upload validation
- [ ] Sanitization

### Implementation

#### 📝 สร้าง Custom Validators

```go
// pkg/validator/custom_validators.go
package validator

import (
    "regexp"
    "strings"
    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New()

    // Register custom validators
    validate.RegisterValidation("strong_password", strongPassword)
    validate.RegisterValidation("username", validUsername)
    validate.RegisterValidation("no_xss", noXSS)
    validate.RegisterValidation("file_type", fileType)
}

// GetValidator returns validator instance
func GetValidator() *validator.Validate {
    return validate
}

// strongPassword validates password strength
// Requirements: min 8 chars, 1 uppercase, 1 lowercase, 1 digit, 1 special char
func strongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()

    if len(password) < 8 {
        return false
    }

    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

    return hasUpper && hasLower && hasDigit && hasSpecial
}

// validUsername validates username format
// Requirements: 3-20 chars, alphanumeric and underscore only, no spaces
func validUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()

    if len(username) < 3 || len(username) > 20 {
        return false
    }

    // Only alphanumeric and underscore
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
    return matched
}

// noXSS checks for potential XSS attempts
func noXSS(fl validator.FieldLevel) bool {
    value := fl.Field().String()

    // Basic XSS patterns
    xssPatterns := []string{
        "<script",
        "javascript:",
        "onerror=",
        "onload=",
        "<iframe",
    }

    lowerValue := strings.ToLower(value)
    for _, pattern := range xssPatterns {
        if strings.Contains(lowerValue, pattern) {
            return false
        }
    }

    return true
}

// fileType validates file mime type
func fileType(fl validator.FieldLevel) bool {
    allowedTypes := fl.Param()
    fileType := fl.Field().String()

    allowed := strings.Split(allowedTypes, "|")
    for _, t := range allowed {
        if fileType == t {
            return true
        }
    }

    return false
}
```

#### 📝 Update DTOs with Validation

```go
// domain/dto/auth.go
package dto

type RegisterDTO struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,username"`
    Password string `json:"password" validate:"required,strong_password"`
    FullName string `json:"full_name" validate:"required,min=2,max=100,no_xss"`
    Bio      string `json:"bio" validate:"omitempty,max=500,no_xss"`
}

type LoginDTO struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordDTO struct {
    OldPassword string `json:"old_password" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,strong_password,nefield=OldPassword"`
}
```

```go
// domain/dto/post.go
package dto

import "github.com/google/uuid"

type CreatePostDTO struct {
    Title    string      `json:"title" validate:"required,min=3,max=200,no_xss"`
    Content  string      `json:"content" validate:"required,min=10,max=50000,no_xss"`
    TagIDs   []uuid.UUID `json:"tag_ids" validate:"omitempty,max=10"`
    MediaIDs []uuid.UUID `json:"media_ids" validate:"omitempty,max=10"`
}

type UpdatePostDTO struct {
    Title   *string     `json:"title" validate:"omitempty,min=3,max=200,no_xss"`
    Content *string     `json:"content" validate:"omitempty,min=10,max=50000,no_xss"`
    TagIDs  []uuid.UUID `json:"tag_ids" validate:"omitempty,max=10"`
}
```

#### 📝 Validation Middleware

```go
// pkg/middleware/validator.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/go-playground/validator/v10"
    customValidator "gofiber-backend/pkg/validator"
    apperrors "gofiber-backend/pkg/errors"
    "gofiber-backend/pkg/utils"
)

// ValidateDTO validates request DTO
func ValidateDTO(dto interface{}) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Parse body
        if err := c.BodyParser(dto); err != nil {
            return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithInternal(err))
        }

        // Validate
        validate := customValidator.GetValidator()
        if err := validate.Struct(dto); err != nil {
            if validationErrs, ok := err.(validator.ValidationErrors); ok {
                // Build field errors
                appErr := apperrors.ErrValidation
                for _, e := range validationErrs {
                    appErr = appErr.WithField(e.Field(), getValidationErrorMessage(e))
                }
                return utils.ErrorResponse(c, appErr)
            }
            return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithInternal(err))
        }

        return c.Next()
    }
}

// getValidationErrorMessage returns human-readable error message
func getValidationErrorMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "This field is required"
    case "email":
        return "Invalid email address"
    case "strong_password":
        return "Password must be at least 8 characters with uppercase, lowercase, digit, and special character"
    case "username":
        return "Username must be 3-20 characters, alphanumeric and underscore only"
    case "no_xss":
        return "Content contains potentially dangerous characters"
    case "min":
        return fmt.Sprintf("Minimum length is %s", e.Param())
    case "max":
        return fmt.Sprintf("Maximum length is %s", e.Param())
    default:
        return "Invalid value"
    }
}
```

### ✅ ตรวจสอบ
- [ ] Password strength validation ทำงาน
- [ ] Username validation ทำงาน
- [ ] XSS protection ทำงาน
- [ ] Validation error messages ชัดเจน
- [ ] Tests สำหรับ validators ผ่าน

---

## 2.3 File Upload Security (Day 5-6)

### Checklist
- [ ] File type validation (magic bytes)
- [ ] File size limits per type
- [ ] Filename sanitization
- [ ] Virus scanning (optional)
- [ ] Test file upload security

### Implementation

#### 📝 File Upload Validator

```go
// pkg/validator/file_validator.go
package validator

import (
    "errors"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "path/filepath"
    "regexp"
    "strings"
)

var (
    ErrInvalidFileType = errors.New("invalid file type")
    ErrFileTooLarge    = errors.New("file too large")
    ErrInvalidFileName = errors.New("invalid file name")
)

// FileConfig holds file validation configuration
type FileConfig struct {
    AllowedMimeTypes []string
    MaxSize          int64
    AllowedExts      []string
}

// Predefined configs
var (
    ImageConfig = FileConfig{
        AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
        MaxSize:          10 * 1024 * 1024, // 10MB
        AllowedExts:      []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
    }

    VideoConfig = FileConfig{
        AllowedMimeTypes: []string{"video/mp4", "video/quicktime", "video/x-msvideo"},
        MaxSize:          100 * 1024 * 1024, // 100MB
        AllowedExts:      []string{".mp4", ".mov", ".avi"},
    }

    DocumentConfig = FileConfig{
        AllowedMimeTypes: []string{"application/pdf", "application/msword"},
        MaxSize:          20 * 1024 * 1024, // 20MB
        AllowedExts:      []string{".pdf", ".doc", ".docx"},
    }
)

// ValidateFile validates uploaded file
func ValidateFile(fileHeader *multipart.FileHeader, config FileConfig) error {
    // Check file size
    if fileHeader.Size > config.MaxSize {
        return fmt.Errorf("%w: max size %d bytes", ErrFileTooLarge, config.MaxSize)
    }

    // Sanitize filename
    if err := ValidateFileName(fileHeader.Filename); err != nil {
        return err
    }

    // Check extension
    ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
    if !contains(config.AllowedExts, ext) {
        return fmt.Errorf("%w: allowed extensions: %v", ErrInvalidFileType, config.AllowedExts)
    }

    // Check MIME type using magic bytes
    file, err := fileHeader.Open()
    if err != nil {
        return err
    }
    defer file.Close()

    // Read first 512 bytes for MIME detection
    buffer := make([]byte, 512)
    _, err = file.Read(buffer)
    if err != nil && err != io.EOF {
        return err
    }

    // Detect content type from magic bytes
    contentType := http.DetectContentType(buffer)
    if !contains(config.AllowedMimeTypes, contentType) {
        return fmt.Errorf("%w: detected %s, allowed: %v", ErrInvalidFileType, contentType, config.AllowedMimeTypes)
    }

    return nil
}

// ValidateFileName sanitizes and validates filename
func ValidateFileName(filename string) error {
    // Check length
    if len(filename) > 255 {
        return fmt.Errorf("%w: filename too long", ErrInvalidFileName)
    }

    // Check for path traversal attempts
    if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
        return fmt.Errorf("%w: path traversal detected", ErrInvalidFileName)
    }

    // Check for null bytes
    if strings.Contains(filename, "\x00") {
        return fmt.Errorf("%w: null byte detected", ErrInvalidFileName)
    }

    // Only allow alphanumeric, dash, underscore, and dot
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, filename)
    if !matched {
        return fmt.Errorf("%w: filename contains invalid characters", ErrInvalidFileName)
    }

    return nil
}

// SanitizeFileName returns sanitized filename
func SanitizeFileName(filename string) string {
    // Remove path
    filename = filepath.Base(filename)

    // Replace spaces with underscores
    filename = strings.ReplaceAll(filename, " ", "_")

    // Remove any character that's not alphanumeric, dash, underscore, or dot
    reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
    filename = reg.ReplaceAllString(filename, "")

    // Limit length
    if len(filename) > 200 {
        ext := filepath.Ext(filename)
        filename = filename[:200-len(ext)] + ext
    }

    return filename
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

#### 📝 Update Media Handler

```go
// interfaces/api/handlers/media_handler.go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "gofiber-backend/pkg/validator"
    apperrors "gofiber-backend/pkg/errors"
    "gofiber-backend/pkg/utils"
)

func (h *MediaHandler) UploadImage(c *fiber.Ctx) error {
    // Get file from request
    file, err := c.FormFile("file")
    if err != nil {
        return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("No file uploaded"))
    }

    // Validate file
    if err := validator.ValidateFile(file, validator.ImageConfig); err != nil {
        if errors.Is(err, validator.ErrInvalidFileType) {
            return utils.ErrorResponse(c, apperrors.ErrValidation.WithMessage(err.Error()))
        }
        if errors.Is(err, validator.ErrFileTooLarge) {
            return utils.ErrorResponse(c, apperrors.ErrValidation.WithMessage("Image file too large (max 10MB)"))
        }
        return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithInternal(err))
    }

    // Sanitize filename
    sanitizedName := validator.SanitizeFileName(file.Filename)

    // Upload to storage
    userID := c.Locals("userID").(uuid.UUID)
    media, err := h.mediaService.UploadImage(c.Context(), file, sanitizedName, userID)
    if err != nil {
        return utils.ErrorResponse(c, err)
    }

    return utils.Created(c, media, "Image uploaded successfully")
}

func (h *MediaHandler) UploadVideo(c *fiber.Ctx) error {
    file, err := c.FormFile("file")
    if err != nil {
        return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("No file uploaded"))
    }

    // Validate video file
    if err := validator.ValidateFile(file, validator.VideoConfig); err != nil {
        if errors.Is(err, validator.ErrFileTooLarge) {
            return utils.ErrorResponse(c, apperrors.ErrValidation.WithMessage("Video file too large (max 100MB)"))
        }
        return utils.ErrorResponse(c, apperrors.ErrValidation.WithMessage(err.Error()))
    }

    sanitizedName := validator.SanitizeFileName(file.Filename)
    userID := c.Locals("userID").(uuid.UUID)

    media, err := h.mediaService.UploadVideo(c.Context(), file, sanitizedName, userID)
    if err != nil {
        return utils.ErrorResponse(c, err)
    }

    return utils.Created(c, media, "Video uploaded successfully")
}
```

### ✅ ตรวจสอบ
- [ ] File type validation ด้วย magic bytes
- [ ] Size limits per file type
- [ ] Filename sanitization ทำงาน
- [ ] Path traversal protection
- [ ] Tests ผ่านหมด

---

## 2.4 Security Headers & CORS (Day 7-8)

### Checklist
- [ ] Security headers middleware
- [ ] CORS configuration
- [ ] CSP headers
- [ ] HSTS headers
- [ ] Test security headers

### Implementation

#### 📝 Security Headers Middleware

```go
// pkg/middleware/security.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/helmet"
)

// SecurityHeaders adds security headers
func SecurityHeaders() fiber.Handler {
    return helmet.New(helmet.Config{
        XSSProtection:             "1; mode=block",
        ContentTypeNosniff:        "nosniff",
        XFrameOptions:             "SAMEORIGIN",
        ReferrerPolicy:            "strict-origin-when-cross-origin",
        CrossOriginEmbedderPolicy: "require-corp",
        CrossOriginOpenerPolicy:   "same-origin",
        CrossOriginResourcePolicy: "same-origin",
        OriginAgentCluster:        "?1",
        XDNSPrefetchControl:       "off",
        XDownloadOptions:          "noopen",
        XPermittedCrossDomain:     "none",
    })
}

// HSTS enables HTTP Strict Transport Security
func HSTS() fiber.Handler {
    return func(c *fiber.Ctx) error {
        c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        return c.Next()
    }
}

// CSP sets Content Security Policy
func CSP() fiber.Handler {
    return func(c *fiber.Ctx) error {
        csp := "default-src 'self'; " +
            "script-src 'self' 'unsafe-inline'; " +
            "style-src 'self' 'unsafe-inline'; " +
            "img-src 'self' data: https:; " +
            "font-src 'self' data:; " +
            "connect-src 'self' https:; " +
            "frame-ancestors 'none';"

        c.Set("Content-Security-Policy", csp)
        return c.Next()
    }
}
```

#### 📝 Update CORS Configuration

```go
// pkg/middleware/cors.go
package middleware

import (
    "os"
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns configured CORS middleware
func CORS() fiber.Handler {
    allowedOrigins := getAllowedOrigins()

    return cors.New(cors.Config{
        AllowOrigins:     strings.Join(allowedOrigins, ","),
        AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours
        AllowOriginsFunc: func(origin string) bool {
            // Check if origin is in allowed list
            for _, allowed := range allowedOrigins {
                if origin == allowed {
                    return true
                }
            }
            return false
        },
    })
}

func getAllowedOrigins() []string {
    env := os.Getenv("GOFIBER_ENV")

    switch env {
    case "production":
        // Production: only allow specific domains
        origins := os.Getenv("ALLOWED_ORIGINS")
        if origins == "" {
            return []string{"https://yourdomain.com"}
        }
        return strings.Split(origins, ",")

    case "staging":
        return []string{
            "https://staging.yourdomain.com",
            "https://preview.yourdomain.com",
        }

    default: // development
        return []string{
            "http://localhost:3000",
            "http://localhost:5173",
            "http://127.0.0.1:3000",
            "http://127.0.0.1:5173",
        }
    }
}
```

#### 📝 Apply in Main

```go
// cmd/api/main.go
package main

import (
    "github.com/gofiber/fiber/v2"
    "gofiber-backend/pkg/middleware"
)

func main() {
    app := fiber.New()

    // Security middleware (order matters!)
    app.Use(middleware.SecurityHeaders())
    app.Use(middleware.HSTS())          // HTTPS only
    app.Use(middleware.CSP())
    app.Use(middleware.CORS())
    app.Use(middleware.GlobalRateLimiter(redisStore))

    // ... rest of setup
}
```

### ✅ ตรวจสอบ
- [ ] Security headers ถูกส่ง
- [ ] CORS ทำงานถูกต้อง
- [ ] CSP policy เหมาะสม
- [ ] HSTS enabled (production)

---

## 📊 Step 2 Completion Checklist

- [ ] ✅ Rate limiting ใช้งานได้
- [ ] ✅ Custom validators พร้อม
- [ ] ✅ Password strength enforced
- [ ] ✅ File upload security ครบ
- [ ] ✅ Security headers configured
- [ ] ✅ CORS configured properly
- [ ] ✅ All tests passing
- [ ] ✅ Security audit passed

### Commit Message

```bash
git add .
git commit -m "feat: implement comprehensive security measures

- Add rate limiting (global, auth, upload)
- Implement custom validators (password, username, XSS)
- Add file upload security (magic bytes, size limits)
- Configure security headers (CSP, HSTS, X-Frame-Options)
- Update CORS configuration for environments
- Add validation middleware
- Write security tests

Security: Rate limiting prevents brute force and DoS attacks"
```

---

# STEP 3: Database & Performance (Week 3-4)

**เป้าหมาย:** ปรับปรุงประสิทธิภาพ database และ add transactions
**ระยะเวลา:** 7-10 วัน
**ความสำคัญ:** 🟠 HIGH

## ทำไมต้องทำ Step นี้?
- Database transactions = data integrity
- Indexes = query ไว 10-100 เท่า
- Connection pooling = handle traffic มากขึ้น
- Caching = ลด database load

---

## 3.1 Database Transactions (Day 1-3)

### Checklist
- [ ] สร้าง transaction wrapper
- [ ] Update repositories รองรับ transactions
- [ ] Update services ใช้ transactions
- [ ] Test rollback scenarios
- [ ] Test concurrent transactions

### Implementation

#### 📝 สร้าง Transaction Manager

```go
// pkg/database/transaction.go
package database

import (
    "context"
    "gorm.io/gorm"
)

// TxManager manages database transactions
type TxManager struct {
    db *gorm.DB
}

// NewTxManager creates new transaction manager
func NewTxManager(db *gorm.DB) *TxManager {
    return &TxManager{db: db}
}

// WithTransaction executes function within a transaction
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
    return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        return fn(tx)
    })
}

// GetDB returns database instance (for use in transaction)
func (tm *TxManager) GetDB(tx *gorm.DB) *gorm.DB {
    if tx != nil {
        return tx
    }
    return tm.db
}
```

#### 📝 Update Repository Interface

```go
// domain/repositories/post_repository.go
package repositories

import (
    "context"
    "github.com/google/uuid"
    "gofiber-backend/domain/models"
    "gorm.io/gorm"
)

type PostRepository interface {
    // Basic CRUD
    Create(ctx context.Context, post *models.Post) error
    CreateTx(ctx context.Context, tx *gorm.DB, post *models.Post) error

    FindByID(ctx context.Context, id uuid.UUID) (*models.Post, error)
    Update(ctx context.Context, post *models.Post) error
    Delete(ctx context.Context, id uuid.UUID) error

    // With relations
    AttachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error
    AttachTagsTx(ctx context.Context, tx *gorm.DB, postID uuid.UUID, tagIDs []uuid.UUID) error

    AttachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error
    AttachMediaTx(ctx context.Context, tx *gorm.DB, postID uuid.UUID, mediaIDs []uuid.UUID) error

    // Complex operations
    CreateWithRelations(ctx context.Context, post *models.Post, tagIDs, mediaIDs []uuid.UUID) error
}
```

#### 📝 Implement Transaction Methods

```go
// infrastructure/postgres/post_repository_impl.go
package postgres

import (
    "context"
    "github.com/google/uuid"
    "gofiber-backend/domain/models"
    apperrors "gofiber-backend/pkg/errors"
    "gorm.io/gorm"
)

func (r *PostRepositoryImpl) CreateTx(ctx context.Context, tx *gorm.DB, post *models.Post) error {
    if err := tx.WithContext(ctx).Create(post).Error; err != nil {
        return apperrors.ErrDatabase.WithInternal(err)
    }
    return nil
}

func (r *PostRepositoryImpl) AttachTagsTx(ctx context.Context, tx *gorm.DB, postID uuid.UUID, tagIDs []uuid.UUID) error {
    if len(tagIDs) == 0 {
        return nil
    }

    // Create post_tags associations
    for _, tagID := range tagIDs {
        postTag := &models.PostTag{
            PostID: postID,
            TagID:  tagID,
        }
        if err := tx.WithContext(ctx).Create(postTag).Error; err != nil {
            return apperrors.ErrDatabase.WithInternal(err)
        }
    }
    return nil
}

func (r *PostRepositoryImpl) AttachMediaTx(ctx context.Context, tx *gorm.DB, postID uuid.UUID, mediaIDs []uuid.UUID) error {
    if len(mediaIDs) == 0 {
        return nil
    }

    for i, mediaID := range mediaIDs {
        postMedia := &models.PostMedia{
            PostID:       postID,
            MediaID:      mediaID,
            DisplayOrder: i,
        }
        if err := tx.WithContext(ctx).Create(postMedia).Error; err != nil {
            return apperrors.ErrDatabase.WithInternal(err)
        }
    }
    return nil
}

// CreateWithRelations creates post with tags and media in a transaction
func (r *PostRepositoryImpl) CreateWithRelations(ctx context.Context, post *models.Post, tagIDs, mediaIDs []uuid.UUID) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 1. Create post
        if err := r.CreateTx(ctx, tx, post); err != nil {
            return err
        }

        // 2. Attach tags
        if err := r.AttachTagsTx(ctx, tx, post.ID, tagIDs); err != nil {
            return err
        }

        // 3. Attach media
        if err := r.AttachMediaTx(ctx, tx, post.ID, mediaIDs); err != nil {
            return err
        }

        return nil
    })
}
```

#### 📝 Update Service to Use Transactions

```go
// application/serviceimpl/post_service_impl.go
package serviceimpl

import (
    "context"
    "github.com/google/uuid"
    "gofiber-backend/domain/dto"
    "gofiber-backend/domain/models"
    apperrors "gofiber-backend/pkg/errors"
)

func (s *PostServiceImpl) CreatePost(ctx context.Context, userID uuid.UUID, dto dto.CreatePostDTO) (*models.Post, error) {
    // Validate user exists
    if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
        return nil, apperrors.ErrUserNotFound
    }

    // Create post object
    post := &models.Post{
        ID:        uuid.New(),
        AuthorID:  userID,
        Title:     dto.Title,
        Content:   dto.Content,
        Votes:     0,
        Status:    "draft",
    }

    // Create post with relations in transaction
    if err := s.postRepo.CreateWithRelations(ctx, post, dto.TagIDs, dto.MediaIDs); err != nil {
        return nil, err
    }

    // Load relations for response
    post, err := s.postRepo.FindByIDWithRelations(ctx, post.ID)
    if err != nil {
        return nil, err
    }

    return post, nil
}

func (s *PostServiceImpl) UpdatePost(ctx context.Context, postID, userID uuid.UUID, dto dto.UpdatePostDTO) (*models.Post, error) {
    // Find post
    post, err := s.postRepo.FindByID(ctx, postID)
    if err != nil {
        return nil, apperrors.ErrPostNotFound
    }

    // Check ownership
    if post.AuthorID != userID {
        return nil, apperrors.ErrForbidden
    }

    // Use transaction for update
    err = s.db.Transaction(func(tx *gorm.DB) error {
        // Update post fields
        if dto.Title != nil {
            post.Title = *dto.Title
        }
        if dto.Content != nil {
            post.Content = *dto.Content
        }

        if err := s.postRepo.UpdateTx(ctx, tx, post); err != nil {
            return err
        }

        // Update tags if provided
        if dto.TagIDs != nil {
            // Remove old tags
            if err := s.postRepo.RemoveAllTagsTx(ctx, tx, postID); err != nil {
                return err
            }
            // Attach new tags
            if err := s.postRepo.AttachTagsTx(ctx, tx, postID, dto.TagIDs); err != nil {
                return err
            }
        }

        return nil
    })

    if err != nil {
        return nil, err
    }

    // Reload with relations
    return s.postRepo.FindByIDWithRelations(ctx, postID)
}
```

#### 📝 Test Transactions

```go
// application/serviceimpl/post_service_impl_test.go
package serviceimpl

func TestPostService_CreatePost_RollbackOnError(t *testing.T) {
    // Arrange
    mockPostRepo := new(mocks.MockPostRepository)
    mockUserRepo := new(mocks.MockUserRepository)
    postService := NewPostService(mockPostRepo, mockUserRepo)

    ctx := context.Background()
    userID := uuid.New()
    createDTO := dto.CreatePostDTO{
        Title:    "Test Post",
        Content:  "Content",
        TagIDs:   []uuid.UUID{uuid.New()},
        MediaIDs: []uuid.UUID{uuid.New()},
    }

    mockUserRepo.On("FindByID", ctx, userID).Return(&models.User{ID: userID}, nil)

    // Simulate error when attaching tags (should rollback post creation)
    mockPostRepo.On("CreateWithRelations", ctx, mock.AnythingOfType("*models.Post"), createDTO.TagIDs, createDTO.MediaIDs).
        Return(errors.New("database error"))

    // Act
    post, err := postService.CreatePost(ctx, userID, createDTO)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, post)

    // Verify post was NOT created (transaction rolled back)
    mockPostRepo.AssertNotCalled(t, "FindByID")
}
```

### ✅ ตรวจสอบ
- [ ] Transaction wrapper ทำงาน
- [ ] Repositories รองรับ transactions
- [ ] Services ใช้ transactions ถูกต้อง
- [ ] Rollback ทำงานเมื่อเกิด error
- [ ] Tests ผ่านหมด

---

## 3.2 Add Database Indexes (Day 4-5)

### Checklist
- [ ] วิเคราะห์ slow queries
- [ ] สร้าง migration สำหรับ indexes
- [ ] เพิ่ม indexes ตาม query patterns
- [ ] Test query performance
- [ ] Verify index usage

### Implementation

#### 📝 สร้าง Index Migration

```sql
-- migrations/014_add_performance_indexes.sql

-- Posts indexes
CREATE INDEX IF NOT EXISTS idx_posts_author_created ON posts(author_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_votes ON posts(votes DESC);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC);

-- Comments indexes
CREATE INDEX IF NOT EXISTS idx_comments_post_created ON comments(post_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_author ON comments(author_id);

-- Votes indexes (composite for uniqueness check)
CREATE INDEX IF NOT EXISTS idx_votes_target_user ON votes(target_id, target_type, user_id);
CREATE INDEX IF NOT EXISTS idx_votes_user_created ON votes(user_id, created_at DESC);

-- Follows indexes
CREATE INDEX IF NOT EXISTS idx_follows_follower ON follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_follows_followed ON follows(followed_id);
CREATE INDEX IF NOT EXISTS idx_follows_unique ON follows(follower_id, followed_id);

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_read_created ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- Messages indexes
CREATE INDEX IF NOT EXISTS idx_messages_conversation_created ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);

-- Conversation participants indexes
CREATE INDEX IF NOT EXISTS idx_conv_participants_user ON conversation_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_conv_participants_conversation ON conversation_participants(conversation_id);

-- Media indexes
CREATE INDEX IF NOT EXISTS idx_media_user_created ON media(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_status ON media(status);
CREATE INDEX IF NOT EXISTS idx_media_type ON media(media_type);

-- Post media indexes
CREATE INDEX IF NOT EXISTS idx_post_media_post_order ON post_media(post_id, display_order);
CREATE INDEX IF NOT EXISTS idx_post_media_media ON post_media(media_id);

-- Post tags indexes
CREATE INDEX IF NOT EXISTS idx_post_tags_post ON post_tags(post_id);
CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag_id);

-- Tags indexes
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

-- Users indexes (email, username already have unique constraint which creates index)
CREATE INDEX IF NOT EXISTS idx_users_created ON users(created_at DESC);

-- Search history indexes
CREATE INDEX IF NOT EXISTS idx_search_history_user_created ON search_histories(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_search_history_query ON search_histories(query);

-- Saved posts indexes
CREATE INDEX IF NOT EXISTS idx_saved_posts_user_created ON saved_posts(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_saved_posts_post ON saved_posts(post_id);
CREATE INDEX IF NOT EXISTS idx_saved_posts_unique ON saved_posts(user_id, post_id);
```

#### 📝 Run Migration

```bash
# Apply migration
psql -U postgres -d gofiber_db -f migrations/014_add_performance_indexes.sql

# Verify indexes
psql -U postgres -d gofiber_db -c "\di"

# Check index usage
psql -U postgres -d gofiber_db -c "
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan as number_of_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
"
```

#### 📝 Test Query Performance

```sql
-- Before indexes: Get user's posts
EXPLAIN ANALYZE
SELECT * FROM posts
WHERE author_id = 'some-uuid'
ORDER BY created_at DESC
LIMIT 20;

-- After indexes: Should use idx_posts_author_created
-- Seq Scan -> Index Scan (much faster)

-- Test notification query
EXPLAIN ANALYZE
SELECT * FROM notifications
WHERE user_id = 'some-uuid' AND is_read = false
ORDER BY created_at DESC
LIMIT 50;

-- Should use idx_notifications_user_read_created

-- Test post votes query
EXPLAIN ANALYZE
SELECT * FROM posts
ORDER BY votes DESC
LIMIT 20;

-- Should use idx_posts_votes
```

### ✅ ตรวจสอบ
- [ ] Indexes ถูกสร้างครบ
- [ ] Query plans ใช้ indexes
- [ ] Query performance ดีขึ้น (10-100x)
- [ ] ไม่มี duplicate indexes

---

## 3.3 Connection Pool Optimization (Day 6)

### Checklist
- [ ] Configure connection pool size
- [ ] Set connection lifetime
- [ ] Monitor connection usage
- [ ] Test under load

### Implementation

#### 📝 Update Database Configuration

```go
// infrastructure/postgres/database.go
package postgres

import (
    "fmt"
    "log"
    "time"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "gofiber-backend/pkg/config"
)

func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Bangkok",
        cfg.DBHost,
        cfg.DBUser,
        cfg.DBPassword,
        cfg.DBName,
        cfg.DBPort,
        cfg.DBSSLMode,
    )

    // Configure GORM
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
        PrepareStmt: true, // Prepare statements for better performance
    })

    if err != nil {
        return nil, fmt.Errorf("failed to connect database: %w", err)
    }

    // Get generic database interface
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // Connection Pool Configuration
    // SetMaxIdleConns sets the maximum number of connections in the idle connection pool
    sqlDB.SetMaxIdleConns(25)

    // SetMaxOpenConns sets the maximum number of open connections to the database
    sqlDB.SetMaxOpenConns(100)

    // SetConnMaxLifetime sets the maximum amount of time a connection may be reused
    sqlDB.SetConnMaxLifetime(time.Hour)

    // SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
    sqlDB.SetConnMaxIdleTime(10 * time.Minute)

    // Ping database to verify connection
    if err := sqlDB.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Println("✓ Database connected successfully")
    log.Printf("  - Max Open Connections: 100")
    log.Printf("  - Max Idle Connections: 25")
    log.Printf("  - Connection Max Lifetime: 1 hour")
    log.Printf("  - Connection Max Idle Time: 10 minutes")

    return db, nil
}

// GetDBStats returns database connection statistics
func GetDBStats(db *gorm.DB) map[string]interface{} {
    sqlDB, _ := db.DB()
    stats := sqlDB.Stats()

    return map[string]interface{}{
        "max_open_connections": stats.MaxOpenConnections,
        "open_connections":     stats.OpenConnections,
        "in_use":               stats.InUse,
        "idle":                 stats.Idle,
        "wait_count":           stats.WaitCount,
        "wait_duration":        stats.WaitDuration,
        "max_idle_closed":      stats.MaxIdleClosed,
        "max_idle_time_closed": stats.MaxIdleTimeClosed,
        "max_lifetime_closed":  stats.MaxLifetimeClosed,
    }
}
```

#### 📝 Add Health Check Endpoint

```go
// interfaces/api/handlers/health_handler.go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "gorm.io/gorm"
    "gofiber-backend/infrastructure/postgres"
    "gofiber-backend/pkg/utils"
)

type HealthHandler struct {
    db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
    // Check database connection
    sqlDB, err := h.db.DB()
    if err != nil {
        return utils.ErrorResponse(c, fiber.Map{
            "status":   "unhealthy",
            "database": "error getting db instance",
        })
    }

    if err := sqlDB.Ping(); err != nil {
        return utils.ErrorResponse(c, fiber.Map{
            "status":   "unhealthy",
            "database": "ping failed",
        })
    }

    // Get connection stats
    stats := postgres.GetDBStats(h.db)

    return c.JSON(fiber.Map{
        "status":   "healthy",
        "database": "connected",
        "stats":    stats,
    })
}
```

### ✅ ตรวจสอบ
- [ ] Connection pool configured
- [ ] Health check shows stats
- [ ] No connection exhaustion
- [ ] Performance stable under load

---

## 3.4 Implement Caching Strategy (Day 7-9)

### Checklist
- [ ] สร้าง cache service
- [ ] Cache hot data (popular posts, user profiles)
- [ ] Implement cache invalidation
- [ ] Test cache hit rate
- [ ] Monitor cache performance

### Implementation

#### 📝 สร้าง Cache Service

```go
// pkg/cache/cache_service.go
package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
)

type CacheService struct {
    redis *redis.Client
}

func NewCacheService(redisClient *redis.Client) *CacheService {
    return &CacheService{
        redis: redisClient,
    }
}

// Get retrieves value from cache
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.redis.Get(ctx, key).Result()
    if err != nil {
        return err
    }

    return json.Unmarshal([]byte(val), dest)
}

// Set stores value in cache
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return c.redis.Set(ctx, key, data, ttl).Err()
}

// Delete removes value from cache
func (c *CacheService) Delete(ctx context.Context, keys ...string) error {
    return c.redis.Del(ctx, keys...).Err()
}

// DeletePattern deletes all keys matching pattern
func (c *CacheService) DeletePattern(ctx context.Context, pattern string) error {
    iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        if err := c.redis.Del(ctx, iter.Val()).Err(); err != nil {
            return err
        }
    }
    return iter.Err()
}

// GetOrSet retrieves from cache or executes function and caches result
func (c *CacheService) GetOrSet(ctx context.Context, key string, ttl time.Duration, dest interface{}, fn func() (interface{}, error)) error {
    // Try cache first
    err := c.Get(ctx, key, dest)
    if err == nil {
        return nil // Cache hit
    }

    if err != redis.Nil {
        // Error reading cache (not miss)
        return err
    }

    // Cache miss - execute function
    result, err := fn()
    if err != nil {
        return err
    }

    // Store in cache
    if err := c.Set(ctx, key, result, ttl); err != nil {
        // Log error but don't fail the request
        fmt.Printf("Failed to cache result: %v\n", err)
    }

    // Marshal result to dest
    data, _ := json.Marshal(result)
    return json.Unmarshal(data, dest)
}

// Exists checks if key exists
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
    count, err := c.redis.Exists(ctx, key).Result()
    return count > 0, err
}

// TTL returns time to live for key
func (c *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
    return c.redis.TTL(ctx, key).Result()
}

// Cache key builders
func PostCacheKey(postID string) string {
    return fmt.Sprintf("post:%s", postID)
}

func UserCacheKey(userID string) string {
    return fmt.Sprintf("user:%s", userID)
}

func PopularPostsCacheKey() string {
    return "posts:popular:24h"
}

func UserPostsCacheKey(userID string, page int) string {
    return fmt.Sprintf("user:%s:posts:page:%d", userID, page)
}

func PostCommentsCacheKey(postID string, page int) string {
    return fmt.Sprintf("post:%s:comments:page:%d", postID, page)
}
```

#### 📝 Use Cache in Service

```go
// application/serviceimpl/post_service_impl.go
package serviceimpl

import (
    "context"
    "time"
    "github.com/google/uuid"
    "gofiber-backend/pkg/cache"
)

type PostServiceImpl struct {
    postRepo  repositories.PostRepository
    userRepo  repositories.UserRepository
    cache     *cache.CacheService
}

func (s *PostServiceImpl) GetPostByID(ctx context.Context, postID uuid.UUID) (*dto.PostResponse, error) {
    cacheKey := cache.PostCacheKey(postID.String())

    // Try cache first
    var cachedPost dto.PostResponse
    err := s.cache.GetOrSet(ctx, cacheKey, 15*time.Minute, &cachedPost, func() (interface{}, error) {
        // Cache miss - fetch from database
        post, err := s.postRepo.FindByIDWithRelations(ctx, postID)
        if err != nil {
            return nil, apperrors.ErrPostNotFound
        }

        return dto.PostToPostResponse(post), nil
    })

    if err != nil {
        return nil, err
    }

    return &cachedPost, nil
}

func (s *PostServiceImpl) GetPopularPosts(ctx context.Context, limit int) ([]*dto.PostResponse, error) {
    cacheKey := cache.PopularPostsCacheKey()

    var posts []*dto.PostResponse
    err := s.cache.GetOrSet(ctx, cacheKey, 10*time.Minute, &posts, func() (interface{}, error) {
        // Fetch popular posts from last 24 hours
        dbPosts, err := s.postRepo.GetPopularPosts(ctx, 24*time.Hour, limit)
        if err != nil {
            return nil, err
        }

        responses := make([]*dto.PostResponse, len(dbPosts))
        for i, post := range dbPosts {
            responses[i] = dto.PostToPostResponse(post)
        }

        return responses, nil
    })

    if err != nil {
        return nil, err
    }

    return posts, nil
}

// Invalidate cache when post is updated/deleted
func (s *PostServiceImpl) UpdatePost(ctx context.Context, postID, userID uuid.UUID, dto dto.UpdatePostDTO) (*models.Post, error) {
    // ... update logic ...

    // Invalidate caches
    s.cache.Delete(ctx,
        cache.PostCacheKey(postID.String()),
        cache.PopularPostsCacheKey(),
        cache.UserPostsCacheKey(userID.String(), 1),
    )

    return post, nil
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, postID, userID uuid.UUID) error {
    // ... delete logic ...

    // Invalidate caches
    s.cache.Delete(ctx,
        cache.PostCacheKey(postID.String()),
        cache.PopularPostsCacheKey(),
    )

    // Delete all user posts cache pages
    s.cache.DeletePattern(ctx, fmt.Sprintf("user:%s:posts:page:*", userID.String()))

    return nil
}
```

#### 📝 Cache Statistics Endpoint

```go
// interfaces/api/handlers/cache_handler.go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/redis/go-redis/v9"
)

type CacheHandler struct {
    redis *redis.Client
}

func NewCacheHandler(redis *redis.Client) *CacheHandler {
    return &CacheHandler{redis: redis}
}

func (h *CacheHandler) GetCacheStats(c *fiber.Ctx) error {
    ctx := c.Context()

    // Get Redis info
    info, err := h.redis.Info(ctx, "stats").Result()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to get cache stats",
        })
    }

    // Get key count
    dbSize, err := h.redis.DBSize(ctx).Result()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to get cache size",
        })
    }

    return c.JSON(fiber.Map{
        "status":    "healthy",
        "keys":      dbSize,
        "info":      info,
    })
}

func (h *CacheHandler) ClearCache(c *fiber.Ctx) error {
    ctx := c.Context()

    // Clear all cache (use with caution!)
    if err := h.redis.FlushDB(ctx).Err(); err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Failed to clear cache",
        })
    }

    return c.JSON(fiber.Map{
        "message": "Cache cleared successfully",
    })
}
```

### ✅ ตรวจสอบ
- [ ] Cache service ทำงาน
- [ ] Hot data ถูก cache
- [ ] Cache invalidation ทำงาน
- [ ] Cache hit rate > 70%
- [ ] Response time ดีขึ้น

---

## 📊 Step 3 Completion Checklist

- [ ] ✅ Database transactions implemented
- [ ] ✅ Indexes created และ verified
- [ ] ✅ Connection pool optimized
- [ ] ✅ Caching strategy implemented
- [ ] ✅ Query performance improved
- [ ] ✅ All tests passing
- [ ] ✅ Monitoring in place

### Commit Message

```bash
git add .
git commit -m "feat: database optimization and caching

- Implement database transactions for data integrity
- Add comprehensive database indexes
- Optimize connection pool configuration
- Implement Redis caching strategy
- Add cache invalidation logic
- Create cache statistics endpoint
- Improve query performance by 10-100x

Performance: Query time reduced from seconds to milliseconds"
```

---

# STEP 4: Monitoring & Logging (Week 4-5)

**เป้าหมาย:** เพิ่มระบบ monitoring และ structured logging
**ระยะเวลา:** 7-10 วัน
**ความสำคัญ:** 🔴 CRITICAL

## ทำไมต้องทำ Step นี้?
- Logging = รู้ว่าเกิดอะไรขึ้นใน production
- Metrics = วัดประสิทธิภาพของระบบ
- Alerts = รู้ทันทีเมื่อมีปัญหา
- Debugging = แก้ bug ง่ายและเร็วขึ้น

---

## 4.1 Implement Structured Logging (Day 1-3)

### Checklist
- [ ] ติดตั้ง zerolog
- [ ] Replace log.Printf ทั้งหมด
- [ ] Add context logging
- [ ] Add log levels
- [ ] Configure log output

### Implementation

#### 📝 ติดตั้ง Zerolog

```bash
go get github.com/rs/zerolog
```

#### 📝 Setup Logger

```go
// pkg/logger/logger.go
package logger

import (
    "os"
    "time"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger
func InitLogger(env string) {
    // Set time format
    zerolog.TimeFieldFormat = time.RFC3339

    // Configure based on environment
    if env == "development" {
        // Pretty logging for development
        log.Logger = log.Output(zerolog.ConsoleWriter{
            Out:        os.Stdout,
            TimeFormat: time.RFC3339,
        })
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    } else {
        // JSON logging for production
        zerolog.SetGlobalLevel(zerolog.InfoLevel)
    }

    log.Info().
        Str("env", env).
        Msg("Logger initialized")
}

// GetLogger returns a logger with context
func GetLogger() *zerolog.Logger {
    return &log.Logger
}

// WithRequestID adds request ID to logger
func WithRequestID(requestID string) *zerolog.Logger {
    logger := log.With().Str("request_id", requestID).Logger()
    return &logger
}
```

#### 📝 Logging Middleware

```go
// pkg/middleware/logger.go
package middleware

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/rs/zerolog/log"
)

// RequestLogger logs all HTTP requests
func RequestLogger() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Generate request ID
        requestID := uuid.New().String()
        c.Locals("request_id", requestID)

        // Start timer
        start := time.Now()

        // Process request
        err := c.Next()

        // Calculate duration
        duration := time.Since(start)

        // Log request
        logEvent := log.Info().
            Str("request_id", requestID).
            Str("method", c.Method()).
            Str("path", c.Path()).
            Str("ip", c.IP()).
            Int("status", c.Response().StatusCode()).
            Dur("duration_ms", duration).
            Str("user_agent", c.Get("User-Agent"))

        // Add user ID if authenticated
        if userID := c.Locals("userID"); userID != nil {
            logEvent.Str("user_id", userID.(string))
        }

        // Log error if exists
        if err != nil {
            logEvent.Err(err).Msg("Request failed")
        } else {
            logEvent.Msg("Request completed")
        }

        return err
    }
}
```

#### 📝 Update Service Logging

```go
// application/serviceimpl/auth_service_impl.go
package serviceimpl

import (
    "context"
    "github.com/rs/zerolog/log"
)

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
    log.Info().
        Str("email", email).
        Msg("Login attempt")

    // Find user
    user, err := s.userRepo.FindByEmail(ctx, email)
    if err != nil {
        log.Warn().
            Err(err).
            Str("email", email).
            Msg("User not found")
        return nil, apperrors.ErrInvalidCredentials
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        log.Warn().
            Str("user_id", user.ID.String()).
            Str("email", email).
            Msg("Invalid password")
        return nil, apperrors.ErrInvalidCredentials
    }

    // Generate token
    token, err := s.jwtUtil.GenerateToken(user.ID)
    if err != nil {
        log.Error().
            Err(err).
            Str("user_id", user.ID.String()).
            Msg("Failed to generate token")
        return nil, apperrors.ErrInternal.WithInternal(err)
    }

    log.Info().
        Str("user_id", user.ID.String()).
        Str("email", email).
        Msg("Login successful")

    return &LoginResponse{
        Token: token,
        User:  dto.UserToUserResponse(user),
    }, nil
}
```

#### 📝 Error Logging

```go
// pkg/utils/response.go
package utils

import (
    "github.com/gofiber/fiber/v2"
    "github.com/rs/zerolog/log"
    apperrors "gofiber-backend/pkg/errors"
)

func ErrorResponse(c *fiber.Ctx, err error) error {
    requestID := c.Locals("request_id").(string)

    // Check if it's AppError
    if appErr, ok := apperrors.IsAppError(err); ok {
        // Log internal error (don't send to client)
        if appErr.Internal != nil {
            log.Error().
                Err(appErr.Internal).
                Str("request_id", requestID).
                Str("code", appErr.Code).
                Str("path", c.Path()).
                Str("method", c.Method()).
                Str("ip", c.IP()).
                Msg("Application error")
        } else {
            // Log without internal error
            log.Warn().
                Str("request_id", requestID).
                Str("code", appErr.Code).
                Str("message", appErr.Message).
                Msg("Application error")
        }

        return c.Status(appErr.StatusCode).JSON(Response{
            Success: false,
            Error: &ErrorDetail{
                Code:    appErr.Code,
                Message: appErr.Message,
                Fields:  appErr.Fields,
            },
        })
    }

    // Unknown error - log and return generic error
    log.Error().
        Err(err).
        Str("request_id", requestID).
        Str("path", c.Path()).
        Str("method", c.Method()).
        Str("ip", c.IP()).
        Msg("Unknown error")

    return c.Status(500).JSON(Response{
        Success: false,
        Error: &ErrorDetail{
            Code:    "INTERNAL_ERROR",
            Message: "An unexpected error occurred",
        },
    })
}
```

### ✅ ตรวจสอบ
- [ ] Zerolog installed และ configured
- [ ] All log.Printf replaced
- [ ] Request logging works
- [ ] Error logging works
- [ ] Log levels appropriate

---

## 4.2 Add Prometheus Metrics (Day 4-6)

### Checklist
- [ ] ติดตั้ง Prometheus client
- [ ] Add metrics middleware
- [ ] Define custom metrics
- [ ] Create metrics endpoint
- [ ] Test metrics collection

### Implementation

#### 📝 ติดตั้ง Dependencies

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/gofiber/contrib/fiberprom
```

#### 📝 Setup Prometheus Metrics

```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP Metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
        },
        []string{"method", "path"},
    )

    // Database Metrics
    DBQueriesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_queries_total",
            Help: "Total number of database queries",
        },
        []string{"operation", "table"},
    )

    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
        },
        []string{"operation", "table"},
    )

    // Cache Metrics
    CacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
    )

    CacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
    )

    // Business Metrics
    UserRegistrations = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "user_registrations_total",
            Help: "Total number of user registrations",
        },
    )

    PostsCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "posts_created_total",
            Help: "Total number of posts created",
        },
    )

    MessagesCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "messages_sent_total",
            Help: "Total number of messages sent",
        },
    )

    ActiveWebSocketConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "websocket_connections_active",
            Help: "Number of active WebSocket connections",
        },
    )

    // Error Metrics
    ErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "errors_total",
            Help: "Total number of errors",
        },
        []string{"type", "code"},
    )
)
```

#### 📝 Metrics Middleware

```go
// pkg/middleware/metrics.go
package middleware

import (
    "strconv"
    "time"
    "github.com/gofiber/fiber/v2"
    "gofiber-backend/pkg/metrics"
)

// PrometheusMiddleware collects HTTP metrics
func PrometheusMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        // Process request
        err := c.Next()

        // Record metrics
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Response().StatusCode())
        method := c.Method()
        path := c.Path()

        metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
        metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

        // Record errors
        if err != nil {
            metrics.ErrorsTotal.WithLabelValues("http", status).Inc()
        }

        return err
    }
}
```

#### 📝 Update Cache Service with Metrics

```go
// pkg/cache/cache_service.go
package cache

import (
    "gofiber-backend/pkg/metrics"
)

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.redis.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            metrics.CacheMisses.Inc()
        }
        return err
    }

    metrics.CacheHits.Inc()
    return json.Unmarshal([]byte(val), dest)
}
```

#### 📝 Metrics Endpoint

```go
// cmd/api/main.go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "gofiber-backend/pkg/middleware"
)

func main() {
    app := fiber.New()

    // Metrics middleware
    app.Use(middleware.PrometheusMiddleware())

    // Metrics endpoint (should be protected or internal only)
    app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

    // ... rest of setup
}
```

### ✅ ตรวจสอบ
- [ ] Metrics endpoint accessible (`/metrics`)
- [ ] HTTP metrics collected
- [ ] Custom metrics work
- [ ] Prometheus can scrape metrics

---

## 4.3 Setup Grafana Dashboards (Day 7-9)

### Checklist
- [ ] Install Prometheus
- [ ] Install Grafana
- [ ] Configure data source
- [ ] Create dashboards
- [ ] Setup alerts

### Implementation

#### 📝 Docker Compose for Monitoring

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3001:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    restart: unless-stopped
    depends_on:
      - prometheus

volumes:
  prometheus-data:
  grafana-data:
```

#### 📝 Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'gofiber-backend'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'
```

#### 📝 Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "GoFiber Backend Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Request Duration",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(errors_total[5m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Cache Hit Rate",
        "targets": [
          {
            "expr": "cache_hits_total / (cache_hits_total + cache_misses_total)"
          }
        ],
        "type": "gauge"
      }
    ]
  }
}
```

#### 📝 Start Monitoring Stack

```bash
# Start Prometheus and Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# Access Grafana
# http://localhost:3001
# Username: admin
# Password: admin

# Access Prometheus
# http://localhost:9090
```

### ✅ ตรวจสอบ
- [ ] Prometheus collecting metrics
- [ ] Grafana dashboards created
- [ ] Metrics visualized correctly
- [ ] Alerts configured

---

## 📊 Step 4 Completion Checklist

- [ ] ✅ Structured logging implemented
- [ ] ✅ Prometheus metrics collecting
- [ ] ✅ Grafana dashboards created
- [ ] ✅ Alerts configured
- [ ] ✅ Request ID tracking
- [ ] ✅ Error tracking works
- [ ] ✅ All tests passing

### Commit Message

```bash
git add .
git commit -m "feat: add comprehensive monitoring and logging

- Implement structured logging with zerolog
- Add Prometheus metrics collection
- Create Grafana dashboards
- Setup HTTP request logging
- Add custom business metrics
- Configure Prometheus alerts
- Add metrics middleware
- Implement request ID tracking

Observability: Full visibility into production system"
```

---

# STEP 5: Production Readiness (Week 5-6)

**เป้าหมาย:** เตรียมพร้อมสำหรับ production deployment
**ระยะเวลา:** 7-10 วัน
**ความสำคัญ:** 🔴 CRITICAL

## ทำไมต้องทำ Step นี้?
- CI/CD = deploy อัตโนมัติและปลอดภัย
- API Documentation = ง่ายต่อการใช้งาน
- Load testing = รู้ว่าระบบรับ load เท่าไหร่ได้
- Security audit = ปิดช่องโหว่

---

## 5.1 Create API Documentation (Day 1-2)

### Checklist
- [ ] Install Swagger/OpenAPI
- [ ] Document all endpoints
- [ ] Add request/response examples
- [ ] Generate API docs
- [ ] Deploy docs

### Implementation

#### 📝 Install Swagger

```bash
go get github.com/swaggo/swag/cmd/swag
go get github.com/gofiber/swagger
go get github.com/swaggo/files
```

#### 📝 Add Swagger Annotations

```go
// cmd/api/main.go

// @title GoFiber Social Media API
// @version 1.0
// @description Social media platform backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@yourdomain.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
    // ...
}
```

```go
// interfaces/api/handlers/auth_handler.go

// Login godoc
// @Summary User login
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginDTO true "Login credentials"
// @Success 200 {object} utils.Response{data=dto.LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
    // ...
}

// Register godoc
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterDTO true "User registration data"
// @Success 201 {object} utils.Response{data=dto.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
    // ...
}
```

#### 📝 Generate Swagger Docs

```bash
# Generate docs
swag init -g cmd/api/main.go -o docs

# This creates:
# docs/docs.go
# docs/swagger.json
# docs/swagger.yaml
```

#### 📝 Add Swagger UI Route

```go
// cmd/api/main.go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/swagger"
    _ "gofiber-backend/docs" // Import generated docs
)

func main() {
    app := fiber.New()

    // Swagger UI
    app.Get("/swagger/*", swagger.HandlerDefault)

    // ... rest of setup
}
```

### ✅ ตรวจสอบ
- [ ] Swagger docs generated
- [ ] Swagger UI accessible (`/swagger/index.html`)
- [ ] All endpoints documented
- [ ] Examples clear and accurate

---

## 5.2 Setup CI/CD Pipeline (Day 3-5)

### Checklist
- [ ] Create GitHub Actions workflow
- [ ] Add automated testing
- [ ] Add linting
- [ ] Add security scanning
- [ ] Configure deployment

### Implementation

#### 📝 GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.21'

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: gofiber_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install dependencies
        run: go mod download

      - name: Run tests
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: postgres
          DB_NAME: gofiber_test
          REDIS_HOST: localhost
          REDIS_PORT: 6379
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  security:
    name: Security Scan
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: ./...

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test, lint, security]

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Build
        run: |
          go build -v -o bin/server cmd/api/main.go

      - name: Build Docker image
        run: |
          docker build -t gofiber-backend:${{ github.sha }} .

  deploy:
    name: Deploy
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main'

    steps:
      - name: Deploy to production
        run: |
          echo "Deploy to production server"
          # Add your deployment commands here
```

#### 📝 Linting Configuration

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - goconst
    - gosec

linters-settings:
  errcheck:
    check-type-assertions: true
  govet:
    check-shadowing: true

issues:
  exclude-use-default: false
```

### ✅ ตรวจสอบ
- [ ] CI pipeline runs on push
- [ ] Tests run automatically
- [ ] Linting passes
- [ ] Security scanning passes
- [ ] Build succeeds

---

## 5.3 Load Testing (Day 6-7)

### Checklist
- [ ] Install load testing tool (k6)
- [ ] Create test scenarios
- [ ] Run load tests
- [ ] Analyze results
- [ ] Optimize bottlenecks

### Implementation

#### 📝 Install k6

```bash
# macOS
brew install k6

# Linux
sudo apt-get install k6

# Windows
choco install k6
```

#### 📝 Create Load Test Script

```javascript
// loadtests/api-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
    stages: [
        { duration: '1m', target: 50 },   // Ramp up to 50 users
        { duration: '3m', target: 50 },   // Stay at 50 users
        { duration: '1m', target: 100 },  // Ramp up to 100 users
        { duration: '3m', target: 100 },  // Stay at 100 users
        { duration: '1m', target: 200 },  // Ramp up to 200 users
        { duration: '3m', target: 200 },  // Stay at 200 users
        { duration: '2m', target: 0 },    // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
        errors: ['rate<0.1'],              // Error rate should be less than 10%
    },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export function setup() {
    // Register test user
    const registerRes = http.post(`${BASE_URL}/auth/register`, JSON.stringify({
        email: `loadtest-${Date.now()}@example.com`,
        username: `loadtest${Date.now()}`,
        password: 'Password123!',
        full_name: 'Load Test User',
    }), {
        headers: { 'Content-Type': 'application/json' },
    });

    const token = JSON.parse(registerRes.body).data.token;
    return { token };
}

export default function(data) {
    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
    };

    // Test 1: Get popular posts
    let res = http.get(`${BASE_URL}/posts/popular`, { headers });
    check(res, {
        'get popular posts status 200': (r) => r.status === 200,
        'get popular posts duration < 500ms': (r) => r.timings.duration < 500,
    }) || errorRate.add(1);

    sleep(1);

    // Test 2: Create post
    res = http.post(`${BASE_URL}/posts`, JSON.stringify({
        title: 'Load test post',
        content: 'This is a load test post',
    }), { headers });
    check(res, {
        'create post status 201': (r) => r.status === 201,
    }) || errorRate.add(1);

    sleep(1);

    // Test 3: Get user profile
    res = http.get(`${BASE_URL}/users/me`, { headers });
    check(res, {
        'get profile status 200': (r) => r.status === 200,
    }) || errorRate.add(1);

    sleep(2);
}
```

#### 📝 Run Load Tests

```bash
# Run load test
k6 run loadtests/api-test.js

# Run with custom VUs and duration
k6 run --vus 100 --duration 5m loadtests/api-test.js

# Run with Cloud output
k6 run --out cloud loadtests/api-test.js
```

#### 📝 Analyze Results

```bash
# Output example:
#     ✓ get popular posts status 200
#     ✓ get popular posts duration < 500ms
#     ✓ create post status 201
#     ✓ get profile status 200
#
#     checks.........................: 100.00% ✓ 20000  ✗ 0
#     data_received..................: 15 MB   250 kB/s
#     data_sent......................: 7.5 MB  125 kB/s
#     http_req_blocked...............: avg=1.2ms    min=0s       med=1ms     max=50ms
#     http_req_connecting............: avg=800µs    min=0s       med=700µs   max=30ms
#     http_req_duration..............: avg=250ms    min=50ms     med=200ms   max=800ms
#     http_req_receiving.............: avg=150µs    min=50µs     med=100µs   max=5ms
#     http_req_sending...............: avg=100µs    min=30µs     med=80µs    max=3ms
#     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s
#     http_req_waiting...............: avg=249ms    min=49ms     med=199ms   max=799ms
#     http_reqs......................: 20000   333.33/s
#     iteration_duration.............: avg=5s       min=4s       med=5s      max=7s
#     iterations.....................: 5000    83.33/s
#     vus............................: 100     min=50   max=200
#     vus_max........................: 200     min=200  max=200
```

### ✅ ตรวจสอบ
- [ ] Load tests run successfully
- [ ] 95th percentile < 500ms
- [ ] Error rate < 5%
- [ ] System handles 200+ concurrent users
- [ ] No memory leaks

---

## 5.4 Security Audit (Day 8-10)

### Checklist
- [ ] Run security scanners
- [ ] Fix vulnerabilities
- [ ] Penetration testing
- [ ] Update dependencies
- [ ] Document security measures

### Implementation

#### 📝 Run Security Tools

```bash
# 1. Gosec - Go security checker
gosec -fmt=json -out=security-report.json ./...

# 2. Nancy - Check for vulnerabilities in dependencies
go list -json -m all | nancy sleuth

# 3. Snyk - Vulnerability scanning
snyk test

# 4. OWASP Dependency-Check
dependency-check --project "GoFiber Backend" --scan .

# 5. Trivy - Container security
trivy image gofiber-backend:latest
```

#### 📝 Security Checklist

```markdown
# Security Checklist

## Authentication & Authorization
- [x] JWT tokens with secure secret
- [x] Password hashing with bcrypt
- [x] Rate limiting on auth endpoints
- [x] Account lockout after failed attempts
- [ ] Two-factor authentication (optional)
- [ ] OAuth2 security best practices

## Input Validation
- [x] All user inputs validated
- [x] XSS protection
- [x] SQL injection prevention (using ORM)
- [x] File upload validation
- [x] Request size limits

## Data Protection
- [x] Passwords hashed (never stored plain text)
- [x] Sensitive data encrypted
- [x] HTTPS enforced (production)
- [ ] Database encryption at rest
- [ ] Backup encryption

## Security Headers
- [x] CSP (Content Security Policy)
- [x] HSTS (HTTP Strict Transport Security)
- [x] X-Frame-Options
- [x] X-Content-Type-Options
- [x] X-XSS-Protection

## API Security
- [x] Rate limiting
- [x] CORS configured properly
- [x] API versioning
- [x] Error messages don't leak info
- [ ] API key rotation

## Infrastructure
- [x] Containers run as non-root user
- [x] Secrets in environment variables
- [ ] Secrets manager integration
- [ ] Network segmentation
- [ ] Firewall configured

## Monitoring & Logging
- [x] Security events logged
- [x] Failed login attempts logged
- [ ] Intrusion detection system
- [ ] Log analysis/SIEM integration

## Compliance
- [ ] GDPR compliance (if applicable)
- [ ] Data retention policy
- [ ] Privacy policy
- [ ] Terms of service
```

### ✅ ตรวจสอบ
- [ ] No critical vulnerabilities
- [ ] All dependencies up to date
- [ ] Security audit passed
- [ ] Penetration test passed
- [ ] Security documentation complete

---

## 📊 Step 5 Completion Checklist

- [ ] ✅ API documentation complete
- [ ] ✅ CI/CD pipeline working
- [ ] ✅ Load tests passing
- [ ] ✅ Security audit clean
- [ ] ✅ All tests passing
- [ ] ✅ Ready for production

### Commit Message

```bash
git add .
git commit -m "feat: production readiness complete

- Add comprehensive API documentation (Swagger)
- Setup CI/CD pipeline with GitHub Actions
- Implement load testing with k6
- Complete security audit
- Add automated testing and linting
- Configure deployment automation
- Document all endpoints

Production: System ready for deployment"
```

---

# STEP 6: Optimization & Polish (Week 6-8)

**เป้าหมาย:** ปรับแต่งและเพิ่มประสิทธิภาพสูงสุด
**ระยะเวลา:** 7-14 วัน
**ความสำคัญ:** 🟡 MEDIUM

## ทำไมต้องทำ Step นี้?
- Response compression = ลดข้อมูลที่ส่ง 70-80%
- Query optimization = ลดเวลา response
- Image optimization = ประหยัด bandwidth
- Code cleanup = ง่ายต่อการ maintain

---

## 6.1 Response Compression (Day 1-2)

### Implementation

```go
// pkg/middleware/compression.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/compress"
)

func Compression() fiber.Handler {
    return compress.New(compress.Config{
        Level: compress.LevelBestSpeed, // or LevelBestCompression
    })
}
```

```go
// cmd/api/main.go
app.Use(middleware.Compression())
```

---

## 6.2 Database Query Optimization

### Implementation

```go
// Optimize N+1 queries
// Before:
posts, _ := postRepo.GetAll(ctx)
for _, post := range posts {
    author, _ := userRepo.FindByID(ctx, post.AuthorID)
    // N+1 problem!
}

// After:
posts, _ := postRepo.GetAllWithAuthor(ctx)
// Single query with JOIN or Preload

// GORM Preload
db.Preload("Author").
   Preload("Tags").
   Preload("Media").
   Find(&posts)
```

---

## 6.3 Pagination Optimization

```go
// Add cursor-based pagination for better performance
type CursorPaginationParams struct {
    Cursor string
    Limit  int
}

func (r *PostRepository) GetPostsCursor(ctx context.Context, params CursorPaginationParams) ([]*Post, string, error) {
    query := r.db.WithContext(ctx).
        Where("created_at < ?", decodeCursor(params.Cursor)).
        Order("created_at DESC").
        Limit(params.Limit + 1)

    var posts []*Post
    if err := query.Find(&posts).Error; err != nil {
        return nil, "", err
    }

    var nextCursor string
    if len(posts) > params.Limit {
        nextCursor = encodeCursor(posts[params.Limit].CreatedAt)
        posts = posts[:params.Limit]
    }

    return posts, nextCursor, nil
}
```

---

## 6.4 Code Cleanup & Refactoring

### Tasks
- [ ] Remove dead code
- [ ] Fix linting issues
- [ ] Improve code comments
- [ ] Extract common logic
- [ ] Rename unclear variables

---

## 📊 Step 6 Completion Checklist

- [ ] ✅ Response compression enabled
- [ ] ✅ Database queries optimized
- [ ] ✅ Pagination optimized
- [ ] ✅ Code cleaned up
- [ ] ✅ Performance benchmarks improved
- [ ] ✅ All tests still passing

### Final Commit

```bash
git add .
git commit -m "feat: final optimizations and polish

- Enable response compression (gzip)
- Optimize database queries (eliminate N+1)
- Implement cursor-based pagination
- Clean up code and improve readability
- Add performance benchmarks
- Final polish for production

Performance: 30-50% improvement in response times"
```

---

# 🎉 FINAL CHECKLIST - Production Ready!

## ก่อน Deploy Production ต้องเช็คให้ครบ:

### Core Functionality ✅
- [x] All features working
- [x] No critical bugs
- [x] Error handling complete
- [x] Input validation complete

### Testing ✅
- [x] Unit tests >= 70% coverage
- [x] Integration tests passing
- [x] Load tests passing
- [x] Manual testing complete

### Security ✅
- [x] Rate limiting enabled
- [x] Security headers configured
- [x] CORS configured properly
- [x] No vulnerabilities
- [x] Secrets secure
- [x] HTTPS enforced

### Performance ✅
- [x] Database indexed
- [x] Caching implemented
- [x] Connection pooling optimized
- [x] Response time < 500ms (p95)
- [x] Compression enabled

### Monitoring ✅
- [x] Structured logging
- [x] Metrics collecting
- [x] Dashboards created
- [x] Alerts configured

### Documentation ✅
- [x] API documentation (Swagger)
- [x] README updated
- [x] Deployment guide
- [x] Environment variables documented

### DevOps ✅
- [x] CI/CD pipeline
- [x] Docker images
- [x] Health checks
- [x] Graceful shutdown
- [x] Backup strategy

---

# 📈 Progress Tracking

## How to Use This Roadmap

1. **ห้ามข้าม Step** - ทำตามลำดับเท่านั้น
2. **Tick checkboxes** - เมื่อทำเสร็จแต่ละรายการ
3. **Commit บ่อยๆ** - ทุก sub-step ที่เสร็จ
4. **Test ทุกอย่าง** - ก่อนไป step ถัดไป
5. **ถาม help** - ถ้าติดปัญหา

## Timeline Summary

| Step | Duration | Priority | Status |
|------|----------|----------|--------|
| Step 1: Foundation | 10-14 days | 🔴 CRITICAL | ⬜ Not Started |
| Step 2: Security | 7-10 days | 🔴 CRITICAL | ⬜ Not Started |
| Step 3: Database | 7-10 days | 🟠 HIGH | ⬜ Not Started |
| Step 4: Monitoring | 7-10 days | 🔴 CRITICAL | ⬜ Not Started |
| Step 5: Production | 7-10 days | 🔴 CRITICAL | ⬜ Not Started |
| Step 6: Optimization | 7-14 days | 🟡 MEDIUM | ⬜ Not Started |

**Total: 6-8 weeks**

---

# 🎯 Next Steps

## เริ่มจาก Step 1 เลย!

1. สร้าง branch ใหม่: `git checkout -b feature/step-1-foundation`
2. เริ่มทำ Step 1.1: สร้าง Error Handling System
3. Test ให้แน่ใจว่าทำงาน
4. Commit: `git commit -m "feat: add error handling system"`
5. ทำต่อไปจนจบ Step 1
6. Merge เข้า main เมื่อ Step 1 เสร็จ 100%

---

**Good luck! 🚀**

หากมีคำถามหรือติดปัญหาตรงไหน สามารถถามได้เลยครับ!
