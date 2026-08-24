package repository

import (
	"database/sql"

	"github.com/ash2006xo/nullfeed_weblog/internal/model"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(boardID, authorID int, content, authorName string) (*model.Comment, error) {
	query := `
		INSERT INTO comments (board_id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, board_id, author_id, content, created_at
	`

	var cm model.Comment
	err := r.db.QueryRow(query, boardID, authorID, content).Scan(
		&cm.ID, &cm.BoardID, &cm.AuthorID, &cm.Content, &cm.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	cm.AuthorName = authorName

	return &cm, nil
}

func (r *CommentRepository) ListByBoard(boardID int) ([]model.Comment, error) {
	query := `
		SELECT c.id, c.board_id, c.author_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON u.id = c.author_id
		WHERE c.board_id = $1
		ORDER BY c.created_at ASC
	`

	rows, err := r.db.Query(query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var cm model.Comment
		if err := rows.Scan(&cm.ID, &cm.BoardID, &cm.AuthorID, &cm.AuthorName, &cm.Content, &cm.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, cm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}