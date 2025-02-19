package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Email             string             `bson:"email"`
	Username          string             `bson:"username"`
	Password          string             `bson:"password" `
	Admin             bool               `bson:"admin" `
	VerificationToken string             `bson:"verification_token"`
	EmailVerified     bool               `bson:"email_verified" gorm:"default:false"`
}
