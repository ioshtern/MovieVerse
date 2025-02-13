package models

import "time"

type ChatSession struct {
	ID        uint      `gorm:"primaryKey"`
	ClientID  uint      `gorm:"not null"`
	Status    string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	ClosedAt  *time.Time
	Messages  []ChatMessage `gorm:"foreignKey:ChatSessionID"`
}

type ChatMessage struct {
	ID            uint      `gorm:"primaryKey"`
	ChatSessionID uint      `gorm:"not null"`
	Sender        string    `gorm:"not null"`
	Content       string    `gorm:"not null"`
	Timestamp     time.Time `gorm:"not null"`
}
