package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"MovieVerse/models"  
)

var db *gorm.DB

func initDatabase() {
	var err error
	dsn := "user=postgres password=3052 dbname=movieverse port=5433 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	fmt.Println("Database connected successfully")

	db.AutoMigrate(&models.User{}, &models.Movie{}, &models.Review{})
	fmt.Println("Database migrated")
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	message, ok := input["message"].(string)
	if !ok || message == "" {
		json.NewEncoder(w).Encode(models.Response{
			Status:  "fail",
			Message: "Invalid JSON message",
		})
		return
	}

	fmt.Println("Received message:", message)
	json.NewEncoder(w).Encode(models.Response{
		Status:  "success",
		Message: "Data successfully received",
	})
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(models.Response{
		Status:  "success",
		Message: "GET request received",
	})
}

func main() {
	initDatabase()

	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlePostRequest(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetRequest(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
