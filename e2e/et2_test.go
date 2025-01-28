package main

import (
	"MovieVerse/controllers"
	"MovieVerse/models"
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testDB *gorm.DB
var testMux *http.ServeMux

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "user=postgres password=3052 dbname=movieverse port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&models.User{})

	db.Exec("DELETE FROM users1")
	return db
}

func setupTestServer() {
	testMux = http.NewServeMux()
	testMux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		controllers.LoginUser(testDB, w, r)
	})
}

func TestLoginUser_EndToEnd(t *testing.T) {
	testDB = setupTestDB(t)
	setupTestServer()

	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{Email: "test@example.com", Password: string(password)}
	testDB.Create(&testUser)

	payload := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	testMux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	var createdUser models.User
	testDB.Where("email = ?", testUser.Email).First(&createdUser)

	expected := `{"message":"Login successful","userId":"` + fmt.Sprintf("%d", createdUser.ID) + `"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Handler returned unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	testDB = setupTestDB(t)
	setupTestServer()

	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{Email: "test@example.com", Password: string(password)}
	testDB.Create(&testUser)

	payload := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	testMux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusUnauthorized)
	}

	expected := `{"error":"Invalid email or password"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Handler returned unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}
