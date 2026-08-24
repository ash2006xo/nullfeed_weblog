package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ash2006xo/nullfeed_weblog/internal/auth"
)

const contextUserIDKey = "user_id"
const contextUsernameKey = "username"

func RequireAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("auth_token")
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication required"})
			}

			claims, err := auth.ParseToken(cookie.Value, jwtSecret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired session"})
			}

			c.Set(contextUserIDKey, claims.UserID)
			c.Set(contextUsernameKey, claims.Username)

			return next(c)
		}
	}
}

func OptionalAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("auth_token")
			if err == nil {
				if claims, err := auth.ParseToken(cookie.Value, jwtSecret); err == nil {
					c.Set(contextUserIDKey, claims.UserID)
					c.Set(contextUsernameKey, claims.Username)
				}
			}
			return next(c)
		}
	}
}

func CurrentUserID(c echo.Context) (int, bool) {
	v := c.Get(contextUserIDKey)
	if v == nil {
		return 0, false
	}
	return v.(int), true
}