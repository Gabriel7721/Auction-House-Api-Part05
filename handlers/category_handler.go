package handlers

import (
	"main/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}
