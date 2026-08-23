package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ash2006xo/nullfeed_weblog/internal/config"
	"github.com/ash2006xo/nullfeed_weblog/internal/db"
	"github.com/ash2006xo/nullfeed_weblog/internal/handler"
	"github.com/ash2006xo/nullfeed_weblog/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database)
	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		if err := database.Ping(); err != nil {
			return c.JSON(500, map[string]string{"status": "error", "db": "unreachable"})
		}
		return c.JSON(200, map[string]string{"status": "ok", "db": "connected"})
	})

	api := e.Group("/api")
	api.POST("/signup", authHandler.Signup)
	api.POST("/login", authHandler.Login)

	log.Printf("starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}