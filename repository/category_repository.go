package repository

import (
	"main/models"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) FindAll() ([]models.Category, error) {
	var categories []models.Category

	err := r.db.
		Order("name ASC").
		Find(&categories).Error

	return categories, err
}

func (r *CategoryRepository) FindByID(id uint) (*models.Category, error) {
	var category models.Category

	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}

	return &category, nil
}
