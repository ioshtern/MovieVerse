package controllers

import (
	"MovieVerse/models"
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
)

func GetMovies(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var movies []models.Movie
	if err := db.Find(&movies).Error; err != nil {
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(movies)
	if err != nil {
		return
	}
}

func GetMovieByID(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	var movie models.Movie
	if err := db.First(&movie, id).Error; err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(movie)
	if err != nil {
		return
	}
}

func CreateMovie(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if err := db.Create(&movie).Error; err != nil {
		http.Error(w, "Failed to create movie", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(movie)
	if err != nil {
		return
	}
}

func UpdateMovie(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}
	if err := db.First(&movie, id).Error; err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := db.Save(&movie).Error; err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(movie)
	if err != nil {
		return
	}
}

func DeleteMovie(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}
	if err := db.Delete(&models.Movie{}, id).Error; err != nil {
		http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"message": "Movie deleted successfully"})
	if err != nil {
		return
	}
}
