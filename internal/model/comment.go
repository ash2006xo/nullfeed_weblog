package model

import "time"

type Comment struct {
	ID         int       `json:"id"`
	BoardID    int       `json:"board_id"`
	AuthorID   int       `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}
