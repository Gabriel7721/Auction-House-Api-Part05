package models

import "time"

type Watchlist struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index;uniqueIndex:idx_watchlist_user_auction"`
	AuctionID uint      `json:"auction_id" gorm:"not null;index;uniqueIndex:idx_watchlist_user_auction"`
	CreatedAt time.Time `json:"created_at"`

	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Auction Auction `json:"auction,omitempty" gorm:"foreignKey:AuctionID"`
}
