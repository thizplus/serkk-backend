package serviceimpl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories/mocks"
	"gofiber-template/pkg/testutil"
)

func setupUserService() (*UserServiceImpl, *mocks.MockUserRepository, *mocks.MockFollowRepository) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockFollowRepo := new(mocks.MockFollowRepository)
	jwtSecret := "test-secret-key"

	service := &UserServiceImpl{
		userRepo:   mockUserRepo,
		followRepo: mockFollowRepo,
		jwtSecret:  jwtSecret,
	}

	return service, mockUserRepo, mockFollowRepo
}

func TestRegister_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	req := &dto.CreateUserRequest{
		Email:       "test@example.com",
		Username:    "testuser",
		Password:    "password123",
		DisplayName: "Test User",
	}

	// Mock: email and username don't exist
	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("not found"))
	mockUserRepo.On("GetByUsername", ctx, req.Username).Return(nil, errors.New("not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	user, err := service.Register(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Username, user.Username)
	assert.Equal(t, req.DisplayName, user.DisplayName)
	assert.Equal(t, "user", user.Role)
	assert.True(t, user.IsActive)
	assert.NotEmpty(t, user.Password)
	mockUserRepo.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	existingUser := testutil.CreateTestUser()
	req := &dto.CreateUserRequest{
		Email:       existingUser.Email,
		Username:    "newuser",
		Password:    "password123",
		DisplayName: "New User",
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	// Act
	user, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email already exists", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	existingUser := testutil.CreateTestUser()
	req := &dto.CreateUserRequest{
		Email:       "new@example.com",
		Username:    existingUser.Username,
		Password:    "password123",
		DisplayName: "New User",
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("not found"))
	mockUserRepo.On("GetByUsername", ctx, req.Username).Return(existingUser, nil)

	// Act
	user, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "username already exists", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Password: string(hashedPassword),
		IsActive: true,
		Role:     "user",
	}

	req := &dto.LoginRequest{
		Email:    user.Email,
		Password: password,
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	// Act
	token, returnedUser, err := service.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, returnedUser)
	assert.Equal(t, user.Email, returnedUser.Email)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_InvalidEmail(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	req := &dto.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("not found"))

	// Act
	token, user, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Equal(t, "invalid email or password", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: string(hashedPassword),
		IsActive: true,
	}

	req := &dto.LoginRequest{
		Email:    user.Email,
		Password: "wrongpassword",
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	// Act
	token, returnedUser, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, returnedUser)
	assert.Equal(t, "invalid email or password", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_InactiveAccount(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
		IsActive: false,
	}

	req := &dto.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	}

	mockUserRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	// Act
	token, returnedUser, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, returnedUser)
	assert.Equal(t, "account is disabled", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	user := testutil.CreateTestUser()
	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	// Act
	result, err := service.GetProfile(ctx, user.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	mockUserRepo.AssertExpectations(t)
}

func TestGetProfile_UserNotFound(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	userID := uuid.New()
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, errors.New("not found"))

	// Act
	result, err := service.GetProfile(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestUpdateProfile_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	user := testutil.CreateTestUser()
	req := &dto.UpdateUserRequest{
		DisplayName: "Updated Name",
		Bio:         "Updated bio",
		Location:    "Bangkok",
		Website:     "https://example.com",
		Avatar:      "https://example.com/avatar.jpg",
	}

	mockUserRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	mockUserRepo.On("Update", ctx, user.ID, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	result, err := service.UpdateProfile(ctx, user.ID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.DisplayName, result.DisplayName)
	assert.Equal(t, req.Bio, result.Bio)
	assert.Equal(t, req.Location, result.Location)
	assert.Equal(t, req.Website, result.Website)
	assert.Equal(t, req.Avatar, result.Avatar)
	mockUserRepo.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	userID := uuid.New()
	mockUserRepo.On("Delete", ctx, userID).Return(nil)

	// Act
	err := service.DeleteUser(ctx, userID)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestListUsers_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	ctx := context.Background()

	users := []*models.User{
		testutil.CreateTestUser(),
		testutil.CreateTestUser(),
	}
	offset, limit := 0, 10
	totalCount := int64(2)

	mockUserRepo.On("List", ctx, offset, limit).Return(users, nil)
	mockUserRepo.On("Count", ctx).Return(totalCount, nil)

	// Act
	result, count, err := service.ListUsers(ctx, offset, limit)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, totalCount, count)
	mockUserRepo.AssertExpectations(t)
}

func TestGenerateJWT_Success(t *testing.T) {
	// Arrange
	service, _, _ := setupUserService()
	user := testutil.CreateTestUser()

	// Act
	token, err := service.GenerateJWT(user)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.jwtSecret), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, user.ID.String(), claims["user_id"])
	assert.Equal(t, user.Username, claims["username"])
	assert.Equal(t, user.Email, claims["email"])
	assert.Equal(t, user.Role, claims["role"])
}

func TestValidateJWT_Success(t *testing.T) {
	// Arrange
	service, mockUserRepo, _ := setupUserService()
	user := testutil.CreateTestUser()

	// Generate a valid token
	token, _ := service.GenerateJWT(user)

	mockUserRepo.On("GetByID", mock.Anything, user.ID).Return(user, nil)

	// Act
	result, err := service.ValidateJWT(token)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	mockUserRepo.AssertExpectations(t)
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	// Arrange
	service, _, _ := setupUserService()
	invalidToken := "invalid.token.string"

	// Act
	result, err := service.ValidateJWT(invalidToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	// Arrange
	service, _, _ := setupUserService()
	user := testutil.CreateTestUser()

	// Generate an expired token
	claims := jwt.MapClaims{
		"user_id":  user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(service.jwtSecret))

	// Act
	result, err := service.ValidateJWT(tokenString)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}
