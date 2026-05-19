package repository

import (
	"errors"
	"main/models"

	"gorm.io/gorm"
)

type BidRepository struct {
	db *gorm.DB
}

func NewBidRepository(db *gorm.DB) *BidRepository {
	return &BidRepository{db: db}
}

func (r *BidRepository) Create(bid *models.Bid) error {
	return r.db.Create(bid).Error
}

func (r *BidRepository) CreateTx(tx *gorm.DB, bid *models.Bid) error {
	return tx.Create(bid).Error
}

func (r *BidRepository) FindByAuctionID(auctionID uint) ([]models.Bid, error) {
	var bids []models.Bid

	err := r.db.
		Preload("User").
		Where("auction_id = ?", auctionID).
		Order("amount DESC").
		Order("created_at DESC").
		Find(&bids).Error

	return bids, err
}

func (r *BidRepository) FindHighestByAuctionID(auctionID uint) (*models.Bid, error) {
	var bid models.Bid

	err := r.db.
		Where("auction_id = ?", auctionID).
		Order("amount DESC").
		Order("created_at ASC").
		First(&bid).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &bid, nil
}

func (r *BidRepository) FindHighestByAuctionIDTx(tx *gorm.DB, auctionID uint) (*models.Bid, error) {
	var bid models.Bid

	err := tx.
		Where("auction_id = ?", auctionID).
		Order("amount DESC").
		Order("created_at ASC").
		First(&bid).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &bid, nil
}

func (r *BidRepository) FindByUserID(userID uint) ([]models.Bid, error) {
	var bids []models.Bid

	err := r.db.
		Preload("User").
		Preload("Auction").
		Preload("Auction.Product").
		Preload("Auction.Seller").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bids).Error

	return bids, err
}
