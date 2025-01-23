package models

type User struct {
	ID                uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Email             string `json:"email" gorm:"uniqueIndex;not null"`
	Username          string `json:"username" gorm:"not null"`
	Password          string `json:"password" gorm:"not null"`
	Admin             bool   `json:"admin" gorm:"default:false"`
	VerificationToken string `json:"verification_token"`
	EmailVerified     bool   `json:"email_verified" gorm:"default:false"`
}
