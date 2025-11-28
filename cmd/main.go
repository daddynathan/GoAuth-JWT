package main

import (
	"friend-help/internal/cache"
	"friend-help/internal/repo"
	"friend-help/internal/service"
	"friend-help/internal/transport/https"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// @title           inmt API
// @version         1.0
// @description     API для инмт сайта
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := godotenv.Load("app.env"); err != nil {
		log.Fatal("could not load .env file. Using OS environment variables: ", err)
	}
	db, err := repo.ConnectToBase()
	if err != nil {
		log.Fatal("DB connection failed: ", err)
	}
	if err := repo.RunMigrations(db); err != nil {
		log.Fatal("could not run migration: ", err)
	}
	mysqlAuthRepo := repo.NewmysqlAuthRepo(db)
	jwtService, err := service.NewJwtService()
	if err != nil {
		log.Fatal("JWT init failed: ", err)
	}
	cache, err := cache.NewRedisService()
	if err != nil {
		log.Fatal("Redis init failed: ", err)
	}
	authService := service.NewAuthService(mysqlAuthRepo, jwtService, cache)
	HTTPHandlers := https.NewHTTPHandlers(authService)
	port := os.Getenv("APP_PORT")
	if port == "" {
		log.Fatal("APP_PORT not set in environment or .env file.")
	}
	https.NewHTTPServer(HTTPHandlers, port)
}
