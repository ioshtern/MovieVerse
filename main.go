package main

import (
	"MovieVerse/controllers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"MovieVerse/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	err = db.AutoMigrate(&models.User{}, &models.Movie{}, &models.Review{})
	if err != nil {
		return
	}
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
		err := json.NewEncoder(w).Encode(models.Response{
			Status:  "fail",
			Message: "Invalid JSON message",
		})
		if err != nil {
			return
		}
		return
	}

	fmt.Println("Received message:", message)
	err := json.NewEncoder(w).Encode(models.Response{
		Status:  "success",
		Message: "Data successfully received",
	})
	if err != nil {
		return
	}
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(models.Response{
		Status:  "success",
		Message: "GET request received",
	})
	if err != nil {
		return
	}
}

func main() {
	http.HandleFunc("/", serveHTML)

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

	http.HandleFunc("/movies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetMovieByID(db, w, r)
			} else {
				controllers.GetMovies(db, w, r)
			}
		} else if r.Method == http.MethodPost {
			controllers.CreateMovie(db, w, r)
		} else if r.Method == http.MethodPut {
			controllers.UpdateMovie(db, w, r)
		} else if r.Method == http.MethodDelete {
			controllers.DeleteMovie(db, w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetUserByID(db, w, r)
			} else {
				controllers.GetUsers(db, w, r)
			}
		} else if r.Method == http.MethodPost {
			controllers.CreateUser(db, w, r)
		} else if r.Method == http.MethodPut {
			controllers.UpdateUser(db, w, r)
		} else if r.Method == http.MethodDelete {
			controllers.DeleteUser(db, w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetReviewByID(db, w, r)
			} else {
				controllers.GetReviews(db, w, r)
			}
		} else if r.Method == http.MethodPost {
			controllers.CreateReview(db, w, r)
		} else if r.Method == http.MethodPut {
			controllers.UpdateReview(db, w, r)
		} else if r.Method == http.MethodDelete {
			controllers.DeleteReview(db, w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/admin.html")
}
