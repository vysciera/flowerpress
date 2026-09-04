package main

import (
	"flowerpress/internal/config"
	"flowerpress/internal/database"
	"flowerpress/internal/httpapi"
	"flowerpress/internal/service"
	"flowerpress/internal/store/turso"

	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Repos
	userRepository := turso.NewUserRepository(db)
	sessionRepository := turso.NewSessionRepository(db)
	projectRepository := turso.NewProjectRepository(db)

	// Services
	userService := service.NewUserService(userRepository)

	sessionService := service.NewSessionService(
		sessionRepository,
		userRepository,
		7*24*time.Hour,
	)

	projectService := service.NewProjectService(projectRepository)

	apiServer := httpapi.NewServer(
		userService,
		sessionService,
		projectService,
		cfg.SecureCookies,
	)

	httpServer := httpapi.NewHTTPServer(
		cfg.Address,
		apiServer.Handler(),
	)

	shutdown := make(chan os.Signal, 1)

	signal.Notify(
		shutdown,
		os.Interrupt,
		syscall.SIGTERM,
	)

	// Init Go routine
	go func() {
		log.Printf(
			"flowerpress is listening on %s",
			cfg.Address,
		)

		err := httpServer.ListenAndServe()

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server: %v", err)
		}
	}()

	// Graceful shutdown
	<-shutdown

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf(
			"HTTP shutdown: %v",
			err,
		)
	}

	log.Println("flowerpress stopped")
}
