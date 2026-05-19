package repository

import (
	"main/dto"
	"main/models"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuctionRepository struct {
	db *gorm.DB
}

func NewAuctionRepository(db *gorm.DB) *AuctionRepository {
	return &AuctionRepository{db: db}
}

func (r *AuctionRepository) Create(auction *models.Auction) error {
	return r.db.Create(auction).Error
}

func (r *AuctionRepository) CreateTx(tx *gorm.DB, auction *models.Auction) error {
	return tx.Create(auction).Error
}

func (r *AuctionRepository) FindByID(id uint) (*models.Auction, error) {
	var auction models.Auction

	err := r.db.
		Preload("Product").
		Preload("Product.Category").
		Preload("Seller").
		Preload("Winner").
		First(&auction, id).Error

	if err != nil {
		return nil, err
	}

	return &auction, nil
}

func (r *AuctionRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*models.Auction, error) {
	var auction models.Auction

	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Product").
		Preload("Seller").
		First(&auction, id).Error

	if err != nil {
		return nil, err
	}

	return &auction, nil
}

func (r *AuctionRepository) List(query dto.AuctionQuery) ([]models.Auction, int64, error) {
	var auctions []models.Auction
	var total int64

	page := query.Page
	limit := query.Limit

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	dbQuery := r.db.
		Model(&models.Auction{}).
		Joins("JOIN products ON products.id = auctions.product_id")

	if query.Search != "" {
		search := "%" + query.Search + "%"

		dbQuery = dbQuery.Where(
			"products.title LIKE ? OR products.description LIKE ?",
			search,
			search,
		)
	}

	if query.CategoryID > 0 {
		dbQuery = dbQuery.Where("products.category_id = ?", query.CategoryID)
	}

	if query.Status != "" {
		dbQuery = dbQuery.Where("auctions.status = ?", query.Status)
	}

	if query.MinPrice > 0 {
		dbQuery = dbQuery.Where("auctions.current_price >= ?", query.MinPrice)
	}

	if query.MaxPrice > 0 {
		dbQuery = dbQuery.Where("auctions.current_price <= ?", query.MaxPrice)
	}

	if query.SellerID > 0 {
		dbQuery = dbQuery.Where("auctions.seller_id = ?", query.SellerID)
	}

	if query.EndingSoon {
		now := time.Now()
		next24Hours := now.Add(24 * time.Hour)

		dbQuery = dbQuery.Where(
			"auctions.ends_at BETWEEN ? AND ?",
			now,
			next24Hours,
		)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbQuery.
		Preload("Product").
		Preload("Product.Category").
		Preload("Seller").
		Preload("Winner").
		Order("auctions.ends_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&auctions).Error

	if err != nil {
		return nil, 0, err
	}

	return auctions, total, nil
}

func (r *AuctionRepository) Update(auction *models.Auction) error {
	return r.db.Save(auction).Error
}

func (r *AuctionRepository) UpdateTx(tx *gorm.DB, auction *models.Auction) error {
	return tx.Save(auction).Error
}

func (r *AuctionRepository) FindBySellerID(sellerID uint) ([]models.Auction, error) {
	var auctions []models.Auction

	err := r.db.
		Preload("Product").
		Preload("Product.Category").
		Preload("Winner").
		Where("seller_id = ?", sellerID).
		Order("created_at DESC").
		Find(&auctions).Error

	return auctions, err
}

func (r *AuctionRepository) FindExpiredLiveAuctions() ([]models.Auction, error) {
	var auctions []models.Auction

	err := r.db.
		Where("status IN ?", []models.AuctionStatus{
			models.AuctionStatusLive,
			models.AuctionStatusScheduled,
		}).
		Where("ends_at <= ?", time.Now()).
		Find(&auctions).Error

	return auctions, err
}

func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}

	return int(math.Ceil(float64(total) / float64(limit)))
}

func (r *AuctionRepository) ActivateReadyScheduledAuctions(now time.Time) error {
	return r.db.
		Model(&models.Auction{}).
		Where(
			"status = ? AND starts_at <= ? AND ends_at > ?",
			models.AuctionStatusScheduled,
			now,
			now,
		).
		Update("status", models.AuctionStatusLive).
		Error
}
