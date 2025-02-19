package controllers

import (
	"MovieVerse/models"
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
)

var movieCollection *mongo.Collection

func InitMovieController(db *mongo.Database) {
	movieCollection = db.Collection("movies")
}

// Get all movies
func GetMovies(w http.ResponseWriter, r *http.Request) {
	cursor, err := movieCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var movies []models.Movie
	if err = cursor.All(context.TODO(), &movies); err != nil {
		http.Error(w, "Error decoding movies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// Get movie by ID
func GetMovieByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid movie ID format", http.StatusBadRequest)
		return
	}

	var movie models.Movie
	err = movieCollection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&movie)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

// Create new movie
func CreateMovie(w http.ResponseWriter, r *http.Request) {
	var movie models.Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	movie.ID = primitive.NewObjectID()
	_, err := movieCollection.InsertOne(context.TODO(), movie)
	if err != nil {
		http.Error(w, "Failed to create movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movie)
}

func UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid movie ID format", http.StatusBadRequest)
		return
	}

	var updatedData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	update := bson.M{"$set": updatedData}
	_, err = movieCollection.UpdateOne(context.TODO(), bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Movie updated successfully"})
}
func DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid movie ID format", http.StatusBadRequest)
		return
	}

	_, err = movieCollection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Movie deleted successfully"})
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

func GetMoviesWithFilters(w http.ResponseWriter, r *http.Request) {
	genres := r.URL.Query()["genres"]
	countries := r.URL.Query()["country"]
	yearFrom := r.URL.Query().Get("yearMin")
	yearTo := r.URL.Query().Get("yearMax")
	sortField := r.URL.Query().Get("sort")
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

	// Построение фильтра запроса
	filter := bson.M{}
	if len(genres) > 0 {
		filter["genres"] = bson.M{"$all": genres}
	}
	if len(countries) > 0 {
		filter["country"] = bson.M{"$in": countries}
	}
	if yearFrom != "" {
		if yearFromInt, err := strconv.Atoi(yearFrom); err == nil {
			filter["release_year"] = bson.M{"$gte": yearFromInt}
		}
	}
	if yearTo != "" {
		if yearToInt, err := strconv.Atoi(yearTo); err == nil {
			if val, exists := filter["release_year"]; exists {
				filter["release_year"] = bson.M{"$gte": val.(bson.M)["$gte"], "$lte": yearToInt}
			} else {
				filter["release_year"] = bson.M{"$lte": yearToInt}
			}
		}
	}

	// Определение сортировки
	sortOptions := bson.D{}
	if sortField == "" {
		sortField = "title"
	}
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}
	sortOptions = append(sortOptions, bson.E{Key: sortField, Value: sortOrder})

	// Подсчет количества записей
	totalRecords, err := movieCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error counting movies", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	skip := (page - 1) * limit

	// Запрос в MongoDB
	cursor, err := movieCollection.Find(context.TODO(), filter, options.Find().SetSort(sortOptions).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		http.Error(w, "Error fetching movies", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var movies []models.Movie
	if err = cursor.All(context.TODO(), &movies); err != nil {
		http.Error(w, "Error decoding movies", http.StatusInternalServerError)
		return
	}

	// Определение статуса запроса
	status := "none"
	if len(filter) > 0 && len(sortOptions) > 0 {
		status = "filtering and sorting"
	} else if len(filter) > 0 {
		status = "filtering"
	} else if len(sortOptions) > 0 {
		status = "sorting"
	}

	// Логирование запроса
	log.Printf("endpoint: /movies, method: GET, status: %s, filters: %+v, sort: %s, order: %s, page: %d, limit: %d",
		status, filter, sortField, order, page, limit)

	// Формирование JSON-ответа
	response := map[string]interface{}{
		"movies":      movies,
		"total_pages": totalPages,
		"page":        page,
		"limit":       limit,
		"filters":     filter,
		"status":      status,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
