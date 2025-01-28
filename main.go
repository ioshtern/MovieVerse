package main

import (
	"MovieVerse/controllers"
	"MovieVerse/models"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := "user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" port=" + os.Getenv("DB_PORT") +
		" sslmode=" + os.Getenv("DB_SSLMODE")

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	log.Println("Database connected successfully")

	err = db.AutoMigrate(&models.User{}, &models.Movie{}, &models.Review{})
	if err != nil {
		log.Fatal("Database migration failed:", err)
	}
	log.Println("Database migrated successfully")
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
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func main() {
	initLogger()
	initDatabase()

	rlimiter = NewRateLimiter(1, 1)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodGet {
			serveHTML(w, r)
			return
		}
		respondWithError(w, http.StatusNotFound, "Endpoint not found")
	})

	http.HandleFunc("/post", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlePostRequest(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /post endpoint")
		}
	}))
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("userToken")
		if err != nil || cookie.Value == "" {
			http.ServeFile(w, r, "static/index.html")
			return
		}

		page, err := os.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "Unable to load the main page", http.StatusInternalServerError)
			return
		}

		updatedPage := strings.ReplaceAll(string(page), `<a href="login.html" class="btn btn-outline-light ms-2">Login</a>`, "")
		updatedPage = strings.ReplaceAll(updatedPage, `<a href="signup.html" class="btn btn-outline-light ms-2">Sign Up</a>`, "")

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(updatedPage))
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/signup.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/signup.html")
	})
	http.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")

	})
	http.HandleFunc("/admin.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/admin.html")

	})
	http.Handle("/signup", rateLimitedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.CreateUser(db, w, r)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/login", rateLimitedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.LoginUser(db, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	http.Handle("/index", controllers.ValidateJWT(http.HandlerFunc(controllers.ProtectedHandler)))

	http.HandleFunc("/admin.html", func(w http.ResponseWriter, r *http.Request) {
		controllers.ValidateJWT(controllers.AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/admin.html")
		})))(w, r)
	})

	http.Handle("/logout", controllers.ValidateJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Clear the user's JWT token cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "userToken",
				Value:    "",
				Path:     "/",
				Expires:  time.Now().Add(-1 * time.Hour),
				HttpOnly: true,
				Secure:   true, // Set to true if running on HTTPS
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/index.html", http.StatusSeeOther)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.HandleFunc("/verify-email", func(w http.ResponseWriter, r *http.Request) {
		controllers.VerifyEmail(db, w, r)
	})

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
	http.ServeFile(w, r, "static/index.html")
	logger.Info("Admin HTML served")
}
