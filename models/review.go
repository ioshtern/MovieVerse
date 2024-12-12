package models


type Review struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
	MovieID uint   `json:"movie_id"`
}
