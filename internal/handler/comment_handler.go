package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	custommw "github.com/ash2006xo/nullfeed_weblog/internal/middleware"
	"github.com/ash2006xo/nullfeed_weblog/internal/repository"
)

type CommentHandler struct {
	commentRepo *repository.CommentRepository
	boardRepo   *repository.BoardRepository
}

func NewCommentHandler(commentRepo *repository.CommentRepository, boardRepo *repository.BoardRepository) *CommentHandler {
	return &CommentHandler{commentRepo: commentRepo, boardRepo: boardRepo}
}

type createCommentRequest struct {
	Content string `json:"content"`
}

func (h *CommentHandler) Create(c echo.Context) error {
	userID, ok := custommw.CurrentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}
	username, _ := custommw.CurrentUsername(c)

	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid board id"})
	}

	var req createCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content is required"})
	}

	board, err := h.boardRepo.GetByID(boardID)
	if err != nil {
		if errors.Is(err, repository.ErrBoardNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "board not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load board"})
	}

	canView, err := h.boardRepo.CanView(board, &userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to check permissions"})
	}
	if !canView {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "board not found"})
	}

	comment, err := h.commentRepo.Create(boardID, userID, req.Content, username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create comment"})
	}

	return c.JSON(http.StatusCreated, comment)
}

func (h *CommentHandler) List(c echo.Context) error {
	boardID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid board id"})
	}

	board, err := h.boardRepo.GetByID(boardID)
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

	comments, err := h.commentRepo.ListByBoard(boardID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load comments"})
	}

	return c.JSON(http.StatusOK, comments)
}