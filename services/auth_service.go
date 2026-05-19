package services

import (
	"errors"
	"main/dto"
	"main/models"
	"main/repository"
	"main/utils"
	"strings"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(input dto.RegisterInput) (models.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	name := strings.TrimSpace(input.Name)

	_, err := s.userRepo.FindByEmail(input.Email)
	if err == nil {
		return models.User{}, errors.New("email already exists")
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         models.RoleUser,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return models.User{}, err
	}

	return user, nil

}

func (s *AuthService) Login(input dto.LoginInput) (string, models.User, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", models.User{}, errors.New("invalid email or password")
	}

	if err := utils.ComparePassword(user.PasswordHash, input.Password); err != nil {
		return "", models.User{}, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", models.User{}, err
	}

	return token, *user, nil
}

func (s *AuthService) GetMe(userID uint) (models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return models.User{}, err
	}
	return *user, nil
}
