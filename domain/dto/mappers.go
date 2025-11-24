package dto

import (
	"encoding/json"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
)

// UserToUserResponse is DEPRECATED
// Since User is now an alias to UsersIdentity, use UsersIdentityToUserResponse instead
// This is kept for backward compatibility only
func UserToUserResponse(user *models.User) *UserResponse {
	return UsersIdentityToUserResponse(user)
}

// IdentityWithProfileToUserResponse combines UsersIdentity and UserProfile data
// This is the new mapper for V2 architecture (uses users_identity instead of users_cache)
func IdentityWithProfileToUserResponse(identity *models.UsersIdentity, profile *models.UserProfile) *UserResponse {
	if identity == nil {
		return nil
	}

	resp := &UserResponse{
		ID:        identity.ID,
		Email:     identity.Email,
		Username:  identity.Username,
		Role:      "user",  // Default role (Auth Service manages actual roles)
		IsActive:  true,    // Default active (Auth Service manages actual status)
		CreatedAt: identity.CreatedAt,
		UpdatedAt: identity.UpdatedAt,
	}

	// All profile data (including displayName and avatar) comes from user_profiles
	if profile != nil {
		resp.DisplayName = profile.DisplayName
		resp.Avatar = profile.Avatar
		resp.Bio = profile.Bio
		resp.Location = profile.Location
		resp.Website = profile.Website
		resp.Karma = profile.Karma
		resp.FollowersCount = profile.FollowersCount
		resp.FollowingCount = profile.FollowingCount
	}

	return resp
}

// UserWithProfileToUserResponse is DEPRECATED
// Use IdentityWithProfileToUserResponse instead (V2 architecture)
// This is kept for backward compatibility only
func UserWithProfileToUserResponse(identity *models.UsersIdentity, profile *models.UserProfile) *UserResponse {
	return IdentityWithProfileToUserResponse(identity, profile)
}

// UsersIdentityToUserResponse converts UsersIdentity to UserResponse
// Automatically includes Profile data if it's preloaded
// Used for mapping identity data from Auth Service (via users_identity table)
func UsersIdentityToUserResponse(identity *models.UsersIdentity) *UserResponse {
	if identity == nil {
		return nil
	}

	resp := &UserResponse{
		ID:        identity.ID,
		Email:     identity.Email,
		Username:  identity.Username,
		Role:      "user",  // Default role (Auth Service manages actual roles)
		IsActive:  true,    // Default active (Auth Service manages actual status)
		CreatedAt: identity.CreatedAt,
		UpdatedAt: identity.UpdatedAt,
	}

	// If Profile is preloaded, include it
	if identity.Profile != nil {
		resp.DisplayName = identity.Profile.DisplayName
		resp.Avatar = identity.Profile.Avatar
		resp.Bio = identity.Profile.Bio
		resp.Location = identity.Profile.Location
		resp.Website = identity.Profile.Website
		resp.Karma = identity.Profile.Karma
		resp.FollowersCount = identity.Profile.FollowersCount
		resp.FollowingCount = identity.Profile.FollowingCount
	}

	return resp
}

// CreateUserRequestToUser is DEPRECATED
// Backend Service should NOT create users - this belongs to Auth Service
// Keeping for backward compatibility, but will be removed
func CreateUserRequestToUser(req *CreateUserRequest) *models.User {
	return &models.User{
		Email:    req.Email,
		Username: req.Username,
		// Note: DisplayName is now in UserProfile, not UsersIdentity
		// Password field removed - Backend doesn't manage passwords
	}
}

// UpdateUserRequestToUser is DEPRECATED
// Use UpdateUserRequestToUserProfile for ALL profile updates instead
// displayName and avatar are now in user_profiles (moved from Auth Service)
func UpdateUserRequestToUser(req *UpdateUserRequest) *models.User {
	// Since User is now UsersIdentity, and displayName/avatar are in user_profiles,
	// this function effectively does nothing. Use UpdateUserRequestToUserProfile instead.
	return &models.User{}
}

// UpdateUserRequestToUserProfile maps update request to UserProfile
// Now includes displayName and avatar (moved from Auth Service)
func UpdateUserRequestToUserProfile(req *UpdateUserRequest) *models.UserProfile {
	return &models.UserProfile{
		DisplayName: req.DisplayName, // ⭐ New: now in user_profiles
		Avatar:      req.Avatar,      // ⭐ New: now in user_profiles
		Bio:         req.Bio,
		Location:    req.Location,
		Website:     req.Website,
	}
}

func TaskToTaskResponse(task *models.Task, identity *models.UsersIdentity) *TaskResponse {
	if task == nil {
		return nil
	}
	taskResp := &TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		UserID:      task.UserID,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
	if identity != nil {
		taskResp.User = *UsersIdentityToUserResponse(identity)
	}
	return taskResp
}

