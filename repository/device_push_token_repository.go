package repository

import (
	"errors"
	"main/models"
	"time"

	"gorm.io/gorm"
)

type DevicePushTokenRepository struct {
	db *gorm.DB
}

func NewDevicePushTokenRepository(db *gorm.DB) *DevicePushTokenRepository {
	return &DevicePushTokenRepository{
		db: db,
	}
}

func (r *DevicePushTokenRepository) Upsert(
	userID uint,
	expoPushToken string,
	platform models.DevicePlatform,
) (*models.DevicePushToken, error) {
	var existing models.DevicePushToken

	err := r.db.
		Where("expo_push_token = ?", expoPushToken).
		First(&existing).Error

	if err == nil {
		existing.UserID = userID
		existing.Platform = platform
		existing.IsActive = true
		existing.LastSeenAt = time.Now()

		if err := r.db.Save(&existing).Error; err != nil {
			return nil, err
		}

		return &existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	token := models.DevicePushToken{
		UserID:        userID,
		ExpoPushToken: expoPushToken,
		Platform:      platform,
		IsActive:      true,
		LastSeenAt:    time.Now(),
	}

	if err := r.db.Create(&token).Error; err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *DevicePushTokenRepository) DeactivateByUserAndToken(
	userID uint,
	expoPushToken string,
) error {
	return r.db.
		Model(&models.DevicePushToken{}).
		Where("user_id = ? AND expo_push_token = ?", userID, expoPushToken).
		Updates(map[string]any{
			"is_active": false,
		}).Error
}

func (r *DevicePushTokenRepository) FindActiveByUserID(
	userID uint,
) ([]models.DevicePushToken, error) {
	var tokens []models.DevicePushToken

	err := r.db.
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("updated_at DESC").
		Find(&tokens).Error

	return tokens, err
}

func (r *DevicePushTokenRepository) MarkInactiveByExpoPushToken(
	expoPushToken string,
) error {
	return r.db.
		Model(&models.DevicePushToken{}).
		Where("expo_push_token = ?", expoPushToken).
		Update("is_active", false).Error
}
