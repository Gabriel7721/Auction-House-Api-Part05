package repository

import (
	"main/models"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) CreateTx(tx *gorm.DB, product *models.Product) error {
	return tx.Create(product).Error
}

func (r *ProductRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product

	err := r.db.
		Preload("Seller").
		Preload("Category").
		First(&product, id).Error

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) UpdateTx(tx *gorm.DB, product *models.Product) error {
	return tx.Save(product).Error
}
