package main

import (
	"flowerpress/internal/config"
	"flowerpress/internal/database"
	"fmt"
	"log"
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

	fmt.Printf("flowerpress-next\n")
	fmt.Printf("environment: %s\n", cfg.Environment)
	fmt.Printf("address: %s\n", cfg.Address)
	fmt.Printf("database: %s\n", cfg.DatabasePath)
}
