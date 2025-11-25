package serviceimpl

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/websocket"
)

type VoteServiceImpl struct {
	voteRepo       repositories.VoteRepository
	postRepo       repositories.PostRepository
	commentRepo    repositories.CommentRepository
	userRepo       repositories.UserRepository
	notifService   services.NotificationService
	profileService services.UserProfileService
}

func NewVoteService(
	voteRepo repositories.VoteRepository,
	postRepo repositories.PostRepository,
	commentRepo repositories.CommentRepository,
	userRepo repositories.UserRepository,
	notifService services.NotificationService,
	profileService services.UserProfileService,
) services.VoteService {
	return &VoteServiceImpl{
		voteRepo:       voteRepo,
		postRepo:       postRepo,
		commentRepo:    commentRepo,
		userRepo:       userRepo,
		notifService:   notifService,
		profileService: profileService,
	}
}

func (s *VoteServiceImpl) Vote(ctx context.Context, userID uuid.UUID, req *dto.VoteRequest) (*dto.VoteResponse, error) {
	// Check if user already voted
	existingVote, _ := s.voteRepo.GetVote(ctx, userID, req.TargetID, req.TargetType)

	vote := &models.Vote{
		UserID:     userID,
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		VoteType:   req.VoteType,
		CreatedAt:  time.Now(),
	}

	// Calculate vote change
	voteChange := 0
	if existingVote == nil {
		// New vote
		if req.VoteType == "up" {
			voteChange = 1
		} else {
			voteChange = -1
		}
	} else if existingVote.VoteType != req.VoteType {
		// Changing vote
		if req.VoteType == "up" {
			voteChange = 2 // from -1 to +1
		} else {
			voteChange = -2 // from +1 to -1
		}
	}

	// Save vote (upsert)
	err := s.voteRepo.Vote(ctx, vote)
	if err != nil {
		return nil, err
	}

	// Update vote count on target
	if voteChange != 0 {
		if req.TargetType == "post" {
			_ = s.postRepo.UpdateVoteCount(ctx, req.TargetID, voteChange)

			// Get updated vote count and broadcast via WebSocket
			post, err := s.postRepo.GetByID(ctx, req.TargetID)
			if err == nil && post != nil {
				// Broadcast vote:updated event to post owner only
				websocket.Manager.BroadcastToUser(post.AuthorID, "vote:updated", map[string]interface{}{
					"targetId":   req.TargetID.String(),
					"targetType": "post",
					"voteCount":  post.Votes,
					"voteType":   req.VoteType,
				})
				log.Printf("📢 Broadcasted vote:updated to post owner %s for post %s (new count: %d)", post.AuthorID, req.TargetID, post.Votes)

				// ✅ NEW: Update author's karma
				if post.AuthorID != userID {
					karmaDelta := 0

					// Calculate karma change based on vote type and existing vote
					if existingVote == nil {
						// New vote
						if req.VoteType == "up" {
							karmaDelta = +10 // New upvote = +10
						} else {
							karmaDelta = -2 // New downvote = -2
						}
					} else if existingVote.VoteType != req.VoteType {
						// Vote changed
						if req.VoteType == "up" {
							karmaDelta = +12 // Changed from down to up (+10 + 2)
						} else {
							karmaDelta = -12 // Changed from up to down (-10 - 2)
						}
					}

					// Update karma (only if not voting on own post)
					if karmaDelta != 0 {
						err := s.profileService.UpdateKarma(ctx, post.AuthorID, karmaDelta)
						if err != nil {
							log.Printf("⚠️ Failed to update karma for user %s: %v", post.AuthorID, err)
							// Don't fail the vote, just log
						} else {
							log.Printf("✅ Karma updated: User %s got %+d karma (post vote)", post.AuthorID, karmaDelta)
						}
					}
				}
			}

			// Send notification to post author (only for upvotes, and only if new vote)
			if req.VoteType == "up" && existingVote == nil {
				if post != nil && post.AuthorID != userID {
					_ = s.notifService.CreateNotification(
						ctx,
						post.AuthorID,
						userID,
						"vote",
						"ถูกใจโพสต์ของคุณ",
						&req.TargetID,
						nil,
					)
				}
			}
		} else if req.TargetType == "comment" {
			_ = s.commentRepo.UpdateVoteCount(ctx, req.TargetID, voteChange)

			// Get updated vote count and broadcast via WebSocket
			comment, err := s.commentRepo.GetByID(ctx, req.TargetID)
			if err == nil && comment != nil {
				// Broadcast vote:updated event to comment owner only
				websocket.Manager.BroadcastToUser(comment.AuthorID, "vote:updated", map[string]interface{}{
					"targetId":   req.TargetID.String(),
					"targetType": "comment",
					"voteCount":  comment.Votes,
					"voteType":   req.VoteType,
				})
				log.Printf("📢 Broadcasted vote:updated to comment owner %s for comment %s (new count: %d)", comment.AuthorID, req.TargetID, comment.Votes)

				// ✅ NEW: Update author's karma
				if comment.AuthorID != userID {
					karmaDelta := 0

					// Calculate karma change based on vote type and existing vote
					if existingVote == nil {
						// New vote
						if req.VoteType == "up" {
							karmaDelta = +5 // New upvote = +5
						} else {
							karmaDelta = -1 // New downvote = -1
						}
					} else if existingVote.VoteType != req.VoteType {
						// Vote changed
						if req.VoteType == "up" {
							karmaDelta = +6 // Changed from down to up (+5 + 1)
						} else {
							karmaDelta = -6 // Changed from up to down (-5 - 1)
						}
					}

					// Update karma (only if not voting on own comment)
					if karmaDelta != 0 {
						err := s.profileService.UpdateKarma(ctx, comment.AuthorID, karmaDelta)
						if err != nil {
							log.Printf("⚠️ Failed to update karma for user %s: %v", comment.AuthorID, err)
							// Don't fail the vote, just log
						} else {
							log.Printf("✅ Karma updated: User %s got %+d karma (comment vote)", comment.AuthorID, karmaDelta)
						}
					}
				}
			}

			// Send notification to comment author (only for upvotes, and only if new vote)
			if req.VoteType == "up" && existingVote == nil {
				if comment != nil && comment.AuthorID != userID {
					_ = s.notifService.CreateNotification(
						ctx,
						comment.AuthorID,
						userID,
						"vote",
						"ถูกใจความคิดเห็นของคุณ",
						&comment.PostID,
						&req.TargetID,
					)
				}
			}
		}
	}

	return &dto.VoteResponse{
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		VoteType:   req.VoteType,
		CreatedAt:  vote.CreatedAt,
	}, nil
}

