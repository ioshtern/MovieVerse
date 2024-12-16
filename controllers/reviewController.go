package controllers

import (
	"MovieVerse/models"
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
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

func GetReviewByID(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	var review models.Review
	if err := db.Preload("Movie").Preload("User").First(&review, id).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(review)
	if err != nil {
		return
	}
}

func CreateUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if err := db.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}

func UpdateReview(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var review models.Review
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}
	if err := db.First(&review, id).Error; err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := db.Save(&review).Error; err != nil {
		http.Error(w, "Failed to update review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(review)
	if err != nil {
		return
	}
}

func DeleteReview(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}
	if err := db.Delete(&models.Review{}, id).Error; err != nil {
		http.Error(w, "Failed to delete review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"message": "Review deleted successfully"})
	if err != nil {
		return
	}
}
