package controllers

import (
	"MovieVerse/models"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func GetUsers(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		return
	}
}

func generateVerificationToken() (string, error) {
	// Generate a random 32-byte token
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	// Encode the token to base64 for safe storage
	return base64.URLEncoding.EncodeToString(token), nil
}

// CreateUser handles user signup and sends a verification email
func CreateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var user models.User

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Log the user data for debugging
	log.Printf("Incoming user data: %+v", user)

	// Check if email already exists
	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		respondWithError(w, http.StatusBadRequest, "Email already exists")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking for existing email: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error during email check")
		return
	}

	// Generate verification token
	verificationToken, err := generateVerificationToken()
	if err != nil {
		log.Printf("Verification token error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate verification token")
		return
	}
	user.VerificationToken = verificationToken

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.Password = string(hashedPassword) // Store the hashed password

	// Save user to the database
	if err := db.Create(&user).Error; err != nil {
		log.Printf("Error saving user to database: %v", err)
		return
	}

	// Send verification email
	if err := sendVerificationEmail(user.Email, user.VerificationToken); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to send verification email")
		return
	}

	// Respond to the client
	log.Printf("User %s created successfully", user.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created. Please check your email for verification.",
	})
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// VerifyEmail handles email verification
func VerifyEmail(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Check if email is already verified
	if user.EmailVerified {
		http.Error(w, "Email is already verified", http.StatusBadRequest)
		return
	}

	// Mark the user as verified
	user.EmailVerified = true
	user.VerificationToken = "" // Clear the token after verification

	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully. You can now log in.",
	})
}

// sendVerificationEmail sends a verification email with a token
func sendVerificationEmail(email, token string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "gamebeast66@gmail.com")
	mailer.SetHeader("To", email)
	mailer.SetHeader("Subject", "MovieVerse - Email Verification")
	verificationURL := "http://localhost:8080/verify-email?token=" + token
	mailer.SetBody("text/plain", "Please verify your email by clicking the link: "+verificationURL)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, "gamebeast66@gmail.com", "xfjv jsee cpcg rusr")
	if err := dialer.DialAndSend(mailer); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", email, err)
	}

	log.Printf("Verification email sent to %s", email)
	return nil
}
func LoginUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	var user models.User
	// Check if the user exists in the database
	if err := db.Where("email = ?", credentials.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			log.Printf("Error finding user: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate a token (use a proper library for real implementation)
	token := "exampleToken123" // Replace with a generated token (e.g., JWT or a UUID)

	// Set a cookie with the token
	http.SetCookie(w, &http.Cookie{
		Name:     "userToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"userId":  fmt.Sprintf("%d", user.ID),
	})
}

func GetUserByID(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}

func UpdateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var user models.User
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	if err := db.First(&user, id).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}

func DeleteUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	if err := db.Delete(&models.User{}, id).Error; err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
	if err != nil {
		return
	}
}
