package testutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateTestUser(t *testing.T) {
	// Act
	user := CreateTestUser()

	// Assert
	assert.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotEmpty(t, user.Email)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Password)
	assert.NotEmpty(t, user.DisplayName)
}

func TestCreateTestUserWithData(t *testing.T) {
	// Arrange
	email := "test@example.com"
	username := "testuser"
	password := "password123"

	// Act
	user := CreateTestUserWithData(email, username, password)

	// Assert
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, username, user.Username)
	assert.NotEmpty(t, user.Password)
}

func TestCreateTestPost(t *testing.T) {
	// Arrange
	authorID := uuid.New()

	// Act
	post := CreateTestPost(authorID)

	// Assert
	assert.NotNil(t, post)
	assert.Equal(t, authorID, post.AuthorID)
	assert.NotEmpty(t, post.Title)
	assert.NotEmpty(t, post.Content)
	assert.Equal(t, "published", post.Status)
}

func TestCreateTestPostWithData(t *testing.T) {
	// Arrange
	authorID := uuid.New()
	title := "Test Title"
	content := "Test Content"

	// Act
	post := CreateTestPostWithData(authorID, title, content)

	// Assert
	assert.NotNil(t, post)
	assert.Equal(t, authorID, post.AuthorID)
	assert.Equal(t, title, post.Title)
	assert.Equal(t, content, post.Content)
}

func TestCreateTestContext(t *testing.T) {
	// Act
	ctx, cancel := CreateTestContext()
	defer cancel()

	// Assert
	assert.NotNil(t, ctx)
	_, ok := ctx.Deadline()
	assert.True(t, ok) // Context has timeout
}
