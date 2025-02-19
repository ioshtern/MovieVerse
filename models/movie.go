package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Movie struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Director    string             `json:"director" bson:"director"`
	Country     string             `json:"country" bson:"country"`
	Genres      []string           `json:"genres" bson:"genres"`
	ReleaseYear int                `json:"release_year" bson:"release_year"`
	Description string             `json:"description" bson:"description"`
	ImageLink   string             `json:"image_link" bson:"image_link"`
	Price       float64            `json:"price" bson:"price"`
}
