package model

import "time"

type Board struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	ImageURL   *string   `json:"image_url"`
	IsPrivate  bool      `json:"is_private"`
	AuthorID   int       `json:"author_id"`
	AuthorName string    `json:"author_name"`
	CreatedAt  time.Time `json:"created_at"`
}
