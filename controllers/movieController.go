package controllers

import (
	"MovieVerse/models"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
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

func init() {
	// Set up logging to file
	logFile, err := os.OpenFile("user_actions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Include timestamp and file/line number
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

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	query := db.Model(&models.Movie{})

	isFiltering := false
	isSorting := sort != ""

	if len(genres) > 0 {
		isFiltering = true
		genreFilter := fmt.Sprintf(`%s`, strings.Join(genres, `","`))
		query = query.Where("genres::jsonb @> ?::jsonb", genreFilter)
	}

	if len(countries) > 0 {
		isFiltering = true
		query = query.Where("country IN ?", countries)
	}

	if yearFrom != "" {
		if yearFromInt, err := strconv.Atoi(yearFrom); err == nil {
			isFiltering = true
			query = query.Where("release_year >= ?", yearFromInt)
		}
	}
	if yearTo != "" {
		if yearToInt, err := strconv.Atoi(yearTo); err == nil {
			isFiltering = true
			query = query.Where("release_year <= ?", yearToInt)
		}
	}

	if sort == "" {
		sort = "title"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	if sort == "genres" {
		query = query.Order(fmt.Sprintf("genres->>0 %s", order))
	} else {
		query = query.Order(fmt.Sprintf("\"%s\" %s", sort, order))
	}

	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		http.Error(w, "Error counting movies", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	offset := (page - 1) * limit

	var movies []models.Movie
	if err := query.Offset(offset).Limit(limit).Find(&movies).Error; err != nil {
		http.Error(w, "Error fetching movies", http.StatusInternalServerError)
		return
	}

	status := "none"
	if isFiltering && isSorting {
		status = "filtering and sorting"
	} else if isFiltering {
		status = "filtering"
	} else if isSorting {
		status = "sorting"
	}

	log.Printf("endpoint: /movies, method: GET, status: %s, filters: %+v, sort: %s, order: %s, page: %d, limit: %d",
		status, map[string]interface{}{
			"genres":    genres,
			"countries": countries,
			"year_from": yearFrom,
			"year_to":   yearTo,
		}, sort, order, page, limit)

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
		"status": status,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
