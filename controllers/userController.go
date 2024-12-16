package controllers

import (
	"MovieVerse/models"
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
)

func GetReviews(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var reviews []models.Review
	if err := db.Preload("Movie").Preload("User").Find(&reviews).Error; err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(reviews)
	if err != nil {
		return
	}
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

func CreateReview(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var review models.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if err := db.Create(&review).Error; err != nil {
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(review)
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
