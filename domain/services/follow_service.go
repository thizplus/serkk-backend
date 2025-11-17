package services

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
)

type FollowService interface {
	// Follow/Unfollow
	Follow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*dto.FollowResponse, error)
	Unfollow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error

	// Check relationship
	IsFollowing(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (*dto.FollowStatusResponse, error)

	// Get followers/following (offset-based, deprecated)
	GetFollowers(ctx context.Context, userID uuid.UUID, offset, limit int, currentUserID *uuid.UUID) (*dto.FollowerListResponse, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, offset, limit int, currentUserID *uuid.UUID) (*dto.FollowingListResponse, error)

	// Get followers/following with cursor (cursor-based pagination)
	GetFollowersWithCursor(ctx context.Context, userID uuid.UUID, cursor string, limit int, currentUserID *uuid.UUID) (*dto.FollowerListCursorResponse, error)
	GetFollowingWithCursor(ctx context.Context, userID uuid.UUID, cursor string, limit int, currentUserID *uuid.UUID) (*dto.FollowingListCursorResponse, error)

	// Mutual follows
	GetMutualFollows(ctx context.Context, userID uuid.UUID, offset, limit int) (*dto.FollowerListResponse, error)
}