func (s *VoteServiceImpl) Unvote(ctx context.Context, userID uuid.UUID, req *dto.UnvoteRequest) error {
	// Get existing vote
	existingVote, err := s.voteRepo.GetVote(ctx, userID, req.TargetID, req.TargetType)
	if err != nil || existingVote == nil {
		return nil // Already not voted
	}

	// Calculate vote change
	voteChange := 0
	if existingVote.VoteType == "up" {
		voteChange = -1
	} else {
		voteChange = 1
	}

	// Remove vote
	err = s.voteRepo.Unvote(ctx, userID, req.TargetID, req.TargetType)
	if err != nil {
		return err
	}

	// Update vote count on target
	if req.TargetType == "post" {
		_ = s.postRepo.UpdateVoteCount(ctx, req.TargetID, voteChange)

		// Get updated vote count and broadcast via WebSocket
		post, err := s.postRepo.GetByID(ctx, req.TargetID)
		if err == nil && post != nil {
			// Broadcast vote:updated event to post owner only
			websocket.Manager.BroadcastToUser(post.AuthorID, "vote:updated", map[string]interface{}{
				"targetId":   req.TargetID.String(),
				"targetType": "post",
				"voteCount":  post.Votes,
				"voteType":   "",
			})
			log.Printf("📢 Broadcasted vote:updated (unvote) to post owner %s for post %s (new count: %d)", post.AuthorID, req.TargetID, post.Votes)

			// ✅ NEW: Reverse karma change
			if post.AuthorID != userID {
				karmaDelta := 0
				if existingVote.VoteType == "up" {
					karmaDelta = -10 // Remove upvote = -10
				} else {
					karmaDelta = +2 // Remove downvote = +2 (reverse penalty)
				}

				err := s.profileService.UpdateKarma(ctx, post.AuthorID, karmaDelta)
				if err != nil {
					log.Printf("⚠️ Failed to reverse karma for user %s: %v", post.AuthorID, err)
				} else {
					log.Printf("✅ Karma reversed: User %s got %+d karma (post unvote)", post.AuthorID, karmaDelta)
				}
			}
		}
	} else if req.TargetType == "comment" {
		_ = s.commentRepo.UpdateVoteCount(ctx, req.TargetID, voteChange)

		// Get updated vote count and broadcast via WebSocket
		comment, err := s.commentRepo.GetByID(ctx, req.TargetID)
		if err == nil && comment != nil {
			// Broadcast vote:updated event to comment owner only
			websocket.Manager.BroadcastToUser(comment.AuthorID, "vote:updated", map[string]interface{}{
				"targetId":   req.TargetID.String(),
				"targetType": "comment",
				"voteCount":  comment.Votes,
				"voteType":   "",
			})
			log.Printf("📢 Broadcasted vote:updated (unvote) to comment owner %s for comment %s (new count: %d)", comment.AuthorID, req.TargetID, comment.Votes)

			// ✅ NEW: Reverse karma change
			if comment.AuthorID != userID {
				karmaDelta := 0
				if existingVote.VoteType == "up" {
					karmaDelta = -5 // Remove upvote = -5
				} else {
					karmaDelta = +1 // Remove downvote = +1 (reverse penalty)
				}

				err := s.profileService.UpdateKarma(ctx, comment.AuthorID, karmaDelta)
				if err != nil {
					log.Printf("⚠️ Failed to reverse karma for user %s: %v", comment.AuthorID, err)
				} else {
					log.Printf("✅ Karma reversed: User %s got %+d karma (comment unvote)", comment.AuthorID, karmaDelta)
				}
			}
		}
	}

	return nil
}

