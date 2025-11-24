package serviceimpl

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
)

type UserProfileServiceImpl struct {
	userProfileRepo repositories.UserProfileRepository
}

func NewUserProfileService(userProfileRepo repositories.UserProfileRepository) services.UserProfileService {
	return &UserProfileServiceImpl{
		userProfileRepo: userProfileRepo,
	}
}

func (s *UserProfileServiceImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	profile, err := s.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}
	return profile, nil
}

func (s *UserProfileServiceImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName, avatar, bio, location, website string) (*models.UserProfile, error) {
	// Get existing profile
	profile, err := s.userProfileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	// Update fields (now includes displayName and avatar)
	if displayName != "" {
		profile.DisplayName = displayName
	}
	if avatar != "" {
		profile.Avatar = avatar
	}
	profile.Bio = bio
	profile.Location = location
	profile.Website = website

	// Save
	err = s.userProfileRepo.Update(ctx, userID, profile)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *UserProfileServiceImpl) UpdateKarma(ctx context.Context, userID uuid.UUID, delta int) error {
	return s.userProfileRepo.UpdateKarma(ctx, userID, delta)
}

func (s *UserProfileServiceImpl) GetTopUsersByKarma(ctx context.Context, limit int) ([]*models.UserProfile, error) {
	return s.userProfileRepo.GetTopUsersByKarma(ctx, limit)
}
