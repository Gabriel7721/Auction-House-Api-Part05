package dto

import "time"

type CreateAuctionRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	CategoryID  uint     `json:"category_id" binding:"required"`
	Condition   string   `json:"condition" binding:"required,oneof=new used refurbished"`
	Images      []string `json:"images"`

	StartPrice   float64   `json:"start_price" binding:"required,gt=0"`
	ReservePrice *float64  `json:"reserve_price"`
	BuyNowPrice  *float64  `json:"buy_now_price"`
	BidIncrement float64   `json:"bid_increment" binding:"required,gt=0"`
	StartsAt     time.Time `json:"starts_at" binding:"required"`
	EndsAt       time.Time `json:"ends_at" binding:"required"`
}

type UpdateAuctionRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	CategoryID  *uint    `json:"category_id"`
	Condition   *string  `json:"condition"`
	Images      []string `json:"images"`

	ReservePrice *float64   `json:"reserve_price"`
	BuyNowPrice  *float64   `json:"buy_now_price"`
	BidIncrement *float64   `json:"bid_increment"`
	StartsAt     *time.Time `json:"starts_at"`
	EndsAt       *time.Time `json:"ends_at"`
}

type PlaceBidRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type AuctionQuery struct {
	Page       int     `form:"page"`
	Limit      int     `form:"limit"`
	Search     string  `form:"search"`
	CategoryID uint    `form:"category_id"`
	Status     string  `form:"status"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	EndingSoon bool    `form:"ending_soon"`
	SellerID   uint    `form:"seller_id"`
}

type PaginatedResponse struct {
	Data       any   `json:"data"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
