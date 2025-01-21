package main

import (
	"MovieVerse/controllers"
	"MovieVerse/models"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	db       *gorm.DB
	logger   = logrus.New()
	rlimiter *RateLimiter
)

type RateLimiter struct {
	limiter *rate.Limiter
	mutex   sync.Mutex
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.limiter.Allow()
}

func rateLimitedHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rlimiter.Allow() {
			// Set headers for the response
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusTooManyRequests) // Set status code first

			_, err := w.Write([]byte("<script>alert('Too Many Requests. Please try again later.');</script>"))
			if err != nil {
				logger.WithError(err).Error("Failed to write alert response for rate limiting")
			}
			return
		}
		next(w, r)
	}
}

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
}

func initDatabase() {
	var err error
	dsn := "user=postgres password=3052 dbname=movieverse port=5433 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to the database")
	}
	logger.Info("Database connected successfully")

	err = db.AutoMigrate(&models.User{}, &models.Movie{}, &models.Review{})
	if err != nil {
		logger.WithError(err).Fatal("Database migration failed")
	}
	logger.Info("Database migrated successfully")
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		logger.WithError(err).Error("Invalid JSON format in POST request")
		return
	}

	message, ok := input["message"].(string)
	if !ok || message == "" {
		http.Error(w, "Invalid JSON message", http.StatusBadRequest)
		logger.Error("Invalid or empty JSON message in POST request")
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

func main() {
	initLogger()
	initDatabase()

	rlimiter = NewRateLimiter(1, 1)

	http.HandleFunc("/", rateLimitedHandler(serveHTML))

	http.HandleFunc("/post", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlePostRequest(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /post endpoint")
		}
	}))

	http.HandleFunc("/get", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetRequest(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /get endpoint")
		}
	}))

	http.HandleFunc("/movies", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controllers.GetMoviesWithFilters(db, w, r)
			logAction(logrus.Fields{"endpoint": "/movies", "method": "GET"}, "Movies retrieved")
		case http.MethodPost:
			controllers.CreateMovie(db, w, r)
			logAction(logrus.Fields{"endpoint": "/movies", "method": "POST"}, "Movie created")
		case http.MethodPut:
			controllers.UpdateMovie(db, w, r)
			logAction(logrus.Fields{"endpoint": "/movies", "method": "PUT"}, "Movie updated")
		case http.MethodDelete:
			controllers.DeleteMovie(db, w, r)
			logAction(logrus.Fields{"endpoint": "/movies", "method": "DELETE"}, "Movie deleted")
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /movies endpoint")
		}
	}))

	http.HandleFunc("/users", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetUserByID(db, w, r)
				logAction(logrus.Fields{"endpoint": "/users", "method": "GET", "user_id": id}, "User details retrieved")
			} else {
				controllers.GetUsers(db, w, r)
				logAction(logrus.Fields{"endpoint": "/users", "method": "GET"}, "All users retrieved")
			}
		case http.MethodPost:
			controllers.CreateUser(db, w, r)
			logAction(logrus.Fields{"endpoint": "/users", "method": "POST"}, "User created")
		case http.MethodPut:
			controllers.UpdateUser(db, w, r)
			logAction(logrus.Fields{"endpoint": "/users", "method": "PUT"}, "User updated")
		case http.MethodDelete:
			controllers.DeleteUser(db, w, r)
			logAction(logrus.Fields{"endpoint": "/users", "method": "DELETE"}, "User deleted")
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /users endpoint")
		}
	}))

	http.HandleFunc("/reviews", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if id := r.URL.Query().Get("id"); id != "" {
				controllers.GetReviewByID(db, w, r)
				logAction(logrus.Fields{"endpoint": "/reviews", "method": "GET", "review_id": id}, "Review details retrieved")
			} else {
				controllers.GetReviews(db, w, r)
				logAction(logrus.Fields{"endpoint": "/reviews", "method": "GET"}, "All reviews retrieved")
			}
		case http.MethodPost:
			controllers.CreateReview(db, w, r)
			logAction(logrus.Fields{"endpoint": "/reviews", "method": "POST"}, "Review created")
		case http.MethodPut:
			controllers.UpdateReview(db, w, r)
			logAction(logrus.Fields{"endpoint": "/reviews", "method": "PUT"}, "Review updated")
		case http.MethodDelete:
			controllers.DeleteReview(db, w, r)
			logAction(logrus.Fields{"endpoint": "/reviews", "method": "DELETE"}, "Review deleted")
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /reviews endpoint")
		}
	}))

	logger.Info("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.WithError(err).Fatal("Could not start server")
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/admin.html")
	logger.Info("Admin HTML served")
}
