package controllers

import (
	"comproBackend/config"
	"comproBackend/middleware"
	"comproBackend/models"
	"net/http"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var existingUser models.User
	if result := config.DB.Where("username = ?", input.Username).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	if input.Password != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	var User models.User
	User.Username = input.Username
	User.Password = input.Password
	User.Role = "pending"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(User.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	User.Password = string(hashedPassword)

	if err := config.DB.Create(&User).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully, wait for verifier approval", "data": User})
}

func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var User models.User
	if result := config.DB.Where("username = ?", input.Username).First(&User); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if User.Role == "pending" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User registration is pending approval", "status": "pending"})
		return
	}

	if User.Role == "rejected" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User registration is rejected", "status": "rejected"})
		return
	}

	if User.NeedsReset {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User is waiting for password reset approval", "status": "reset_pending"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(User.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID:   User.ID,
		Username: User.Username,
		Role:     User.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middleware.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"data":    User,
	})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully."})
}

func GetPendingAndResetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Where("role = ? OR needs_reset = ?", "pending", true).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func ResetPassword(c *gin.Context) {
	var PasswordResetInput struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.ShouldBindJSON(&PasswordResetInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "username = ?", PasswordResetInput.Username).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role == "pending" || user.Role == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot reset password for pending or rejected users"})
		return
	}

	user.NeedsReset = true
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Username = PasswordResetInput.Username
	user.Password = PasswordResetInput.Password

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wait for verificator approval"})

}

func UpdateFCMToken(c *gin.Context) {
	var input struct {
		FCMToken string `json:"fcm_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.FCMToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "FCM token is required"})
		return
	}

	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.FCMToken = input.FCMToken
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FCM token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully"})
}

func Approval(c *gin.Context) {
	var input struct {
		NewRole string `json:"role"`
	}

	action := c.Query("action")
	userID, err := strconv.ParseUint(c.Query("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	_ = c.ShouldBindJSON(&input)

	if c.GetString("role") != "verificator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var user models.User

	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var message string
	if action == "reject" {
		if user.Role == "pending" {
			user.Role = "rejected"
			message = "Registration Reject Successful"
		} else {
			user.NeedsReset = false
			message = "Reset Request Reject Successful"
		}
		if err := config.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": message, "data": user})
		return
	}

	if input.NewRole != "" {
		user.Role = input.NewRole
	}

	user.NeedsReset = false
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User Approved successfully", "data": user})
}
