package serviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/utils"
)

type CommentServiceImpl struct {
	commentRepo  repositories.CommentRepository
	postRepo     repositories.PostRepository
	voteRepo     repositories.VoteRepository
	notifService services.NotificationService
}

func NewCommentService(
	commentRepo repositories.CommentRepository,
	postRepo repositories.PostRepository,
	voteRepo repositories.VoteRepository,
	notifService services.NotificationService,
) services.CommentService {
	return &CommentServiceImpl{
		commentRepo:  commentRepo,
		postRepo:     postRepo,
		voteRepo:     voteRepo,
		notifService: notifService,
	}
}

func (s *CommentServiceImpl) CreateComment(ctx context.Context, userID uuid.UUID, req *dto.CreateCommentRequest) (*dto.CommentResponse, error) {
	// Verify post exists
	post, err := s.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	depth := 0
	var parentComment *models.Comment

	// If replying to a comment, verify parent and calculate depth
	if req.ParentID != nil {
		parentComment, err = s.commentRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, errors.New("parent comment not found")
		}

		// Check max depth (10)
		if parentComment.Depth >= 10 {
			return nil, errors.New("maximum comment depth reached")
		}

		depth = parentComment.Depth + 1
	}

	// Create comment
	comment := &models.Comment{
		ID:        uuid.New(),
		PostID:    req.PostID,
		AuthorID:  userID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		Votes:     0,
		Depth:     depth,
		IsDeleted: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, err
	}

	// Increment post comment count
	_ = s.postRepo.IncrementCommentCount(ctx, req.PostID)

	// Send notification to post author or parent comment author
	if req.ParentID != nil && parentComment != nil {
		// Reply notification
		if parentComment.AuthorID != userID {
			_ = s.notifService.CreateNotification(
				ctx,
				parentComment.AuthorID,
				userID,
				"reply",
				"ตอบกลับความคิดเห็นของคุณ",
				&req.PostID,
				&comment.ID,
			)
		}
	} else {
		// Comment notification to post author
		if post.AuthorID != userID {
			_ = s.notifService.CreateNotification(
				ctx,
				post.AuthorID,
				userID,
				"reply",
				"แสดงความคิดเห็นในโพสต์ของคุณ",
				&req.PostID,
				&comment.ID,
			)
		}
	}

	return s.GetComment(ctx, comment.ID, &userID)
}

func (s *CommentServiceImpl) GetComment(ctx context.Context, commentID uuid.UUID, userID *uuid.UUID) (*dto.CommentResponse, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	resp := dto.CommentToCommentResponse(comment)

	// Add user-specific data if authenticated
	if userID != nil {
		// Get user's vote
		vote, _ := s.voteRepo.GetVote(ctx, *userID, commentID, "comment")
		if vote != nil {
			resp.UserVote = &vote.VoteType
		}

		// Get reply count
		replyCount, _ := s.commentRepo.CountReplies(ctx, commentID)
		replyCountInt := int(replyCount)
		resp.ReplyCount = &replyCountInt
	}

	return resp, nil
}

func (s *CommentServiceImpl) UpdateComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, req *dto.UpdateCommentRequest) (*dto.CommentResponse, error) {
	// Get existing comment
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if comment.AuthorID != userID {
		return nil, errors.New("unauthorized: not comment owner")
	}

	// Update content
	comment.Content = req.Content
	comment.UpdatedAt = time.Now()

	err = s.commentRepo.Update(ctx, commentID, comment)
	if err != nil {
		return nil, err
	}

	return s.GetComment(ctx, commentID, &userID)
}

func (s *CommentServiceImpl) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	// Get comment
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Check ownership
	if comment.AuthorID != userID {
		return errors.New("unauthorized: not comment owner")
	}

	// Soft delete
	err = s.commentRepo.Delete(ctx, commentID)
	if err != nil {
		return err
	}

	// Decrement post comment count
	_ = s.postRepo.DecrementCommentCount(ctx, comment.PostID)

	return nil
}

func (s *CommentServiceImpl) ListCommentsByPost(ctx context.Context, postID uuid.UUID, offset, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListResponse, error) {
	comments, err := s.commentRepo.ListByPost(ctx, postID, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	count, err := s.commentRepo.CountByPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListResponse(ctx, comments, count, offset, limit, userID)
}

func (s *CommentServiceImpl) ListCommentsByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int, userID *uuid.UUID) (*dto.CommentListResponse, error) {
	comments, err := s.commentRepo.ListByAuthor(ctx, authorID, offset, limit)
	if err != nil {
		return nil, err
	}

	count, err := s.commentRepo.CountByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListResponse(ctx, comments, count, offset, limit, userID)
}

