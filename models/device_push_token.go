package models

import "time"

type DevicePlatform string

const (
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformUnknown DevicePlatform = "unknown"
)

type DevicePushToken struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index"`
	ExpoPushToken string         `json:"expo_push_token" gorm:"size:512;not null;uniqueIndex"`
	Platform      DevicePlatform `json:"platform" gorm:"type:varchar(30);not null;default:'unknown';index"`
	IsActive      bool           `json:"is_active" gorm:"not null;default:true;index"`
	LastSeenAt    time.Time      `json:"last_seen_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	User User `json:"-" gorm:"foreignKey:UserID"`
}