func CreateTaskRequestToTask(req *CreateTaskRequest) *models.Task {
	return &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}
}

func UpdateTaskRequestToTask(req *UpdateTaskRequest) *models.Task {
	return &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}
}

func JobToJobResponse(job *models.Job) *JobResponse {
	if job == nil {
		return nil
	}
	return &JobResponse{
		ID:        job.ID,
		Name:      job.Name,
		CronExpr:  job.CronExpr,
		Payload:   job.Payload,
		Status:    job.Status,
		LastRun:   job.LastRun,
		NextRun:   job.NextRun,
		IsActive:  job.IsActive,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}
}

func CreateJobRequestToJob(req *CreateJobRequest) *models.Job {
	return &models.Job{
		Name:     req.Name,
		CronExpr: req.CronExpr,
		Payload:  req.Payload,
	}
}

func UpdateJobRequestToJob(req *UpdateJobRequest) *models.Job {
	return &models.Job{
		Name:     req.Name,
		CronExpr: req.CronExpr,
		Payload:  req.Payload,
		IsActive: req.IsActive,
	}
}

func FileToFileResponse(file *models.File) *FileResponse {
	if file == nil {
		return nil
	}
	return &FileResponse{
		ID:        file.ID,
		FileName:  file.FileName,
		FileSize:  file.FileSize,
		MimeType:  file.MimeType,
		URL:       file.URL,
		CDNPath:   file.CDNPath,
		UserID:    file.UserID,
		CreatedAt: file.CreatedAt,
		UpdatedAt: file.UpdatedAt,
	}
}

// Post mappers
func PostToPostResponse(post *models.Post) *PostResponse {
	if post == nil {
		return nil
	}

	resp := &PostResponse{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		Author:       *UsersIdentityToUserResponse(&post.Author),
		Votes:        post.Votes,
		CommentCount: post.CommentCount,
		Type:         post.Type,
		Status:       post.Status,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}

	// Map media
	if len(post.Media) > 0 {
		resp.Media = make([]MediaResponse, len(post.Media))
		for i, media := range post.Media {
			resp.Media[i] = *MediaToMediaResponse(&media)
		}
	}

	// Map tags
	if len(post.Tags) > 0 {
		resp.Tags = make([]TagResponse, len(post.Tags))
		for i, tag := range post.Tags {
			resp.Tags[i] = *TagToTagResponse(&tag)
		}
	}

	// Map source post (for crossposts)
	if post.SourcePost != nil {
		resp.SourcePost = PostToPostResponse(post.SourcePost)
	}

	return resp
}

// Comment mappers
func CommentToCommentResponse(comment *models.Comment) *CommentResponse {
	if comment == nil {
		return nil
	}

	resp := &CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		Author:    *UsersIdentityToUserResponse(&comment.Author),
		Content:   comment.Content,
		Votes:     comment.Votes,
		Depth:     comment.Depth,
		IsDeleted: comment.IsDeleted,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}

	// Map post summary if available
	if comment.Post.ID != (uuid.UUID{}) {
		resp.Post = PostToPostSummaryResponse(&comment.Post)
	}

	return resp
}

// PostToPostSummaryResponse converts a Post model to a lightweight PostSummaryResponse
func PostToPostSummaryResponse(post *models.Post) *PostSummaryResponse {
	if post == nil {
		return nil
	}

	return &PostSummaryResponse{
		ID:        post.ID,
		Title:     post.Title,
		Author:    *UsersIdentityToUserResponse(&post.Author),
		CreatedAt: post.CreatedAt,
	}
}

func CommentToCommentWithReplies(comment *models.Comment, replies []models.Comment) *CommentWithRepliesResponse {
	if comment == nil {
		return nil
	}

	resp := &CommentWithRepliesResponse{
		CommentResponse: *CommentToCommentResponse(comment),
	}

	if len(replies) > 0 {
		resp.Replies = make([]CommentWithRepliesResponse, len(replies))
		for i, reply := range replies {
			replyResp := CommentToCommentWithReplies(&reply, nil)
			if replyResp != nil {
				resp.Replies[i] = *replyResp
			}
		}
	}

	return resp
}

// Vote mappers
func VoteToVoteResponse(vote *models.Vote) *VoteResponse {
	if vote == nil {
		return nil
	}

	return &VoteResponse{
		TargetID:   vote.TargetID,
		TargetType: vote.TargetType,
		VoteType:   vote.VoteType,
		CreatedAt:  vote.CreatedAt,
	}
}

// Tag mappers
func TagToTagResponse(tag *models.Tag) *TagResponse {
	if tag == nil {
		return nil
	}

	return &TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		PostCount: tag.PostCount,
		CreatedAt: tag.CreatedAt,
	}
}

