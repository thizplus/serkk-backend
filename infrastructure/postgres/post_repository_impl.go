package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostRepositoryImpl struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) repositories.PostRepository {
	return &PostRepositoryImpl{db: db}
}

func (r *PostRepositoryImpl) Create(ctx context.Context, post *models.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	var post models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			// Order media by display_order from post_media junction table
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Where("post_media.post_id = ?", id).
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Preload("SourcePost").
		Preload("SourcePost.Author").
		Preload("SourcePost.Media", func(db *gorm.DB) *gorm.DB {
			// Order source post media as well
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("SourcePost.Tags").
		Where("id = ? AND is_deleted = ?", id, false).
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepositoryImpl) Update(ctx context.Context, id uuid.UUID, post *models.Post) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Updates(post).Error
}

func (r *PostRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": now,
		}).Error
}

func (r *PostRepositoryImpl) List(ctx context.Context, offset, limit int, sortBy repositories.PostSortBy) ([]*models.Post, error) {
	var posts []*models.Post
	query := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			// Order media by display_order
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Preload("SourcePost").
		Preload("SourcePost.Author").
		Preload("SourcePost.Media", func(db *gorm.DB) *gorm.DB {
			// Order source post media
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("SourcePost.Tags").
		Where("is_deleted = ? AND status = ?", false, "published")

	switch sortBy {
	case repositories.SortByHot:
		// Hot score: votes / (hours + 2)^1.5
		query = query.Order(r.hotScoreSQL() + " DESC")
	case repositories.SortByNew:
		query = query.Order("created_at DESC")
	case repositories.SortByTop:
		query = query.Order("votes DESC")
	case repositories.SortByControversial:
		// High comment count but mixed votes
		query = query.Order("comment_count DESC, ABS(votes) DESC")
	default:
		query = query.Order("created_at DESC")
	}

	err := query.Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) ListByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Preload("SourcePost").
		Preload("SourcePost.Author").
		Preload("SourcePost.Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("SourcePost.Tags").
		Where("author_id = ? AND is_deleted = ?", authorID, false).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) ListByTag(ctx context.Context, tagName string, offset, limit int, sortBy repositories.PostSortBy) ([]*models.Post, error) {
	var posts []*models.Post

	// Debug logging
	log.Printf("🔍 Repository searching for tag: '%s'", tagName)

	query := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Preload("SourcePost").
		Preload("SourcePost.Author").
		Preload("SourcePost.Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("SourcePost.Tags").
		Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Joins("JOIN tags ON tags.id = post_tags.tag_id").
		Where("LOWER(TRIM(tags.name)) = LOWER(TRIM(?)) AND posts.is_deleted = ? AND posts.status = ?", tagName, false, "published")

	switch sortBy {
	case repositories.SortByHot:
		query = query.Order(r.hotScoreSQL() + " DESC")
	case repositories.SortByNew:
		query = query.Order("posts.created_at DESC")
	case repositories.SortByTop:
		query = query.Order("posts.votes DESC")
	default:
		query = query.Order("posts.created_at DESC")
	}

	err := query.Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) ListByTagID(ctx context.Context, tagID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy) ([]*models.Post, error) {
	var posts []*models.Post
	query := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Preload("SourcePost").
		Preload("SourcePost.Author").
		Preload("SourcePost.Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("SourcePost.Tags").
		Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Where("post_tags.tag_id = ? AND posts.is_deleted = ? AND posts.status = ?", tagID, false, "published")

	switch sortBy {
	case repositories.SortByHot:
		query = query.Order(r.hotScoreSQL() + " DESC")
	case repositories.SortByNew:
		query = query.Order("posts.created_at DESC")
	case repositories.SortByTop:
		query = query.Order("posts.votes DESC")
	default:
		query = query.Order("posts.created_at DESC")
	}

	err := query.Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) Search(ctx context.Context, query string, offset, limit int) ([]*models.Post, error) {
	var posts []*models.Post
	searchQuery := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Preload("SourcePost").
		Preload("SourcePost.Author").
		Preload("SourcePost.Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("SourcePost.Tags").
		Where(`is_deleted = ? AND status = ? AND (
			title ILIKE ? OR
			content ILIKE ? OR
			EXISTS (
				SELECT 1 FROM post_tags
				JOIN tags ON tags.id = post_tags.tag_id
				WHERE post_tags.post_id = posts.id
				AND tags.name ILIKE ?
			)
		)`, false, "published", searchQuery, searchQuery, searchQuery).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) GetCrossposts(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN post_media ON post_media.media_id = media.id").
				Order("post_media.display_order ASC")
		}).
		Preload("Tags").
		Where("source_post_id = ? AND is_deleted = ? AND status = ?", postID, false, "published").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Post{}).Where("is_deleted = ? AND status = ?", false, "published").Count(&count).Error
	return count, err
}

func (r *PostRepositoryImpl) CountByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("author_id = ? AND is_deleted = ?", authorID, false).
		Count(&count).Error
	return count, err
}

func (r *PostRepositoryImpl) IncrementCommentCount(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
}

func (r *PostRepositoryImpl) DecrementCommentCount(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error
}

func (r *PostRepositoryImpl) UpdateVoteCount(ctx context.Context, postID uuid.UUID, voteChange int) error {
	return r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("votes", gorm.Expr("votes + ?", voteChange)).Error
}

func (r *PostRepositoryImpl) AttachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error {
	// Insert into post_media junction table with display_order
	// Order is determined by the position in the mediaIDs array
	for i, mediaID := range mediaIDs {
		postMedia := &models.PostMedia{
			PostID:       postID,
			MediaID:      mediaID,
			DisplayOrder: i, // Index in array = display order
		}

		// Use FirstOrCreate to avoid duplicates (if media already attached)
		err := r.db.WithContext(ctx).
			Where("post_id = ? AND media_id = ?", postID, mediaID).
			FirstOrCreate(postMedia).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostRepositoryImpl) DetachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error {
	// Delete from post_media junction table
	return r.db.WithContext(ctx).
		Where("post_id = ? AND media_id IN ?", postID, mediaIDs).
		Delete(&models.PostMedia{}).Error
}

func (r *PostRepositoryImpl) GetPostsByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*models.Post, error) {
	var posts []*models.Post
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media").
		Preload("Tags").
		Joins("JOIN post_media ON posts.id = post_media.post_id").
		Where("post_media.media_id = ? AND posts.is_deleted = ?", mediaID, false).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepositoryImpl) AttachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	post := &models.Post{ID: postID}
	var tagList []models.Tag
	for _, tagID := range tagIDs {
		tagList = append(tagList, models.Tag{ID: tagID})
	}
	return r.db.WithContext(ctx).Model(post).Association("Tags").Append(tagList)
}

func (r *PostRepositoryImpl) DetachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	post := &models.Post{ID: postID}
	var tagList []models.Tag
	for _, tagID := range tagIDs {
		tagList = append(tagList, models.Tag{ID: tagID})
	}
	return r.db.WithContext(ctx).Model(post).Association("Tags").Delete(tagList)
}

func (r *PostRepositoryImpl) SyncTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	post := &models.Post{ID: postID}
	var tagList []models.Tag
	for _, tagID := range tagIDs {
		tagList = append(tagList, models.Tag{ID: tagID})
	}
	return r.db.WithContext(ctx).Model(post).Association("Tags").Replace(tagList)
}

// hotScoreSQL generates SQL for hot score calculation: votes / (hours + 2)^1.5
func (r *PostRepositoryImpl) hotScoreSQL() string {
	// Calculate hours since post creation
	// Hot score = votes / (hours_since_creation + 2)^1.5
	return fmt.Sprintf(
		"posts.votes / POWER((EXTRACT(EPOCH FROM (NOW() - posts.created_at)) / 3600.0) + 2, %.1f)",
		1.5,
	)
}

// Compiler check to ensure PostRepositoryImpl implements PostRepository
var _ repositories.PostRepository = (*PostRepositoryImpl)(nil)
