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
	"time"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	movieCollection := client.Database("movieverse").Collection("movies")
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
func GetMovieByID(w http.ResponseWriter, r *http.Request) {
	movieCollection := client.Database("movieverse").Collection("movies")
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
func CreateMovie(w http.ResponseWriter, r *http.Request) {
	movieCollection := client.Database("movieverse").Collection("movies")

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
	movieCollection := client.Database("movieverse").Collection("movies")

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
	movieCollection := client.Database("movieverse").Collection("movies")

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
	logFile, err := os.OpenFile("user_actions.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func GetMoviesWithFilters(w http.ResponseWriter, r *http.Request) {
	movieCollection := client.Database("movieverse").Collection("movies")
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

	sortOptions := bson.D{}
	if sortField == "" {
		sortField = "title"
	}
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}
	sortOptions = append(sortOptions, bson.E{Key: sortField, Value: sortOrder})

	totalRecords, err := movieCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error counting movies", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	skip := (page - 1) * limit

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

	status := "none"
	if len(filter) > 0 && len(sortOptions) > 0 {
		status = "filtering and sorting"
	} else if len(filter) > 0 {
		status = "filtering"
	} else if len(sortOptions) > 0 {
		status = "sorting"
	}

	log.Printf("endpoint: /movies, method: GET, status: %s, filters: %+v, sort: %s, order: %s, page: %d, limit: %d",
		status, filter, sortField, order, page, limit)

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
func SearchAndFilterMovies(w http.ResponseWriter, r *http.Request) {
	movieCollection := client.Database("movieverse").Collection("movies")
	query := r.URL.Query()

	filter := bson.M{}

	searchTerm := query.Get("q")
	if searchTerm != "" {
		filter["title"] = bson.M{"$regex": searchTerm, "$options": "i"}
	}

	category := query.Get("category")
	if category != "" {
		filter["genres"] = category
	}

	minPriceStr := query.Get("minPrice")
	maxPriceStr := query.Get("maxPrice")
	if minPriceStr != "" || maxPriceStr != "" {
		priceFilter := bson.M{}
		if minPriceStr != "" {
			minPrice, err := strconv.ParseFloat(minPriceStr, 64)
			if err == nil {
				priceFilter["$gte"] = minPrice
			}
		}
		if maxPriceStr != "" {
			maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
			if err == nil {
				priceFilter["$lte"] = maxPrice
			}
		}
		filter["price"] = priceFilter
	}

	availability := query.Get("availability")
	if availability != "" {
		avail, err := strconv.ParseBool(availability)
		if err == nil {
			filter["inStock"] = avail
		}
	}

	sortField := query.Get("sort")
	if sortField == "" {
		sortField = "title"
	}
	order := query.Get("order")
	sortOrder := 1
	if order == "desc" {
		sortOrder = -1
	}
	sortOptions := bson.D{{Key: sortField, Value: sortOrder}}

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	opts := options.Find().SetSort(sortOptions).SetSkip(int64(skip)).SetLimit(int64(limit))
	cursor, err := movieCollection.Find(context.TODO(), filter, opts)
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

	totalCount, err := movieCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		totalCount = int64(len(movies))
	}

	response := map[string]interface{}{
		"movies": movies,
		"page":   page,
		"limit":  limit,
		"total":  totalCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type MovieItem struct {
	ID       string  `json:"id" bson:"id"`
	Title    string  `json:"title" bson:"title"`
	Price    float64 `json:"price" bson:"price"`
	Image    string  `json:"image" bson:"image"`
	Quantity int     `json:"quantity" bson:"quantity"`
}

type CheckoutRequest struct {
	Movies []MovieItem `json:"movies" bson:"movies"`
}

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Movies      []MovieItem        `bson:"movies" json:"movies"`
	Total       float64            `bson:"total" json:"total"`
	OrderStatus string             `bson:"order_status" json:"order_status"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

func Checkout(w http.ResponseWriter, r *http.Request) {
	log.Println("Checkout handler invoked")

	if r.Method != http.MethodPost {
		log.Println("Invalid request method:", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req CheckoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error decoding JSON payload:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	log.Printf("Decoded checkout payload: %+v\n", req)

	if len(req.Movies) == 0 {
		log.Println("Cart is empty")
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	dummyUserID := primitive.NewObjectID()
	log.Println("Using dummy user ID:", dummyUserID.Hex())

	var total float64
	for i, item := range req.Movies {
		lineTotal := item.Price * float64(item.Quantity)
		log.Printf("Movie %d: Price=%.2f, Quantity=%d, LineTotal=%.2f", i, item.Price, item.Quantity, lineTotal)
		total += lineTotal
	}
	log.Printf("Total order cost: %.2f", total)

	order := Order{
		UserID:      dummyUserID,
		Movies:      req.Movies,
		Total:       total,
		OrderStatus: "pending",
		CreatedAt:   time.Now(),
	}
	log.Printf("Order to insert: %+v", order)

	orderCollection := client.Database("movieverse").Collection("orders")
	result, err := orderCollection.InsertOne(context.TODO(), order)
	if err != nil {
		log.Println("Error inserting order into MongoDB:", err)
		http.Error(w, "Failed to process checkout", http.StatusInternalServerError)
		return
	}
	log.Println("Order inserted successfully. InsertedID:", result.InsertedID)

	w.Header().Set("Content-Type", "application/json")
	response := models.Response{
		Status:  "success",
		Message: "Checkout successful!",
	}
	log.Printf("Sending success response: %+v", response)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding success response:", err)
	}
}

type ActivityLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Action    string             `bson:"action" json:"action"`
	Detail    string             `bson:"detail" json:"detail"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}

func LogUserActivity(userID primitive.ObjectID, action, detail string) {
	logEntry := ActivityLog{
		UserID:    userID,
		Action:    action,
		Detail:    detail,
		Timestamp: time.Now(),
	}
	activityCollection := client.Database("movieverse").Collection("activity_logs")
	_, err := activityCollection.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Println("Error logging user activity:", err)
	}
}
func GetAnalyticsDashboard(w http.ResponseWriter, r *http.Request) {
	ordersCollection := client.Database("movieverse").Collection("orders")

	salesPipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalSales", Value: bson.D{{Key: "$sum", Value: "$total"}}},
			{Key: "orderCount", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursor, err := ordersCollection.Aggregate(context.TODO(), salesPipeline)
	if err != nil {
		http.Error(w, "Error fetching sales data", http.StatusInternalServerError)
		return
	}
	var salesResults []bson.M
	if err = cursor.All(context.TODO(), &salesResults); err != nil {
		http.Error(w, "Error decoding sales data", http.StatusInternalServerError)
		return
	}
	totalSales := 0.0
	orderCount := 0
	if len(salesResults) > 0 {
		totalSales = salesResults[0]["totalSales"].(float64)
		switch v := salesResults[0]["orderCount"].(type) {
		case int32:
			orderCount = int(v)
		case int64:
			orderCount = int(v)
		}
	}

	purchasesPipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$movies"}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "id", Value: "$movies.id"},
				{Key: "title", Value: "$movies.title"},
			}},
			{Key: "totalQuantity", Value: bson.D{{Key: "$sum", Value: "$movies.quantity"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "totalQuantity", Value: -1}}}},
		{{Key: "$limit", Value: 5}},
	}
	cursorPurchases, err := ordersCollection.Aggregate(context.TODO(), purchasesPipeline)
	if err != nil {
		http.Error(w, "Error fetching purchase data", http.StatusInternalServerError)
		return
	}
	var purchaseResults []bson.M
	if err = cursorPurchases.All(context.TODO(), &purchaseResults); err != nil {
		http.Error(w, "Error decoding purchase data", http.StatusInternalServerError)
		return
	}

	dashboard := map[string]interface{}{
		"totalSales":          totalSales,
		"orderCount":          orderCount,
		"mostPurchasedMovies": purchaseResults,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}
