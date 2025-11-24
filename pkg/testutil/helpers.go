package testutil

import (
	"context"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateTestUser creates a test user (now UsersIdentity)
// Note: User is now an alias to UsersIdentity
// For full user data with profile, create both UsersIdentity and UserProfile
func CreateTestUser() *models.User {
	return &models.User{
		ID:        uuid.New(),
		Email:     faker.Email(),
		Username:  faker.Username(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// Note: Password, DisplayName, Bio, Avatar, IsActive, Role no longer exist
		// They're either in Auth Service (Password, Role, IsActive) or UserProfile (DisplayName, Bio, Avatar)
	}
}

// CreateTestUserWithData creates a test user with specific data (now UsersIdentity)
// Note: User is now an alias to UsersIdentity
// For full user data with profile, create both UsersIdentity and UserProfile
func CreateTestUserWithData(email, username, password string) *models.User {
	// Note: password is no longer stored in backend (managed by Auth Service)
	return &models.User{
		ID:        uuid.New(),
		Email:     email,
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// Note: Password, DisplayName, Bio, Avatar, IsActive, Role no longer exist
		// They're either in Auth Service (Password, Role, IsActive) or UserProfile (DisplayName, Bio, Avatar)
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
		Status:    "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestPostWithData creates a test post with specific data
func CreateTestPostWithData(authorID uuid.UUID, title, content string) *models.Post {
	return &models.Post{
		ID:        uuid.New(),
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		Votes:     0,
		Status:    "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestComment creates a test comment
func CreateTestComment(postID, authorID uuid.UUID) *models.Comment {
	return &models.Comment{
		ID:        uuid.New(),
		PostID:    postID,
		AuthorID:  authorID,
		Content:   faker.Sentence(),
		Votes:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestTag creates a test tag
func CreateTestTag() *models.Tag {
	return &models.Tag{
		ID:        uuid.New(),
		Name:      faker.Word(),
		CreatedAt: time.Now(),
	}
}

// CreateTestConversation creates a test conversation
func CreateTestConversation(user1ID, user2ID uuid.UUID) *models.Conversation {
	return &models.Conversation{
		ID:            uuid.New(),
		LastMessageAt: time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// CreateTestMessage creates a test message
func CreateTestMessage(conversationID, senderID, receiverID uuid.UUID) *models.Message {
	content := faker.Sentence()
	return &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Type:           models.MessageTypeText,
		Content:        &content,
		IsRead:         false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// CreateTestNotification creates a test notification
func CreateTestNotification(userID, senderID uuid.UUID) *models.Notification {
	postID := uuid.New()
	return &models.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		SenderID:  senderID,
		Type:      "vote",
		Message:   "Someone voted on your post",
		PostID:    &postID,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
}

// CreateTestMedia creates a test media
func CreateTestMedia(userID uuid.UUID, mediaType string) *models.Media {
	return &models.Media{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      mediaType,
		FileName:  "test-media.jpg",
		Extension: "jpg",
		MimeType:  "image/jpeg",
		Size:      1024000,
		URL:       "https://example.com/media.jpg",
		Thumbnail: "https://example.com/media-thumb.jpg",
		Width:     1920,
		Height:    1080,
		CreatedAt: time.Now(),
	}
}

// CreateTestContext creates test context with timeout
func CreateTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// CreateTestContextWithValue creates test context with value
func CreateTestContextWithValue(key, value interface{}) context.Context {
	return context.WithValue(context.Background(), key, value)
}
