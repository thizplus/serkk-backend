package dto

import (
	"time"

	"github.com/google/uuid"
)

// FollowRequest - Request for following a user
type FollowRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required,uuid"`
}

// UnfollowRequest - Request for unfollowing a user
type UnfollowRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required,uuid"`
}

// FollowResponse - Response for follow relationship
type FollowResponse struct {
	FollowerID  uuid.UUID `json:"followerId"`
	FollowingID uuid.UUID `json:"followingId"`
	CreatedAt   time.Time `json:"createdAt"`
}

// FollowerListResponse - Response for followers list (offset-based)
type FollowerListResponse struct {
	Users []UserResponse `json:"users"`
	Meta  PaginationMeta `json:"meta"`
}

// FollowingListResponse - Response for following list (offset-based)
type FollowingListResponse struct {
	Users []UserResponse `json:"users"`
	Meta  PaginationMeta `json:"meta"`
}

// FollowerListCursorResponse - Response for followers list (cursor-based)
type FollowerListCursorResponse struct {
	Users []UserResponse       `json:"users"`
	Meta  CursorPaginationMeta `json:"meta"`
}

// FollowingListCursorResponse - Response for following list (cursor-based)
type FollowingListCursorResponse struct {
	Users []UserResponse       `json:"users"`
	Meta  CursorPaginationMeta `json:"meta"`
}

// FollowStatusResponse - Response for checking follow status
type FollowStatusResponse struct {
	IsFollowing bool `json:"isFollowing"`
	IsMutual    bool `json:"isMutual,omitempty"` // Both users follow each other
}
