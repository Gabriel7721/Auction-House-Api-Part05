package models

import "time"

type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:120;not null"`
	Slug      string    `json:"slug" gorm:"size:150;not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Products []Product `json:"-" gorm:"foreignKey:CategoryID"`
}
