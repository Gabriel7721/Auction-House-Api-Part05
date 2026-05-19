package services

import (
	"errors"
	"main/dto"
	"main/models"
	"main/repository"
	"strings"
)

type PushTokenService struct {
	pushTokenRepo *repository.DevicePushTokenRepository
}

func NewPushTokenService(
	pushTokenRepo *repository.DevicePushTokenRepository,
) *PushTokenService {
	return &PushTokenService{
		pushTokenRepo: pushTokenRepo,
	}
}

func (s *PushTokenService) RegisterPushToken(
	userID uint,
	input dto.RegisterPushTokenRequest,
) (*models.DevicePushToken, error) {
	token := strings.TrimSpace(input.ExpoPushToken)

	if token == "" {
		return nil, errors.New("expo_push_token is required")
	}

	platform := normalizeDevicePlatform(input.Platform)

	return s.pushTokenRepo.Upsert(userID, token, platform)
}

func (s *PushTokenService) UnregisterPushToken(
	userID uint,
	input dto.UnregisterPushTokenRequest,
) error {
	token := strings.TrimSpace(input.ExpoPushToken)

	if token == "" {
		return errors.New("expo_push_token is required")
	}

	return s.pushTokenRepo.DeactivateByUserAndToken(userID, token)
}

func normalizeDevicePlatform(platform string) models.DevicePlatform {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "android":
		return models.DevicePlatformAndroid
	case "ios":
		return models.DevicePlatformIOS
	default:
		return models.DevicePlatformUnknown
	}
}
