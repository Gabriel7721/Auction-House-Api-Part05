package models

import "time"

type AuctionStatus string

const (
	AuctionStatusScheduled AuctionStatus = "scheduled"
	AuctionStatusLive      AuctionStatus = "live"
	AuctionStatusEnded     AuctionStatus = "ended"
	AuctionStatusCancelled AuctionStatus = "cancelled"
)

type Auction struct {
	ID           uint          `json:"id" gorm:"primaryKey"`
	ProductID    uint          `json:"product_id" gorm:"not null;uniqueIndex"`
	SellerID     uint          `json:"seller_id" gorm:"not null;index"`
	StartPrice   float64       `json:"start_price" gorm:"type:decimal(12,2);not null"`
	ReservePrice *float64      `json:"reserve_price,omitempty" gorm:"type:decimal(12,2)"`
	BuyNowPrice  *float64      `json:"buy_now_price,omitempty" gorm:"type:decimal(12,2)"`
	CurrentPrice float64       `json:"current_price" gorm:"type:decimal(12,2);not null;default:0"`
	BidIncrement float64       `json:"bid_increment" gorm:"type:decimal(12,2);not null"`
	StartsAt     time.Time     `json:"starts_at" gorm:"not null;index"`
	EndsAt       time.Time     `json:"ends_at" gorm:"not null;index"`
	Status       AuctionStatus `json:"status" gorm:"type:varchar(30);not null;default:'scheduled';index"`
	WinnerID     *uint         `json:"winner_id,omitempty" gorm:"index"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	Product    Product     `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Seller     User        `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
	Winner     *User       `json:"winner,omitempty" gorm:"foreignKey:WinnerID"`
	Bids       []Bid       `json:"bids,omitempty" gorm:"foreignKey:AuctionID"`
	Watchlists []Watchlist `json:"-" gorm:"foreignKey:AuctionID"`
}
