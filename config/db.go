package config

import (
	"comproBackend/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

	dsn := "root:@tcp(127.0.0.1:3306)/comprodb?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(&models.Log{}, &models.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = db
	fmt.Println("Database connected successfully")
}