func (s *VoteServiceImpl) GetVote(ctx context.Context, userID uuid.UUID, targetID uuid.UUID, targetType string) (*dto.VoteResponse, error) {
	vote, err := s.voteRepo.GetVote(ctx, userID, targetID, targetType)
	if err != nil {
		return nil, err
	}

	if vote == nil {
		return nil, nil
	}

	return dto.VoteToVoteResponse(vote), nil
}

func (s *VoteServiceImpl) GetVoteCount(ctx context.Context, targetID uuid.UUID, targetType string) (*dto.VoteCountResponse, error) {
	upvotes, downvotes, err := s.voteRepo.GetVoteCount(ctx, targetID, targetType)
	if err != nil {
		return nil, err
	}

	return &dto.VoteCountResponse{
		TargetID:   targetID,
		TargetType: targetType,
		Upvotes:    upvotes,
		Downvotes:  downvotes,
		Total:      int(upvotes - downvotes),
	}, nil
}

func (s *VoteServiceImpl) GetUserVotes(ctx context.Context, userID uuid.UUID, targetType string, offset, limit int) ([]*dto.VoteResponse, error) {
	votes, err := s.voteRepo.ListByUser(ctx, userID, targetType, offset, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.VoteResponse, len(votes))
	for i, vote := range votes {
		responses[i] = dto.VoteToVoteResponse(vote)
	}

	return responses, nil
}

var _ services.VoteService = (*VoteServiceImpl)(nil)
