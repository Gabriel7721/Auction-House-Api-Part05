package models

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:120;not null"`
	Email        string    `json:"email" gorm:"size:191;not null;uniqueIndex"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;size:255;not null"`
	Phone        string    `json:"phone,omitempty" gorm:"size:30"`
	AvatarURL    string    `json:"avatar_url,omitempty" gorm:"size:500"`
	Role         UserRole  `json:"role" gorm:"type:varchar(20);not null;default:'user';index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
