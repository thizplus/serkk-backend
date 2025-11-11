//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gofiber-template/domain/repositories"
	"gofiber-template/pkg/config"
	"gofiber-template/pkg/testutil"
)

func setupPostRepoTest(t *testing.T) (*PostRepositoryImpl, *UserRepositoryImpl, func()) {
	// Get test database
	db, err := config.GetTestDB()
	require.NoError(t, err, "Failed to connect to test database")

	// Create repositories
	postRepo := &PostRepositoryImpl{db: db}
	userRepo := &UserRepositoryImpl{db: db}

	// Cleanup function
	cleanup := func() {
		err := config.CleanTestDB(db)
		assert.NoError(t, err, "Failed to clean test database")
	}

	return postRepo, userRepo, cleanup
}

func TestPostRepository_Create(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author first
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	// Create post
	post := testutil.CreateTestPost(author.ID)

	// Act
	err = postRepo.Create(ctx, post)

	// Assert
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, post.ID)

	// Verify post was created
	found, err := postRepo.GetByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, post.Title, found.Title)
	assert.Equal(t, post.Content, found.Content)
	assert.Equal(t, author.ID, found.AuthorID)
}

func TestPostRepository_GetByID(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author and post
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	post := testutil.CreateTestPost(author.ID)
	err = postRepo.Create(ctx, post)
	require.NoError(t, err)

	// Act
	found, err := postRepo.GetByID(ctx, post.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, post.ID, found.ID)
	assert.Equal(t, post.Title, found.Title)
	assert.Equal(t, post.Content, found.Content)
}

func TestPostRepository_GetByID_NotFound(t *testing.T) {
	postRepo, _, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	nonExistentID := uuid.New()

	// Act
	found, err := postRepo.GetByID(ctx, nonExistentID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestPostRepository_Update(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author and post
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	post := testutil.CreateTestPost(author.ID)
	err = postRepo.Create(ctx, post)
	require.NoError(t, err)

	// Modify post
	post.Title = "Updated Title"
	post.Content = "Updated content"

	// Act
	err = postRepo.Update(ctx, post.ID, post)

	// Assert
	assert.NoError(t, err)

	// Verify update
	found, err := postRepo.GetByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", found.Title)
	assert.Equal(t, "Updated content", found.Content)
}

func TestPostRepository_Delete(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author and post
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	post := testutil.CreateTestPost(author.ID)
	err = postRepo.Create(ctx, post)
	require.NoError(t, err)

	// Act
	err = postRepo.Delete(ctx, post.ID)

	// Assert
	assert.NoError(t, err)

	// Verify deletion (soft delete - IsDeleted = true)
	found, err := postRepo.GetByID(ctx, post.ID)
	// Should still exist but marked as deleted
	if err == nil {
		assert.True(t, found.IsDeleted)
	}
}

func TestPostRepository_List(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	// Create multiple posts
	for i := 0; i < 5; i++ {
		post := testutil.CreateTestPost(author.ID)
		err := postRepo.Create(ctx, post)
		require.NoError(t, err)
	}

	// Act
	posts, err := postRepo.List(ctx, 0, 10, repositories.SortByNew)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(posts), 5)
}

func TestPostRepository_ListByAuthor(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create two authors
	author1 := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author1)
	require.NoError(t, err)

	author2 := testutil.CreateTestUser()
	err = userRepo.Create(ctx, author2)
	require.NoError(t, err)

	// Create posts for author1
	for i := 0; i < 3; i++ {
		post := testutil.CreateTestPost(author1.ID)
		err := postRepo.Create(ctx, post)
		require.NoError(t, err)
	}

	// Create posts for author2
	for i := 0; i < 2; i++ {
		post := testutil.CreateTestPost(author2.ID)
		err := postRepo.Create(ctx, post)
		require.NoError(t, err)
	}

	// Act - Get posts by author1
	posts, err := postRepo.ListByAuthor(ctx, author1.ID, 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(posts), 3)
	for _, post := range posts {
		assert.Equal(t, author1.ID, post.AuthorID)
	}
}

func TestPostRepository_Count(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	// Create posts
	for i := 0; i < 3; i++ {
		post := testutil.CreateTestPost(author.ID)
		err := postRepo.Create(ctx, post)
		require.NoError(t, err)
	}

	// Act
	count, err := postRepo.Count(ctx)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(3))
}

func TestPostRepository_CountByAuthor(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	// Create posts
	for i := 0; i < 4; i++ {
		post := testutil.CreateTestPost(author.ID)
		err := postRepo.Create(ctx, post)
		require.NoError(t, err)
	}

	// Act
	count, err := postRepo.CountByAuthor(ctx, author.ID)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(4))
}

func TestPostRepository_UpdateVoteCount(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author and post
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	post := testutil.CreateTestPost(author.ID)
	err = postRepo.Create(ctx, post)
	require.NoError(t, err)

	initialVotes := post.Votes

	// Act - Increment vote
	err = postRepo.UpdateVoteCount(ctx, post.ID, 1)
	assert.NoError(t, err)

	// Verify
	found, err := postRepo.GetByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, initialVotes+1, found.Votes)

	// Act - Decrement vote
	err = postRepo.UpdateVoteCount(ctx, post.ID, -1)
	assert.NoError(t, err)

	// Verify
	found, err = postRepo.GetByID(ctx, post.ID)
	assert.NoError(t, err)
	assert.Equal(t, initialVotes, found.Votes)
}

func TestPostRepository_Search(t *testing.T) {
	postRepo, userRepo, cleanup := setupPostRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create author
	author := testutil.CreateTestUser()
	err := userRepo.Create(ctx, author)
	require.NoError(t, err)

	// Create posts with specific keywords
	post1 := testutil.CreateTestPostWithData(author.ID, "Golang Tutorial", "Learn Golang basics")
	err = postRepo.Create(ctx, post1)
	require.NoError(t, err)

	post2 := testutil.CreateTestPostWithData(author.ID, "Python Guide", "Learn Python programming")
	err = postRepo.Create(ctx, post2)
	require.NoError(t, err)

	// Act - Search for "Golang"
	posts, err := postRepo.Search(ctx, "Golang", 0, 10)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(posts), 1)

	// Verify at least one post contains "Golang"
	foundGolang := false
	for _, post := range posts {
		if post.ID == post1.ID {
			foundGolang = true
			break
		}
	}
	assert.True(t, foundGolang, "Should find the Golang post")
}
