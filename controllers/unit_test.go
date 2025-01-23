package controllers

import (
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

func newTestDB(t *testing.T) *gorm.DB {
	dsn := "user=postgres password=Bruhmomento dbname=movieverse port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	db.AutoMigrate(&models.User{})

	db.Exec("DELETE FROM users")
	return db
}

func TestLoginUser_Success(t *testing.T) {
	db := newTestDB(t)

	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{Email: "test@example.com", Password: string(password)}
	db.Create(&testUser)

	var createdUser models.User
	db.Where("email = ?", testUser.Email).First(&createdUser)

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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LoginUser(db, w, r)
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"message":"Login successful","userId":"` + fmt.Sprintf("%d", createdUser.ID) + `"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Handler returned unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	db := newTestDB(t)

	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{Email: "test@example.com", Password: string(password)}
	db.Create(&testUser)

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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LoginUser(db, w, r)
	})
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusUnauthorized)
	}

	expected := `{"error":"Invalid email or password"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Handler returned unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}
