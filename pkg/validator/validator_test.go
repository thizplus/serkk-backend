package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestUser struct {
	Email    string `validate:"required,email"`
	Username string `validate:"required,username"`
	Password string `validate:"required,password"`
	Age      int    `validate:"required,gte=18,lte=100"`
}

type TestPost struct {
	Title   string   `validate:"required,min=3,max=200"`
	Content string   `validate:"required,min=10"`
	Tags    []string `validate:"max=5"`
}

func TestValidator_ValidStruct(t *testing.T) {
	v := New()

	user := TestUser{
		Email:    "test@example.com",
		Username: "testuser123",
		Password: "password123",
		Age:      25,
	}

	errors := v.Validate(user)

	assert.Nil(t, errors)
}

func TestValidator_RequiredField(t *testing.T) {
	v := New()

	user := TestUser{
		Email:    "",
		Username: "testuser",
		Password: "password123",
		Age:      25,
	}

	errors := v.Validate(user)

	assert.NotNil(t, errors)
	assert.Len(t, errors, 1)
	assert.Equal(t, "email", errors[0].Field)
	assert.Equal(t, "required", errors[0].Tag)
	assert.Contains(t, errors[0].Message, "required")
}

func TestValidator_EmailValidation(t *testing.T) {
	v := New()

	user := TestUser{
		Email:    "invalid-email",
		Username: "testuser",
		Password: "password123",
		Age:      25,
	}

	errors := v.Validate(user)

	assert.NotNil(t, errors)
	assert.Len(t, errors, 1)
	assert.Equal(t, "email", errors[0].Field)
	assert.Equal(t, "email", errors[0].Tag)
	assert.Contains(t, errors[0].Message, "valid email")
}

func TestValidator_UsernameValidation(t *testing.T) {
	v := New()

	// Invalid username (too short)
	user1 := TestUser{
		Email:    "test@example.com",
		Username: "ab",
		Password: "password123",
		Age:      25,
	}

	errors := v.Validate(user1)
	assert.NotNil(t, errors)

	// Invalid username (special characters)
	user2 := TestUser{
		Email:    "test@example.com",
		Username: "test@user",
		Password: "password123",
		Age:      25,
	}

	errors = v.Validate(user2)
	assert.NotNil(t, errors)

	// Valid username
	user3 := TestUser{
		Email:    "test@example.com",
		Username: "test_user-123",
		Password: "password123",
		Age:      25,
	}

	errors = v.Validate(user3)
	assert.Nil(t, errors)
}

func TestValidator_PasswordValidation(t *testing.T) {
	v := New()

	// Too short
	user1 := TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "pass",
		Age:      25,
	}

	errors := v.Validate(user1)
	assert.NotNil(t, errors)

	// No number
	user2 := TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password",
		Age:      25,
	}

	errors = v.Validate(user2)
	assert.NotNil(t, errors)

	// No letter
	user3 := TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "12345678",
		Age:      25,
	}

	errors = v.Validate(user3)
	assert.NotNil(t, errors)

	// Valid password
	user4 := TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Age:      25,
	}

	errors = v.Validate(user4)
	assert.Nil(t, errors)
}

func TestValidator_MinMaxValidation(t *testing.T) {
	v := New()

	// Title too short
	post1 := TestPost{
		Title:   "Hi",
		Content: "This is a test content that is long enough",
		Tags:    []string{"tag1"},
	}

	errors := v.Validate(post1)
	assert.NotNil(t, errors)
	assert.Contains(t, errors[0].Message, "at least")

	// Content too short
	post2 := TestPost{
		Title:   "Valid Title",
		Content: "Short",
		Tags:    []string{"tag1"},
	}

	errors = v.Validate(post2)
	assert.NotNil(t, errors)

	// Valid post
	post3 := TestPost{
		Title:   "Valid Title",
		Content: "This is a test content that is long enough",
		Tags:    []string{"tag1", "tag2"},
	}

	errors = v.Validate(post3)
	assert.Nil(t, errors)
}

func TestValidator_AgeValidation(t *testing.T) {
	v := New()

	// Too young
	user1 := TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Age:      15,
	}

	errors := v.Validate(user1)
	assert.NotNil(t, errors)
	assert.Contains(t, errors[0].Message, "greater than or equal")

	// Too old
	user2 := TestUser{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		Age:      150,
	}

	errors = v.Validate(user2)
	assert.NotNil(t, errors)
	assert.Contains(t, errors[0].Message, "less than or equal")
}

func TestValidator_MultipleErrors(t *testing.T) {
	v := New()

	user := TestUser{
		Email:    "invalid",
		Username: "ab",
		Password: "short",
		Age:      10,
	}

	errors := v.Validate(user)

	assert.NotNil(t, errors)
	assert.GreaterOrEqual(t, len(errors), 3) // At least 3 validation errors
}

func TestValidator_ValidateVar(t *testing.T) {
	v := New()

	// Valid email
	err := v.ValidateVar("test@example.com", "email")
	assert.NoError(t, err)

	// Invalid email
	err = v.ValidateVar("invalid-email", "email")
	assert.Error(t, err)

	// Required field
	err = v.ValidateVar("", "required")
	assert.Error(t, err)

	err = v.ValidateVar("value", "required")
	assert.NoError(t, err)
}
