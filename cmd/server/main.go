package main

import (
	"log"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/ash2006xo/nullfeed_weblog/internal/config"
	"github.com/ash2006xo/nullfeed_weblog/internal/db"
	"github.com/ash2006xo/nullfeed_weblog/internal/handler"
	custommw "github.com/ash2006xo/nullfeed_weblog/internal/middleware"
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

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	userRepo := repository.NewUserRepository(database)
	boardRepo := repository.NewBoardRepository(database)
	commentRepo := repository.NewCommentRepository(database)

	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret)
	boardHandler := handler.NewBoardHandler(boardRepo)
	commentHandler := handler.NewCommentHandler(commentRepo, boardRepo)
	uploadHandler := handler.NewUploadHandler("web/uploads")

	e := echo.New()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	e.Static("/", "web")
	e.GET("/weblog/:id", func(c echo.Context) error {
		return c.File("web/board.html")
	})

	e.GET("/health", func(c echo.Context) error {
		if err := database.Ping(); err != nil {
			return c.JSON(500, map[string]string{"status": "error", "db": "unreachable"})
		}
		return c.JSON(200, map[string]string{"status": "ok", "db": "connected"})
	})

	api := e.Group("/api")

	api.POST("/signup", authHandler.Signup)
	api.POST("/login", authHandler.Login)
	api.POST("/logout", authHandler.Logout)
	api.GET("/me", authHandler.Me, custommw.RequireAuth(cfg.JWTSecret))

	api.GET("/boards", boardHandler.List, custommw.OptionalAuth(cfg.JWTSecret))
	api.GET("/boards/:id", boardHandler.Get, custommw.OptionalAuth(cfg.JWTSecret))
	api.POST("/boards", boardHandler.Create, custommw.RequireAuth(cfg.JWTSecret))
	api.DELETE("/boards/:id", boardHandler.Delete, custommw.RequireAuth(cfg.JWTSecret))
	api.GET("/boards/:id/shares", boardHandler.ListShares, custommw.RequireAuth(cfg.JWTSecret))
	api.PUT("/boards/:id/shares", boardHandler.ReplaceShares, custommw.RequireAuth(cfg.JWTSecret))
	api.POST("/uploads/image", uploadHandler.UploadImage, custommw.RequireAuth(cfg.JWTSecret))

	api.GET("/boards/:id/comments", commentHandler.List, custommw.OptionalAuth(cfg.JWTSecret))
	api.POST("/boards/:id/comments", commentHandler.Create, custommw.RequireAuth(cfg.JWTSecret))

	log.Printf("starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
