package repository

import (
	"errors"
	"main/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WatchlistRepository struct {
	db *gorm.DB
}

func NewWatchlistRepository(db *gorm.DB) *WatchlistRepository {
	return &WatchlistRepository{db: db}
}

func (r *WatchlistRepository) Add(userID uint, auctionID uint) error {
	watchlist := models.Watchlist{
		UserID:    userID,
		AuctionID: auctionID,
	}

	return r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&watchlist).Error
}

func (r *WatchlistRepository) Remove(userID uint, auctionID uint) error {
	return r.db.
		Where("user_id = ? AND auction_id = ?", userID, auctionID).
		Delete(&models.Watchlist{}).Error
}

func (r *WatchlistRepository) Exists(userID uint, auctionID uint) (bool, error) {
	var watchlist models.Watchlist

	err := r.db.
		Where("user_id = ? AND auction_id = ?", userID, auctionID).
		First(&watchlist).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *WatchlistRepository) FindByUserID(userID uint) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist

	err := r.db.
		Preload("Auction").
		Preload("Auction.Product").
		Preload("Auction.Product.Category").
		Preload("Auction.Seller").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&watchlists).Error

	return watchlists, err
}
