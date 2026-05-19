package models

import "time"

type ProductCondition string
type ProductStatus string

const (
	ProductConditionNew         ProductCondition = "new"
	ProductConditionUsed        ProductCondition = "used"
	ProductConditionRefurbished ProductCondition = "refurbished"

	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusActive    ProductStatus = "active"
	ProductStatusSold      ProductStatus = "sold"
	ProductStatusCancelled ProductStatus = "cancelled"
)

type Product struct {
	ID          uint             `json:"id" gorm:"primaryKey"`
	SellerID    uint             `json:"seller_id" gorm:"not null;index"`
	CategoryID  uint             `json:"category_id" gorm:"not null;index"`
	Title       string           `json:"title" gorm:"size:200;not null;index"`
	Description string           `json:"description" gorm:"type:text;not null"`
	Images      StringArray      `json:"images" gorm:"type:json"`
	Condition   ProductCondition `json:"condition" gorm:"type:varchar(30);not null;index"`
	Status      ProductStatus    `json:"status" gorm:"type:varchar(30);not null;default:'draft';index"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

	Seller   User     `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
	Category Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Auction  *Auction `json:"auction,omitempty" gorm:"foreignKey:ProductID"`
}
