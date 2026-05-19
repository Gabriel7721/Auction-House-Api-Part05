package services

import (
	"main/dto"
	"main/repository"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *CategoryService) ListCategories() ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.CategoryResponse, 0, len(categories))

	for _, category := range categories {
		responses = append(responses, dto.CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
			Slug: category.Slug,
		})
	}

	return responses, nil
}
