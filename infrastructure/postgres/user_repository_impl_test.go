// +build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gofiber-template/pkg/config"
	"gofiber-template/pkg/testutil"
)

func setupUserRepoTest(t *testing.T) (*UserRepositoryImpl, func()) {
	// Get test database
	db, err := config.GetTestDB()
	require.NoError(t, err, "Failed to connect to test database")

	// Create repository
	repo := &UserRepositoryImpl{db: db}

	// Cleanup function
	cleanup := func() {
		err := config.CleanTestDB(db)
		assert.NoError(t, err, "Failed to clean test database")
	}

	return repo, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := testutil.CreateTestUser()

	// Act
	err := repo.Create(ctx, user)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)

	// Verify user was created
	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Username, found.Username)
}

func TestUserRepository_GetByID(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := testutil.CreateTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Act
	found, err := repo.GetByID(ctx, user.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Username, found.Username)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	nonExistentID := uuid.New()

	// Act
	found, err := repo.GetByID(ctx, nonExistentID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := testutil.CreateTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Act
	found, err := repo.GetByEmail(ctx, user.Email)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Act
	found, err := repo.GetByEmail(ctx, "notfound@example.com")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := testutil.CreateTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Act
	found, err := repo.GetByUsername(ctx, user.Username)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Username, found.Username)
}

func TestUserRepository_Update(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := testutil.CreateTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Modify user
	user.DisplayName = "Updated Name"
	user.Bio = "Updated bio"

	// Act
	err = repo.Update(ctx, user.ID, user)

	// Assert
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.DisplayName)
	assert.Equal(t, "Updated bio", found.Bio)
}

func TestUserRepository_Delete(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user := testutil.CreateTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Act
	err = repo.Delete(ctx, user.ID)

	// Assert
	assert.NoError(t, err)

	// Verify deletion
	found, err := repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestUserRepository_List(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 5; i++ {
		user := testutil.CreateTestUser()
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	// Act
	users, err := repo.List(ctx, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 5)
}

func TestUserRepository_Count(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 3; i++ {
		user := testutil.CreateTestUser()
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	// Act
	count, err := repo.Count(ctx)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))
}

func TestUserRepository_CreateDuplicate_Email(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user1 := testutil.CreateTestUser()
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create another user with same email
	user2 := testutil.CreateTestUser()
	user2.Email = user1.Email // Same email

	// Act
	err = repo.Create(ctx, user2)

	// Assert
	assert.Error(t, err, "Should fail due to duplicate email")
}

func TestUserRepository_CreateDuplicate_Username(t *testing.T) {
	repo, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	user1 := testutil.CreateTestUser()
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create another user with same username
	user2 := testutil.CreateTestUser()
	user2.Username = user1.Username // Same username

	// Act
	err = repo.Create(ctx, user2)

	// Assert
	assert.Error(t, err, "Should fail due to duplicate username")
}
