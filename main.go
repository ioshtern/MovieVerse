package main

import (
	"MovieVerse/controllers"
	"MovieVerse/models"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	db        *gorm.DB
	logger    = logrus.New()
	rlimiter  *RateLimiter
	broadcast = make(chan ChatWSMessage)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	mutex = &sync.Mutex{}
)

type Message struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type ChatWSMessage struct {
	ChatID    string `json:"chat_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

var activeChats = make(map[string]map[*websocket.Conn]bool)

// handleConnections upgrades the HTTP connection to a WebSocket connection.
// It verifies that the provided chat session exists (or creates one if not).
func handleConnections(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "Missing chat session ID", http.StatusBadRequest)
		return
	}

	// Verify that the chat session exists in the DB.
	sessionID, err := strconv.ParseUint(chatID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat session ID", http.StatusBadRequest)
		return
	}
	var session models.ChatSession
	if err := db.First(&session, uint(sessionID)).Error; err != nil {
		// Chat session not found; create one (using dummy clientID 1, replace as needed).
		newSession, err2 := getOrCreateChatSession(1)
		if err2 != nil {
			http.Error(w, "Failed to create chat session", http.StatusInternalServerError)
			return
		}
		chatID = strconv.Itoa(int(newSession.ID))
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer ws.Close()

	mutex.Lock()
	if activeChats[chatID] == nil {
		activeChats[chatID] = make(map[*websocket.Conn]bool)
	}
	activeChats[chatID][ws] = true
	mutex.Unlock()

	for {
		var msg ChatWSMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			mutex.Lock()
			delete(activeChats[chatID], ws)
			mutex.Unlock()
			break
		}
		msg.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		saveChatMessage(chatID, msg)
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		if conns, ok := activeChats[msg.ChatID]; ok {
			for client := range conns {
				err := client.WriteJSON(msg)
				if err != nil {
					client.Close()
					delete(conns, client)
				}
			}
		}
		mutex.Unlock()
	}
}

func saveChatMessage(chatID string, msg ChatWSMessage) {
	sessionID, err := strconv.ParseUint(chatID, 10, 64)
	if err != nil {
		log.Println("Invalid chatID:", err)
		return
	}
	chatMsg := models.ChatMessage{
		ChatSessionID: uint(sessionID),
		Sender:        msg.Username,
		Content:       msg.Content,
		Timestamp:     time.Now(),
	}
	if err := db.Create(&chatMsg).Error; err != nil {
		log.Println("Failed to save chat message:", err)
	}
}

func getOrCreateChatSession(clientID uint) (*models.ChatSession, error) {
	var session models.ChatSession
	err := db.Where("client_id = ? AND status = ?", clientID, "active").First(&session).Error
	if err == nil {
		return &session, nil
	}
	session = models.ChatSession{
		ClientID:  clientID,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	if err := db.Create(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// Dummy extraction of clientID; replace with your JWT/session extraction logic.
func extractClientID(r *http.Request) uint {
	return 1
}

// startChatHandler is the /start-chat endpoint that creates or retrieves a chat session.
func startChatHandler(w http.ResponseWriter, r *http.Request) {
	clientID := extractClientID(r)
	session, err := getOrCreateChatSession(clientID)
	if err != nil {
		http.Error(w, "Failed to create chat session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// closeChatHandler marks a chat session as closed and removes its active connections.
func closeChatHandler(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "Missing chat_id", http.StatusBadRequest)
		return
	}
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}
	now := time.Now()
	err = db.Model(&models.ChatSession{}).
		Where("id = ?", uint(chatID)).
		Updates(models.ChatSession{
			Status:   "closed",
			ClosedAt: &now,
		}).Error
	if err != nil {
		http.Error(w, "Failed to close chat", http.StatusInternalServerError)
		return
	}
	mutex.Lock()
	if conns, ok := activeChats[chatIDStr]; ok {
		for conn := range conns {
			conn.Close()
			delete(conns, conn)
		}
		delete(activeChats, chatIDStr)
	}
	mutex.Unlock()
	w.Write([]byte("Chat closed successfully"))
}

// chatHistoryHandler returns all chat messages for a given chat session.
func chatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	chatIDStr := r.URL.Query().Get("chat_id")
	if chatIDStr == "" {
		http.Error(w, "Missing chat_id", http.StatusBadRequest)
		return
	}
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chat_id", http.StatusBadRequest)
		return
	}
	var messages []models.ChatMessage
	if err := db.Where("chat_session_id = ?", uint(chatID)).
		Order("timestamp asc").
		Find(&messages).Error; err != nil {
		http.Error(w, "Failed to load chat history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// ActiveChat represents data returned for each active chat session to the admin.
type ActiveChat struct {
	ChatID    string `json:"chat_id"`
	Client    string `json:"client"`
	StartedAt string `json:"started_at"`
	Clients   int    `json:"clients"`
}

// activeChatsHandler returns a list of active chat sessions.
func activeChatsHandler(w http.ResponseWriter, r *http.Request) {
	var chats []ActiveChat
	mutex.Lock()
	for id, conns := range activeChats {
		if len(conns) > 0 {
			sessionID, err := strconv.ParseUint(id, 10, 64)
			clientStr := "Unknown"
			startedAt := "Unknown"
			if err == nil {
				var session models.ChatSession
				if err := db.First(&session, uint(sessionID)).Error; err == nil {
					clientStr = strconv.Itoa(int(session.ClientID))
					startedAt = session.CreatedAt.Format("2006-01-02 15:04:05")
				}
			}
			chats = append(chats, ActiveChat{
				ChatID:    id,
				Client:    clientStr,
				StartedAt: startedAt,
				Clients:   len(conns),
			})
		}
	}
	mutex.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

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
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusTooManyRequests)
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
	err = db.AutoMigrate(&models.User{}, &models.Movie{}, &models.Review{}, &models.ChatSession{}, &models.ChatMessage{})
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
	logAction(logrus.Fields{"action": "get_request"}, "GET request received")
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

	http.Handle("/", controllers.ValidateJWT(controllers.UsersOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		absPath, err := filepath.Abs("static/index.html")
		if err != nil {
			log.Fatal(err)
		}
		http.ServeFile(w, r, absPath)
	}))))

	http.Handle("/index.html", controllers.ValidateJWT(controllers.UsersOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		absPath, err := filepath.Abs("static/index.html")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Handling request for: %s", r.URL.Path)
		http.ServeFile(w, r, absPath)
	}))))

	http.Handle("/admin.html", controllers.ValidateJWT(controllers.AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/admin.html")
	}))))

	// Register endpoints.
	http.Handle("/start-chat", controllers.ValidateJWT(controllers.UsersOnly(http.HandlerFunc(startChatHandler))))
	http.Handle("/chat-history", controllers.ValidateJWT(controllers.UsersOnly(http.HandlerFunc(chatHistoryHandler))))
	http.Handle("/admin/active-chats", controllers.ValidateJWT(controllers.AdminOnly(http.HandlerFunc(activeChatsHandler))))
	http.Handle("/close-chat", controllers.ValidateJWT(controllers.AdminOnly(http.HandlerFunc(closeChatHandler))))

	http.HandleFunc("/post", rateLimitedHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlePostRequest(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			logger.Warn("Invalid request method for /post endpoint")
		}
	}))

	http.HandleFunc("/signup.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/signup.html")
	})
	http.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")
	})

	http.Handle("/login", rateLimitedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.LoginUser(db, w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	http.Handle("/signup", rateLimitedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			controllers.CreateUser(db, w, r)
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/logout", controllers.ValidateJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			http.SetCookie(w, &http.Cookie{
				Name:     "userToken",
				Value:    "",
				Path:     "/",
				Expires:  time.Now().Add(-1 * time.Hour),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/login.html", http.StatusSeeOther)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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

	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	log.Println("WebSocket server started on ws://localhost:8080/ws")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
