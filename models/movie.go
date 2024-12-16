package models


type Movie struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title"`
	Director string `json:"director"`
	Country string `json:"country"`
	Genres 	[]string `json:"genres"`
	ReleaseYear int    `json:"release_year"`
	Description string `json:"description"`

}