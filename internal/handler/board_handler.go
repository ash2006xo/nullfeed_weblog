package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	custommw "github.com/ash2006xo/nullfeed_weblog/internal/middleware"
	"github.com/ash2006xo/nullfeed_weblog/internal/repository"
)

type BoardHandler struct {
	boardRepo *repository.BoardRepository
}

func NewBoardHandler(boardRepo *repository.BoardRepository) *BoardHandler {
	return &BoardHandler{boardRepo: boardRepo}
}

type createBoardRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	ImageURL   *string  `json:"image_url"`
	IsPrivate  bool     `json:"is_private"`
	ShareWith  []string `json:"share_with"`
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
	if req.Title == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title and content are required"})
	}

	board, err := h.boardRepo.Create(req.Title, req.Content, req.ImageURL, req.IsPrivate, userID, username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create board"})
	}

	if req.IsPrivate && len(req.ShareWith) > 0 {
		if err := h.boardRepo.AddShares(board.ID, req.ShareWith); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "board created, but sharing failed"})
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