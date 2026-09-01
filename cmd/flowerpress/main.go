package main

import (
	"flowerpress/internal/config"
	"flowerpress/internal/database"
	"flowerpress/internal/httpapi"
	"flowerpress/internal/service"
	"flowerpress/internal/store/turso"

	"log"
	"net/http"
	"time"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	userRepository := turso.NewUserRepository(db)
	sessionRepository := turso.NewSessionRepository(db)

	userService := service.NewUserService(userRepository)

	sessionService := service.NewSessionService(
		sessionRepository,
		userRepository,
		7*24*time.Hour,
	)

	server := httpapi.NewServer(userService, sessionService)
	log.Printf("flowerpress is listening on %s", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, server.Handler()); err != nil {
		log.Fatal(err)
	}
}
