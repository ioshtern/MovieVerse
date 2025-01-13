package controllers

import (
	"MovieVerse/models"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
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
	var inputMap map[string]interface{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &inputMap); err != nil {
		http.Error(w, "Invalid attribute encountered", http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(body, &movie); err != nil {
		http.Error(w, "Invalid data type", http.StatusBadRequest)
		return
	}

	for key := range inputMap {
		switch key {
		case "title", "release_year", "genres", "director", "country", "description":
		default:
			http.Error(w, "Invalid attribute: "+key, http.StatusBadRequest)
			return
		}
	}

	if err := db.Create(&movie).Error; err != nil {
		http.Error(w, "Failed to create movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(movie); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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
	var updatedMovie models.Movie
	if err := json.NewDecoder(r.Body).Decode(&updatedMovie); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	if updatedMovie.Title != "" {
		movie.Title = updatedMovie.Title
	}
	if updatedMovie.Director != "" {
		movie.Director = updatedMovie.Director
	}
	if updatedMovie.Country != "" {
		movie.Country = updatedMovie.Country
	}
	if updatedMovie.ReleaseYear != 0 {
		movie.ReleaseYear = updatedMovie.ReleaseYear
	}
	if updatedMovie.Description != "" {
		movie.Description = updatedMovie.Description
	}
	if updatedMovie.Genres != nil {
		movie.Genres = updatedMovie.Genres
	}

	if err := db.Save(&movie).Error; err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(movie); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
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

func GetMoviesWithFilters(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	genres := r.URL.Query()["genres"]
	countries := r.URL.Query()["country"]
	yearFrom := r.URL.Query().Get("yearMin")
	yearTo := r.URL.Query().Get("yearMax")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Default values for pagination
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	// Initialize query
	query := db.Model(&models.Movie{})

	// Apply filters for genres
	if len(genres) > 0 {
		// Construct the JSON array for filtering and cast to jsonb (corrected formatting)
		genreFilter := fmt.Sprintf(`%s`, strings.Join(genres, `","`))
		query = query.Where("genres::jsonb @> ?::jsonb", genreFilter)
	}

	// Apply filters for countries
	if len(countries) > 0 {
		query = query.Where("country IN ?", countries)
	}

	// Apply year filters (year_from and year_to)
	if yearFrom != "" {
		if yearFromInt, err := strconv.Atoi(yearFrom); err == nil {
			query = query.Where("release_year >= ?", yearFromInt)
		}
	}
	if yearTo != "" {
		if yearToInt, err := strconv.Atoi(yearTo); err == nil {
			query = query.Where("release_year <= ?", yearToInt)
		}
	}

	// Sorting logic
	if sort == "" {
		sort = "title" // Default sort field
	}
	if order != "asc" && order != "desc" {
		order = "asc" // Default sort order
	}

	// Sort by genres (sorting by the first genre in the JSON array)
	if sort == "genres" {
		query = query.Order(fmt.Sprintf("genres->>0 %s", order)) // Extracts the first genre for sorting
	} else {
		query = query.Order(fmt.Sprintf("\"%s\" %s", sort, order)) // General sorting
	}

	// Count total records
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		http.Error(w, "Error counting movies", http.StatusInternalServerError)
		return
	}

	// Calculate total pages and apply pagination
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	offset := (page - 1) * limit

	// Fetch movies
	var movies []models.Movie
	if err := query.Offset(offset).Limit(limit).Find(&movies).Error; err != nil {
		http.Error(w, "Error fetching movies", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"movies":      movies,
		"total_pages": totalPages,
		"page":        page,
		"limit":       limit,
		"filters": map[string]interface{}{
			"genres":    genres,
			"countries": countries,
			"year_from": yearFrom,
			"year_to":   yearTo,
			"sort":      sort,
			"order":     order,
		},
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
