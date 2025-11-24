package postgres

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) GetByOAuth(ctx context.Context, provider, oauthID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("oauth_provider = ? AND oauth_id = ?", provider, oauthID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Update(ctx context.Context, id uuid.UUID, user *models.User) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Updates(user).Error
}

func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.User{}).Error
}

func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

func (r *UserRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

// SearchForChat searches users for chat (excludes self, blocked users)
func (r *UserRepositoryImpl) SearchForChat(ctx context.Context, currentUserID uuid.UUID, query string, limit int) ([]*models.User, error) {
	var users []*models.User

	// Build subquery for blocked users
	blockedUsersSubquery := r.db.Table("blocks").
		Select("CASE WHEN blocker_id = ? THEN blocked_id WHEN blocked_id = ? THEN blocker_id END as user_id", currentUserID, currentUserID).
		Where("blocker_id = ? OR blocked_id = ?", currentUserID, currentUserID)

	// Join with user_profiles to access display_name and karma
	err := r.db.WithContext(ctx).
		Table("users_identity").
		Select("users_identity.*").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users_identity.id").
		Where("users_identity.id != ?", currentUserID).
		Where("users_identity.deleted_at IS NULL").
		Where(
			r.db.Where("LOWER(users_identity.username) LIKE ?", "%"+query+"%").
				Or("LOWER(user_profiles.display_name) LIKE ?", "%"+query+"%"),
		).
		// Exclude blocked users
		Where("users_identity.id NOT IN (?)", blockedUsersSubquery).
		// Order by: followers first, then following, then by karma
		Order(r.db.Raw("(SELECT COUNT(*) FROM follows WHERE follower_id = users_identity.id AND following_id = ?) DESC", currentUserID)).
		Order(r.db.Raw("(SELECT COUNT(*) FROM follows WHERE follower_id = ? AND following_id = users_identity.id) DESC", currentUserID)).
		Order("user_profiles.karma DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}

// GetSuggestedForChat gets suggested users for chat (followers, following, popular users)
func (r *UserRepositoryImpl) GetSuggestedForChat(ctx context.Context, currentUserID uuid.UUID, limit int) ([]*models.User, error) {
	var users []*models.User

	// Build subquery for blocked users
	blockedUsersSubquery := r.db.Table("blocks").
		Select("CASE WHEN blocker_id = ? THEN blocked_id WHEN blocked_id = ? THEN blocker_id END as user_id", currentUserID, currentUserID).
		Where("blocker_id = ? OR blocked_id = ?", currentUserID, currentUserID)

	// Join with user_profiles to access karma
	err := r.db.WithContext(ctx).
		Table("users_identity").
		Select("users_identity.*").
		Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users_identity.id").
		Where("users_identity.id != ?", currentUserID).
		Where("users_identity.deleted_at IS NULL").
		Where(
			r.db.Where("users_identity.id IN (?)",
				r.db.Table("follows").
					Select("follower_id").
					Where("following_id = ?", currentUserID),
			).Or("users_identity.id IN (?)",
				r.db.Table("follows").
					Select("following_id").
					Where("follower_id = ?", currentUserID),
			),
		).
		// Exclude blocked users
		Where("users_identity.id NOT IN (?)", blockedUsersSubquery).
		// Order by recent conversations first (if exists), then by karma
		Order(r.db.Raw("(SELECT MAX(last_message_at) FROM conversations WHERE user1_id = users_identity.id OR user2_id = users_identity.id) DESC NULLS LAST")).
		Order("user_profiles.karma DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}
