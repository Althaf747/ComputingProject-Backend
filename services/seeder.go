package services

import (
	"comproBackend/config"
	"comproBackend/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// UATTestUsers - seed test users for UAT testing
func SeedUATUsers() {
	testUsers := []struct {
		Username string
		Password string
		Role     string
	}{
		{"verifier_test", "Test@123", "verificator"},
		{"user_test", "Test@123", "user"},
		{"pending_test", "Test@123", "pending"},
		{"reject_test", "Test@123", "pending"},
	}

	for _, u := range testUsers {
		// Check if user already exists
		var existing models.User
		result := config.DB.Where("username = ?", u.Username).First(&existing)
		if result.RowsAffected > 0 {
			log.Printf("[SEEDER] User '%s' already exists, skipping", u.Username)
			continue
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[SEEDER] Failed to hash password for '%s': %v", u.Username, err)
			continue
		}

		// Create user
		user := models.User{
			Username:   u.Username,
			Password:   string(hashedPassword),
			Role:       u.Role,
			NeedsReset: false,
		}

		if err := config.DB.Create(&user).Error; err != nil {
			log.Printf("[SEEDER] Failed to create user '%s': %v", u.Username, err)
		} else {
			log.Printf("[SEEDER] ✅ Created user '%s' (role: %s)", u.Username, u.Role)
		}
	}
}
