package main

import(
	"fmt"
	"flowerpress/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Printf("flowerpress-next\n")
	fmt.Printf("environment: %s\n", cfg.Environment)
	fmt.Printf("address: %s\n", cfg.Address)
	fmt.Printf("database: %s\n", cfg.DatabasePath)
}
