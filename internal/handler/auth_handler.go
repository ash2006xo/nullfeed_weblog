package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/ash2006xo/nullfeed_weblog/internal/auth"
	custommw "github.com/ash2006xo/nullfeed_weblog/internal/middleware"
	"github.com/ash2006xo/nullfeed_weblog/internal/repository"
)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

type signupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Signup(c echo.Context) error {
	var req signupRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username and password are required"})
	}
	if len(req.Username) < 3 || len(req.Username) > 30 || !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(req.Username) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username must be 3-30 characters and use only letters, numbers, _ or -"})
	}
	if len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to process password"})
	}

	user, err := h.userRepo.Create(req.Username, hash)
	if err != nil {
		if errors.Is(err, repository.ErrUsernameTaken) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "username already taken"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	token, err := auth.GenerateToken(user.ID, user.Username, h.jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	setAuthCookie(c, token)

	return c.JSON(http.StatusCreated, user)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	req.Username = strings.TrimSpace(req.Username)
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
	}

	token, err := auth.GenerateToken(user.ID, user.Username, h.jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	setAuthCookie(c, token)

	return c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := custommw.CurrentUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
	}
	username, _ := custommw.CurrentUsername(c)
	return c.JSON(http.StatusOK, map[string]interface{}{"id": userID, "username": username})
}

func setAuthCookie(c echo.Context, token string) {
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = token
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.HttpOnly = true
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
	c.SetCookie(cookie)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	cookie.HttpOnly = true
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Secure = c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https"
	c.SetCookie(cookie)
	return c.NoContent(http.StatusOK)
}
