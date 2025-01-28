package controllers

import (
	"MovieVerse/models"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
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
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func CreateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	log.Printf("Incoming user data: %+v", user)

	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		respondWithError(w, http.StatusBadRequest, "Email already exists")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking for existing email: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal server error during email check")
		return
	}

	verificationToken, err := generateVerificationToken()
	if err != nil {
		log.Printf("Verification token error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate verification token")
		return
	}
	user.VerificationToken = verificationToken

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.Password = string(hashedPassword)

	if err := db.Create(&user).Error; err != nil {
		log.Printf("Error saving user to database: %v", err)
		return
	}

	if err := sendVerificationEmail(user.Email, user.VerificationToken); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to send verification email")
		return
	}

	log.Printf("User %s created successfully", user.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created. Please check your email for verification.",
	})
}

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

	if user.EmailVerified {
		http.Error(w, "Email is already verified", http.StatusBadRequest)
		return
	}
	user.EmailVerified = true
	user.VerificationToken = ""

	if err := db.Save(&user).Error; err != nil {
		http.Error(w, "Failed to verify user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully. You can now log in.",
	})
}

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

var jwtKey = []byte("your_secret_key")

// Claims structure for JWT tokens
type Claims struct {
	UserID uint `json:"userId"`
	Admin  bool `json:"admin"`
	jwt.RegisteredClaims
}

// LoginUser handles user login and issues a JWT
func LoginUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Query the user from the database
	var user models.User
	if err := db.Where("email = ?", credentials.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		} else {
			log.Printf("Error finding user: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Validate the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Create a JWT token for the user
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Admin:  user.Admin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Failed to sign token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Respond with the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Strip "Bearer " prefix if present
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		// Parse and validate the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add claims to the request context
		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user claims from the context
		claims, ok := r.Context().Value("user").(*Claims)
		if !ok || !claims.Admin {
			http.Error(w, "Access denied: Admins only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*Claims)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Protected content accessed",
		"userId":  claims.UserID,
		"admin":   claims.Admin,
	})
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
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
