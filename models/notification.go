package models

import "time"

type NotificationType string

const (
	NotificationTypeBidPlaced      NotificationType = "bid_placed"
	NotificationTypeOutbid         NotificationType = "outbid"
	NotificationTypeAuctionWon     NotificationType = "auction_won"
	NotificationTypeAuctionEnded   NotificationType = "auction_ended"
	NotificationTypeAuctionWatched NotificationType = "auction_watched"
)

type Notification struct {
	ID        uint             `json:"id" gorm:"primaryKey"`
	UserID    uint             `json:"user_id" gorm:"not null;index"`
	Type      NotificationType `json:"type" gorm:"type:varchar(50);not null;index"`
	Title     string           `json:"title" gorm:"size:200;not null"`
	Message   string           `json:"message" gorm:"type:text;not null"`
	IsRead    bool             `json:"is_read" gorm:"not null;default:false;index"`
	CreatedAt time.Time        `json:"created_at" gorm:"index"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
