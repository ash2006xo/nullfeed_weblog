package repository

import (
	"database/sql"
	"errors"

	"github.com/ash2006xo/nullfeed_weblog/internal/model"
)

var ErrBoardNotFound = errors.New("board not found")
var ErrNotAuthorized = errors.New("not authorized")

type BoardRepository struct {
	db *sql.DB
}

func NewBoardRepository(db *sql.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

func (r *BoardRepository) Create(title, content string, imageURL *string, isPrivate bool, authorID int, authorName string) (*model.Board, error) {
	query := `
		INSERT INTO boards (title, content, image_url, is_private, author_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, content, image_url, is_private, author_id, created_at
	`

	var b model.Board
	var img sql.NullString

	err := r.db.QueryRow(query, title, content, imageURL, isPrivate, authorID).Scan(
		&b.ID, &b.Title, &b.Content, &img, &b.IsPrivate, &b.AuthorID, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.AuthorName = authorName

	if img.Valid {
		b.ImageURL = &img.String
	}

	return &b, nil
}

func (r *BoardRepository) ListVisible(userID *int) ([]model.Board, error) {
	query := `
		SELECT b.id, b.title, b.content, b.image_url, b.is_private,
		       b.author_id, u.username, b.created_at
		FROM boards b
		JOIN users u ON u.id = b.author_id
		WHERE b.is_private = false
		   OR b.author_id = $1
		   OR EXISTS (
		       SELECT 1 FROM board_shares s
		       WHERE s.board_id = b.id AND s.user_id = $1
		   )
		ORDER BY b.created_at DESC
	`

	var uid int
	if userID != nil {
		uid = *userID
	} else {
		uid = -1
	}

	rows, err := r.db.Query(query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []model.Board
	for rows.Next() {
		var b model.Board
		var img sql.NullString

		if err := rows.Scan(&b.ID, &b.Title, &b.Content, &img, &b.IsPrivate,
			&b.AuthorID, &b.AuthorName, &b.CreatedAt); err != nil {
			return nil, err
		}
		if img.Valid {
			b.ImageURL = &img.String
		}
		boards = append(boards, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return boards, nil
}

func (r *BoardRepository) GetByID(id int) (*model.Board, error) {
	query := `
		SELECT b.id, b.title, b.content, b.image_url, b.is_private,
		       b.author_id, u.username, b.created_at
		FROM boards b
		JOIN users u ON u.id = b.author_id
		WHERE b.id = $1
	`

	var b model.Board
	var img sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&b.ID, &b.Title, &b.Content, &img, &b.IsPrivate,
		&b.AuthorID, &b.AuthorName, &b.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBoardNotFound
	}
	if err != nil {
		return nil, err
	}
	if img.Valid {
		b.ImageURL = &img.String
	}

	return &b, nil
}

func (r *BoardRepository) CanView(board *model.Board, userID *int) (bool, error) {
	if !board.IsPrivate {
		return true, nil
	}
	if userID == nil {
		return false, nil
	}
	if board.AuthorID == *userID {
		return true, nil
	}

	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM board_shares WHERE board_id = $1 AND user_id = $2)`,
		board.ID, *userID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *BoardRepository) Delete(boardID int, requestingUserID int) error {
	result, err := r.db.Exec(
		`DELETE FROM boards WHERE id = $1 AND author_id = $2`,
		boardID, requestingUserID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		var exists bool
		checkErr := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM boards WHERE id = $1)`, boardID).Scan(&exists)
		if checkErr != nil {
			return checkErr
		}
		if !exists {
			return ErrBoardNotFound
		}
		return ErrNotAuthorized
	}

	return nil
}

func (r *BoardRepository) AddShares(boardID int, usernames []string) error {
	for _, username := range usernames {
		_, err := r.db.Exec(`
			INSERT INTO board_shares (board_id, user_id)
			SELECT $1, u.id FROM users u WHERE u.username = $2
			ON CONFLICT DO NOTHING
		`, boardID, username)
		if err != nil {
			return err
		}
	}
	return nil
}