// Media mappers
func MediaToMediaResponse(media *models.Media) *MediaResponse {
	if media == nil {
		return nil
	}

	return &MediaResponse{
		ID:         media.ID,
		UserID:     media.UserID,
		Type:       media.Type,
		FileName:   media.FileName,
		MimeType:   media.MimeType,
		Size:       media.Size,
		URL:        media.URL,
		Thumbnail:  media.Thumbnail,
		Width:      media.Width,
		Height:     media.Height,
		Duration:   media.Duration,
		SourceType: media.SourceType,
		SourceID:   media.SourceID,
		CreatedAt:  media.CreatedAt,
	}
}

func MediaToMediaUploadResponse(media *models.Media) *MediaUploadResponse {
	if media == nil {
		return nil
	}

	return &MediaUploadResponse{
		ID:        media.ID,
		Type:      media.Type,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		Size:      media.Size,
		URL:       media.URL,
		Thumbnail: media.Thumbnail,
		Width:     media.Width,
		Height:    media.Height,
		Duration:  media.Duration,
		CreatedAt: media.CreatedAt,
	}
}

// Notification mappers
func NotificationToNotificationResponse(notification *models.Notification) *NotificationResponse {
	if notification == nil {
		return nil
	}

	return &NotificationResponse{
		ID:        notification.ID,
		User:      *UsersIdentityToUserResponse(&notification.User),
		Sender:    *UsersIdentityToUserResponse(&notification.Sender),
		Type:      notification.Type,
		Message:   notification.Message,
		PostID:    notification.PostID,
		CommentID: notification.CommentID,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
	}
}

func NotificationSettingsToResponse(settings *models.NotificationSettings) *NotificationSettingsResponse {
	if settings == nil {
		return nil
	}

	return &NotificationSettingsResponse{
		UserID:             settings.UserID,
		Replies:            settings.Replies,
		Mentions:           settings.Mentions,
		Votes:              settings.Votes,
		Follows:            settings.Follows,
		EmailNotifications: settings.EmailNotifications,
		UpdatedAt:          settings.UpdatedAt,
	}
}

// SearchHistory mappers
func SearchHistoryToResponse(history *models.SearchHistory) *SearchHistoryResponse {
	if history == nil {
		return nil
	}

	return &SearchHistoryResponse{
		ID:         history.ID,
		Query:      history.Query,
		Type:       history.Type,
		SearchedAt: history.SearchedAt,
	}
}

// ============================================================================
// Chat mappers
// ============================================================================

// MessageToMessageResponse converts Message model to MessageResponse DTO
func MessageToMessageResponse(message *models.Message) *MessageResponse {
	if message == nil {
		return nil
	}

	resp := &MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Sender:         *UsersIdentityToUserResponse(&message.Sender),
		Receiver:       *UsersIdentityToUserResponse(&message.Receiver),
		Type:           string(message.Type),
		Content:        message.Content,
		IsRead:         message.IsRead,
		ReadAt:         message.ReadAt,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      message.UpdatedAt,

		// Helper fields
		SenderId: message.SenderID,
	}

	// Unmarshal Media JSONB to []MessageMedia
	if message.Media != nil && len(message.Media) > 0 {
		var mediaList []MessageMedia
		if err := json.Unmarshal(message.Media, &mediaList); err == nil {
			resp.Media = mediaList
		}
	}

	return resp
}

// ConversationToConversationResponse converts Conversation model to ConversationResponse DTO
// currentUserID is needed to determine who the "other user" is and which unread count to show
func ConversationToConversationResponse(conversation *models.Conversation, currentUserID uuid.UUID) *ConversationResponse {
	if conversation == nil {
		return nil
	}

	// Determine who is the "other user" and get their unread count
	var otherUser models.UsersIdentity
	var unreadCount int

	if conversation.User1ID == currentUserID {
		otherUser = conversation.User2
		unreadCount = conversation.User1UnreadCount
	} else {
		otherUser = conversation.User1
		unreadCount = conversation.User2UnreadCount
	}

	resp := &ConversationResponse{
		ID:            conversation.ID,
		OtherUser:     *UsersIdentityToUserResponse(&otherUser),
		LastMessageAt: conversation.LastMessageAt,
		UnreadCount:   unreadCount,
		CreatedAt:     conversation.CreatedAt,
		UpdatedAt:     conversation.UpdatedAt,
	}

	return resp
}

// BlockToBlockedUserResponse converts Block model to BlockedUserResponse DTO
func BlockToBlockedUserResponse(block *models.Block) *BlockedUserResponse {
	if block == nil {
		return nil
	}

	return &BlockedUserResponse{
		User:      *UsersIdentityToUserResponse(&block.Blocked),
		BlockedAt: block.CreatedAt,
	}
}
