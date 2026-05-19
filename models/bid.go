package models

import "time"

type Bid struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AuctionID uint      `json:"auction_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Amount    float64   `json:"amount" gorm:"type:decimal(12,2);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	Auction Auction `json:"auction,omitempty" gorm:"foreignKey:AuctionID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
