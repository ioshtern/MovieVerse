package main

import (
	"MovieVerse/controllers"
	"MovieVerse/models"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

var (
	db     *gorm.DB
	logger = logrus.New()
)

func initLogger() {
	file, err := os.OpenFile("user_actions.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	logger.SetOutput(file)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel) 
}

func logAction(fields logrus.Fields, message string) {
	logger.WithFields(fields).Info(message)
	fmt.Printf("%s | Fields: %+v\n", message, fields)
}

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
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		return
	}

	logAction(logrus.Fields{
		"message": message,
		"action":  "post_request",
	}, "POST request received")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.Response{
		Status:  "success",
		Message: "Data successfully received",
	})
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	logAction(logrus.Fields{
		"action": "get_request",
	}, "GET request received")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(models.Response{
		Status:  "success",
		Message: "GET request received",
	})
}

func handleMoviesEndpoint(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			query := db.Model(&models.Movie{})

			// Filtering
			filter := r.URL.Query().Get("filter")
			if filter != "" {
				query = query.Where("title ILIKE ? OR director ILIKE ?", "%"+filter+"%", "%"+filter+"%")
				logAction(logrus.Fields{
					"action": "filter",
					"filter": filter,
				}, "Movies filtered")
			}

			// Sorting
			sort := r.URL.Query().Get("sort")
			order := r.URL.Query().Get("order")
			if sort != "" {
				if order != "asc" && order != "desc" {
					order = "asc"
				}
				if sort == "genres" {
					query = query.Order("genres " + order)
				} else {
					query = query.Order(fmt.Sprintf("%s %s", sort, order))
				}
				logAction(logrus.Fields{
					"action": "sort",
					"sort":   sort,
					"order":  order,
				}, "Movies sorted")
			}

			// Pagination
			page := 1
			limit := 10
			if p := r.URL.Query().Get("page"); p != "" {
				fmt.Sscanf(p, "%d", &page)
			}
			if l := r.URL.Query().Get("limit"); l != "" {
				fmt.Sscanf(l, "%d", &limit)
			}

			offset := (page - 1) * limit
			query = query.Offset(offset).Limit(limit)

			var movies []models.Movie
			if err := query.Find(&movies).Error; err != nil {
				http.Error(w, "Failed to retrieve movies", http.StatusInternalServerError)
				return
			}

			// Total count for pagination
			var total int64
			db.Model(&models.Movie{}).Count(&total)
			totalPages := (total + int64(limit) - 1) / int64(limit)

			response := map[string]interface{}{
				"movies":      movies,
				"page":        page,
				"total_pages": totalPages,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			logAction(logrus.Fields{
				"action": "get_movies",
				"page":   page,
				"limit":  limit,
			}, "Movies retrieved")
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	initLogger()
	initDatabase()

	http.HandleFunc("/", serveHTML)

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
			query := db.Model(&models.Movie{})

			// Filtering
			filter := r.URL.Query().Get("filter")
			if filter != "" {
				query = query.Where("title ILIKE ? OR director ILIKE ?", "%"+filter+"%", "%"+filter+"%")
				logAction(logrus.Fields{
					"action": "filter",
					"filter": filter,
				}, "Movies filtered")
			}

			// Sorting
			sort := r.URL.Query().Get("sort")
			order := r.URL.Query().Get("order")
			if sort != "" {
				if order != "asc" && order != "desc" {
					order = "asc"
				}
				if sort == "genres" {
					query = query.Order("genres " + order)
				} else {
					query = query.Order(fmt.Sprintf("%s %s", sort, order))
				}
				logAction(logrus.Fields{
					"action": "sort",
					"sort":   sort,
					"order":  order,
				}, "Movies sorted")
			}

			// Pagination
			page := 1
			limit := 10
			if p := r.URL.Query().Get("page"); p != "" {
				fmt.Sscanf(p, "%d", &page)
			}
			if l := r.URL.Query().Get("limit"); l != "" {
				fmt.Sscanf(l, "%d", &limit)
			}

			offset := (page - 1) * limit
			query = query.Offset(offset).Limit(limit)

			var movies []models.Movie
			if err := query.Find(&movies).Error; err != nil {
				http.Error(w, "Failed to retrieve movies", http.StatusInternalServerError)
				return
			}

			// Total count for pagination
			var total int64
			db.Model(&models.Movie{}).Count(&total)
			totalPages := (total + int64(limit) - 1) / int64(limit)

			response := map[string]interface{}{
				"movies":      movies,
				"page":        page,
				"total_pages": totalPages,
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
			logAction(logrus.Fields{
				"action": "get_movies",
				"page":   page,
				"limit":  limit,
			}, "Movies retrieved")
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetUserByID(db, w, r)
				logAction(logrus.Fields{
					"endpoint": "/users",
					"method":   "GET",
					"user_id":  id,
				}, "User details retrieved")
			} else {
				controllers.GetUsers(db, w, r)
				logAction(logrus.Fields{
					"endpoint": "/users",
					"method":   "GET",
				}, "All users retrieved")
			}
		} else if r.Method == http.MethodPost {
			controllers.CreateUser(db, w, r)
			logAction(logrus.Fields{
				"endpoint": "/users",
				"method":   "POST",
			}, "User created")
		} else if r.Method == http.MethodPut {
			controllers.UpdateUser(db, w, r)
			logAction(logrus.Fields{
				"endpoint": "/users",
				"method":   "PUT",
			}, "User updated")
		} else if r.Method == http.MethodDelete {
			controllers.DeleteUser(db, w, r)
			logAction(logrus.Fields{
				"endpoint": "/users",
				"method":   "DELETE",
			}, "User deleted")
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/reviews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetReviewByID(db, w, r)
				logAction(logrus.Fields{
					"endpoint":  "/reviews",
					"method":    "GET",
					"review_id": id,
				}, "Review details retrieved")
			} else {
				controllers.GetReviews(db, w, r)
				logAction(logrus.Fields{
					"endpoint": "/reviews",
					"method":   "GET",
				}, "All reviews retrieved")
			}
		} else if r.Method == http.MethodPost {
			controllers.CreateReview(db, w, r)
			logAction(logrus.Fields{
				"endpoint": "/reviews",
				"method":   "POST",
			}, "Review created")
		} else if r.Method == http.MethodPut {
			controllers.UpdateReview(db, w, r)
			logAction(logrus.Fields{
				"endpoint": "/reviews",
				"method":   "PUT",
			}, "Review updated")
		} else if r.Method == http.MethodDelete {
			controllers.DeleteReview(db, w, r)
			logAction(logrus.Fields{
				"endpoint": "/reviews",
				"method":   "DELETE",
			}, "Review deleted")
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
