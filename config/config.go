package config

import (
	"fmt"
	"log"
	"main/models"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var JwtSecret string
var Port string
var DB *gorm.DB
var ExpoPushAPIURL string

func LoadConfig() {
	JwtSecret = os.Getenv("JWT_SECRET")
	Port = os.Getenv("PORT")

	ExpoPushAPIURL = os.Getenv("EXPO_PUSH_API_URL")

	if ExpoPushAPIURL == "" {
		ExpoPushAPIURL = "https://exp.host/--/api/v2/push/send"
	}

	if Port == "" {
		Port = ":9999"
	}

	if JwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

}

func ConnectDatabase() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	DB = db
	log.Println("database connected")
}

func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Auction{},
		&models.Bid{},
		&models.Watchlist{},
		&models.Notification{},
		&models.DevicePushToken{},
	)

	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	log.Println("database migrated")
}

func SeedCategories() {
	categories := []models.Category{
		{Name: "Electronics", Slug: "electronics"},
		{Name: "Fashion", Slug: "fashion"},
		{Name: "Home & Furniture", Slug: "home-furniture"},
		{Name: "Books", Slug: "books"},
		{Name: "Collectibles", Slug: "collectibles"},
		{Name: "Sports", Slug: "sports"},
		{Name: "Art", Slug: "art"},
		{Name: "Other", Slug: "other"},
	}

	for _, category := range categories {
		var existing models.Category

		err := DB.Where("slug = ?", category.Slug).First(&existing).Error
		if err == nil {
			continue
		}

		if err := DB.Create(&category).Error; err != nil {
			log.Println("failed to seed category:", category.Name, err)
		}
	}

	log.Println("categories seeded")
}
