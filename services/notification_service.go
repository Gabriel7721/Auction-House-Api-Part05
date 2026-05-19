package services

import (
	"errors"
	"main/models"
	"main/repository"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

func (s *NotificationService) GetMyNotifications(userID uint) ([]models.Notification, error) {
	return s.notificationRepo.FindByUserID(userID)
}

func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	return s.notificationRepo.FindUnreadCount(userID)
}

func (s *NotificationService) MarkAsRead(userID uint, notificationID uint) (*models.Notification, error) {
	notification, err := s.notificationRepo.FindByIDAndUserID(notificationID, userID)
	if err != nil {
		return nil, errors.New("notification not found")
	}

	if notification.IsRead {
		return notification, nil
	}

	if err := s.notificationRepo.MarkAsRead(notification); err != nil {
		return nil, err
	}

	return notification, nil
}

func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}
