package repository

import (
	"main/models"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) CreateTx(tx *gorm.DB, notification *models.Notification) error {
	return tx.Create(notification).Error
}

func (r *NotificationRepository) FindByUserID(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification

	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error

	return notifications, err
}

func (r *NotificationRepository) FindUnreadCount(userID uint) (int64, error) {
	var count int64

	err := r.db.
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error

	return count, err
}

func (r *NotificationRepository) FindByIDAndUserID(notificationID uint, userID uint) (*models.Notification, error) {
	var notification models.Notification

	err := r.db.
		Where("id = ? AND user_id = ?", notificationID, userID).
		First(&notification).Error

	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *NotificationRepository) MarkAsRead(notification *models.Notification) error {
	notification.IsRead = true
	return r.db.Save(notification).Error
}

func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	return r.db.
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
