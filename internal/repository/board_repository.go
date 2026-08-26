package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
		   OR ($1::int IS NOT NULL AND b.author_id = $1)
		   OR ($1::int IS NOT NULL AND EXISTS (
		       SELECT 1 FROM board_shares s
		       WHERE s.board_id = b.id AND s.user_id = $1
		   ))
		ORDER BY b.created_at DESC
	`

	var uid interface{}
	if userID != nil {
		uid = *userID

	}

	rows, err := r.db.Query(query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boards := make([]model.Board, 0)
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
	if err := r.db.QueryRow(query, id).Scan(
		&b.ID, &b.Title, &b.Content, &img, &b.IsPrivate,
		&b.AuthorID, &b.AuthorName, &b.CreatedAt,
	); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBoardNotFound
	} else if err != nil {
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
	return exists, err
}

func (r *BoardRepository) Delete(boardID, requestingUserID int) error {
	result, err := r.db.Exec(`DELETE FROM boards WHERE id = $1 AND author_id = $2`, boardID, requestingUserID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		var exists bool
		if err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM boards WHERE id = $1)`, boardID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrBoardNotFound
		}
		return ErrNotAuthorized
	}

	return nil
}

func (r *BoardRepository) AddShares(boardID int, usernames []string) error {
	for _, username := range normalizeUsernames(usernames) {
		var userID int
		if err := r.db.QueryRow(`SELECT id FROM users WHERE username = $1`, username).Scan(&userID); errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s", ErrUserNotFound, username)
		} else if err != nil {
			return err
		}
		if _, err := r.db.Exec(`INSERT INTO board_shares (board_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, boardID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *BoardRepository) ReplaceShares(boardID int, usernames []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM board_shares WHERE board_id = $1`, boardID); err != nil {
		return err
	}
	for _, username := range normalizeUsernames(usernames) {
		var userID int
		if err := tx.QueryRow(`SELECT id FROM users WHERE username = $1`, username).Scan(&userID); errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s", ErrUserNotFound, username)
		} else if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO board_shares (board_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, boardID, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *BoardRepository) SharedUsernames(boardID int) ([]string, error) {
	rows, err := r.db.Query(`
		SELECT u.username
		FROM board_shares s
		JOIN users u ON u.id = s.user_id
		WHERE s.board_id = $1
		ORDER BY u.username
	`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usernames := make([]string, 0)
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, rows.Err()
}

func normalizeUsernames(usernames []string) []string {
	seen := make(map[string]struct{}, len(usernames))
	result := make([]string, 0, len(usernames))
	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		if _, ok := seen[username]; ok {
			continue
		}
		seen[username] = struct{}{}
		result = append(result, username)
	}
	return result
}