func (s *CommentServiceImpl) ListReplies(ctx context.Context, parentID uuid.UUID, offset, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListResponse, error) {
	comments, err := s.commentRepo.ListReplies(ctx, parentID, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	count, err := s.commentRepo.CountReplies(ctx, parentID)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListResponse(ctx, comments, count, offset, limit, userID)
}

func (s *CommentServiceImpl) GetCommentTree(ctx context.Context, postID uuid.UUID, maxDepth int, userID *uuid.UUID) (*dto.CommentTreeResponse, error) {
	if maxDepth > 10 {
		maxDepth = 10
	}

	comments, err := s.commentRepo.GetCommentTree(ctx, postID, maxDepth)
	if err != nil {
		return nil, err
	}

	// Build tree structure
	commentMap := make(map[uuid.UUID]*dto.CommentWithRepliesResponse)
	var rootComments []*dto.CommentWithRepliesResponse

	// Batch get user votes if authenticated
	var voteMap map[uuid.UUID]*models.Vote
	if userID != nil {
		commentIDs := make([]uuid.UUID, len(comments))
		for i, comment := range comments {
			commentIDs[i] = comment.ID
		}
		voteMap, _ = s.voteRepo.GetUserVotesForTargets(ctx, *userID, commentIDs, "comment")
	}

	// First pass: create all comment responses
	for _, comment := range comments {
		resp := &dto.CommentWithRepliesResponse{
			CommentResponse: *dto.CommentToCommentResponse(comment),
			Replies:         []dto.CommentWithRepliesResponse{},
		}

		// Add user-specific data
		if userID != nil {
			if vote, ok := voteMap[comment.ID]; ok {
				resp.UserVote = &vote.VoteType
			}
		}

		commentMap[comment.ID] = resp
	}

	// Second pass: build children lists for each comment
	childrenMap := make(map[uuid.UUID][]uuid.UUID)
	for _, comment := range comments {
		if comment.ParentID != nil {
			childrenMap[*comment.ParentID] = append(childrenMap[*comment.ParentID], comment.ID)
		}
	}

	// Third pass: build tree recursively
	var buildTree func(commentID uuid.UUID) dto.CommentWithRepliesResponse
	buildTree = func(commentID uuid.UUID) dto.CommentWithRepliesResponse {
		node := *commentMap[commentID]
		if childIDs, ok := childrenMap[commentID]; ok {
			node.Replies = make([]dto.CommentWithRepliesResponse, 0, len(childIDs))
			for _, childID := range childIDs {
				node.Replies = append(node.Replies, buildTree(childID))
			}
		}
		return node
	}

	// Fourth pass: collect and build root comments
	for _, comment := range comments {
		if comment.ParentID == nil {
			tree := buildTree(comment.ID)
			rootComments = append(rootComments, &tree)
		}
	}

	return &dto.CommentTreeResponse{
		Comments: convertToCommentWithReplies(rootComments),
		Total:    int64(len(comments)),
	}, nil
}

func (s *CommentServiceImpl) GetParentChain(ctx context.Context, commentID uuid.UUID, userID *uuid.UUID) ([]*dto.CommentResponse, error) {
	comments, err := s.commentRepo.GetParentChain(ctx, commentID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.CommentResponse, len(comments))
	for i, comment := range comments {
		resp := dto.CommentToCommentResponse(comment)

		// Add user-specific data
		if userID != nil {
			vote, _ := s.voteRepo.GetVote(ctx, *userID, comment.ID, "comment")
			if vote != nil {
				resp.UserVote = &vote.VoteType
			}
		}

		responses[i] = resp
	}

	return responses, nil
}

// Helper functions
func (s *CommentServiceImpl) buildCommentListResponse(ctx context.Context, comments []*models.Comment, count int64, offset, limit int, userID *uuid.UUID) (*dto.CommentListResponse, error) {
	responses := make([]dto.CommentResponse, len(comments))

	// Collect comment IDs for batch operations
	commentIDs := make([]uuid.UUID, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.ID
	}

	// Batch get user votes if authenticated
	var voteMap map[uuid.UUID]*models.Vote
	if userID != nil {
		voteMap, _ = s.voteRepo.GetUserVotesForTargets(ctx, *userID, commentIDs, "comment")
	}

	// Build responses
	for i, comment := range comments {
		resp := dto.CommentToCommentResponse(comment)

		// Add user-specific data
		if userID != nil {
			if vote, ok := voteMap[comment.ID]; ok {
				resp.UserVote = &vote.VoteType
			}

			// Get reply count
			replyCount, _ := s.commentRepo.CountReplies(ctx, comment.ID)
			replyCountInt := int(replyCount)
			resp.ReplyCount = &replyCountInt
		}

		responses[i] = *resp
	}

	return &dto.CommentListResponse{
		Comments: responses,
		Meta: dto.PaginationMeta{
			Total:  &count,
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

func convertToCommentWithReplies(comments []*dto.CommentWithRepliesResponse) []dto.CommentWithRepliesResponse {
	result := make([]dto.CommentWithRepliesResponse, len(comments))
	for i, c := range comments {
		result[i] = *c
	}
	return result
}

var _ services.CommentService = (*CommentServiceImpl)(nil)

// Cursor-based service methods
func (s *CommentServiceImpl) ListCommentsByPostWithCursor(ctx context.Context, postID uuid.UUID, cursor string, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListCursorResponse, error) {
	// Decode cursor
	decodedCursor, err := utils.DecodePostCursor(cursor)
	if err != nil && cursor != "" {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to check if there are more
	comments, err := s.commentRepo.ListByPostWithCursor(ctx, postID, decodedCursor, limit+1, sortBy)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListCursorResponse(ctx, comments, limit, sortBy, userID)
}

func (s *CommentServiceImpl) ListCommentsByAuthorWithCursor(ctx context.Context, authorID uuid.UUID, cursor string, limit int, userID *uuid.UUID) (*dto.CommentListCursorResponse, error) {
	// Decode cursor
	decodedCursor, err := utils.DecodePostCursor(cursor)
	if err != nil && cursor != "" {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to check if there are more
	comments, err := s.commentRepo.ListByAuthorWithCursor(ctx, authorID, decodedCursor, limit+1)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListCursorResponse(ctx, comments, limit, repositories.CommentSortByNew, userID)
}

func (s *CommentServiceImpl) ListRepliesWithCursor(ctx context.Context, parentID uuid.UUID, cursor string, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListCursorResponse, error) {
	// Decode cursor
	decodedCursor, err := utils.DecodePostCursor(cursor)
	if err != nil && cursor != "" {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to check if there are more
	comments, err := s.commentRepo.ListRepliesWithCursor(ctx, parentID, decodedCursor, limit+1, sortBy)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListCursorResponse(ctx, comments, limit, sortBy, userID)
}

// Helper function to build cursor-based response
func (s *CommentServiceImpl) buildCommentListCursorResponse(ctx context.Context, comments []*models.Comment, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListCursorResponse, error) {
	// Check if there are more results
	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	// Collect comment IDs for batch operations
	commentIDs := make([]uuid.UUID, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.ID
	}

	// Batch get user votes if authenticated
	var voteMap map[uuid.UUID]*models.Vote
	if userID != nil && len(commentIDs) > 0 {
		voteMap, _ = s.voteRepo.GetUserVotesForTargets(ctx, *userID, commentIDs, "comment")
	}

	// Build responses
	responses := make([]dto.CommentResponse, len(comments))
	for i, comment := range comments {
		resp := dto.CommentToCommentResponse(comment)

		// Add user-specific data
		if userID != nil {
			if vote, ok := voteMap[comment.ID]; ok {
				resp.UserVote = &vote.VoteType
			}

			// Get reply count
			replyCount, _ := s.commentRepo.CountReplies(ctx, comment.ID)
			replyCountInt := int(replyCount)
			resp.ReplyCount = &replyCountInt
		}

		responses[i] = *resp
	}

	// Build next cursor
	var nextCursor *string
	if hasMore && len(comments) > 0 {
		lastComment := comments[len(comments)-1]
		var sortValue *float64
		switch sortBy {
		case repositories.CommentSortByTop:
			votes := float64(lastComment.Votes)
			sortValue = &votes
		}
		encoded, err := utils.EncodePostCursor(sortValue, lastComment.CreatedAt, lastComment.ID)
		if err != nil {
			return nil, err
		}
		nextCursor = &encoded
	}

	return &dto.CommentListCursorResponse{
		Comments: responses,
		Meta: dto.CursorPaginationMeta{
			NextCursor: nextCursor,
			HasMore:    hasMore,
			Limit:      limit,
		},
	}, nil
}
