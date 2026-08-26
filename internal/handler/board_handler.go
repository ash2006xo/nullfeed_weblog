package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	custommw "github.com/ash2006xo/nullfeed_weblog/internal/middleware"
	"github.com/ash2006xo/nullfeed_weblog/internal/model"
	"github.com/ash2006xo/nullfeed_weblog/internal/repository"
)

type BoardHandler struct {
	boardRepo *repository.BoardRepository
}

func NewBoardHandler(boardRepo *repository.BoardRepository) *BoardHandler {
	return &BoardHandler{boardRepo: boardRepo}
}

type createBoardRequest struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	ImageURL  *string  `json:"image_url"`
	IsPrivate bool     `json:"is_private"`
	ShareWith []string `json:"share_with"`
}

type sharesRequest struct {
	Usernames []string `json:"usernames"`
}

func (h *BoardHandler) Create(c echo.Context) error {
	userID, ok := custommw.CurrentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}
	username, _ := custommw.CurrentUsername(c)

	var req createBoardRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	if req.Title == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title and content are required"})
	}
	if len(req.Title) > 200 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title must be 200 characters or fewer"})
	}
	if len(req.Content) > 50000 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content must be 50,000 characters or fewer"})
	}
	if !req.IsPrivate {
		req.ShareWith = nil
	}
	if req.ImageURL != nil {
		value := strings.TrimSpace(*req.ImageURL)
		if value == "" {
			req.ImageURL = nil
		} else if !isSafeImageURL(value) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "image URL must use http, https, or a nullfeed upload"})
		} else {
			req.ImageURL = &value
		}
	}

	board, err := h.boardRepo.Create(req.Title, req.Content, req.ImageURL, req.IsPrivate, userID, username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create board"})
	}

	if req.IsPrivate && len(req.ShareWith) > 0 {
		if err := h.boardRepo.AddShares(board.ID, req.ShareWith); err != nil {
			_ = h.boardRepo.Delete(board.ID, userID)
			if errors.Is(err, repository.ErrUserNotFound) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to share board"})
		}
	}

	return c.JSON(http.StatusCreated, board)
}

func (h *BoardHandler) List(c echo.Context) error {
	var userIDPtr *int
	if userID, ok := custommw.CurrentUserID(c); ok {
		userIDPtr = &userID
	}

	boards, err := h.boardRepo.ListVisible(userIDPtr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load boards"})
	}
	return c.JSON(http.StatusOK, boards)
}

func (h *BoardHandler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid board id"})
	}

	board, err := h.boardRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrBoardNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "board not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load board"})
	}

	var userIDPtr *int
	if userID, ok := custommw.CurrentUserID(c); ok {
		userIDPtr = &userID
	}
	canView, err := h.boardRepo.CanView(board, userIDPtr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to check permissions"})
	}
	if !canView {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "board not found"})
	}
	return c.JSON(http.StatusOK, board)
}

func (h *BoardHandler) Delete(c echo.Context) error {
	userID, ok := custommw.CurrentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid board id"})
	}

	err = h.boardRepo.Delete(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrBoardNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "board not found"})
		}
		if errors.Is(err, repository.ErrNotAuthorized) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "you can only delete your own boards"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete board"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *BoardHandler) ListShares(c echo.Context) error {
	board, err := h.getOwnedBoard(c)
	if err != nil {
		return err
	}
	usernames, err := h.boardRepo.SharedUsernames(board.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load shares"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"usernames": usernames})
}

func (h *BoardHandler) ReplaceShares(c echo.Context) error {
	board, err := h.getOwnedBoard(c)
	if err != nil {
		return err
	}
	if !board.IsPrivate {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "only private boards can be shared"})
	}

	var req sharesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if len(req.Usernames) > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "too many shared users"})
	}

	if err := h.boardRepo.ReplaceShares(board.ID, req.Usernames); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update shares"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"usernames": req.Usernames})
}

func (h *BoardHandler) getOwnedBoard(c echo.Context) (*model.Board, error) {
	userID, ok := custommw.CurrentUserID(c)
	if !ok {
		return nil, c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return nil, c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid board id"})
	}
	board, err := h.boardRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrBoardNotFound) {
			return nil, c.JSON(http.StatusNotFound, map[string]string{"error": "board not found"})
		}
		return nil, c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load board"})
	}
	if board.AuthorID != userID {
		return nil, c.JSON(http.StatusForbidden, map[string]string{"error": "only the owner can manage sharing"})
	}
	return board, nil
}

func isSafeImageURL(raw string) bool {
	if strings.HasPrefix(raw, "/uploads/") {
		return true
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
