package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// @title Tibta65 API
// @version 1.0
// @description API backend untuk Tibta65 — auth admin, member, kegiatan, dll.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ketik "Bearer" diikuti spasi dan JWT token. Contoh: "Bearer eyJhbGc..."
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/hashgen/main.go <password>")
		os.Exit(1)
	}

	password := os.Args[1]

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		os.Exit(1)
	}

	fmt.Println(string(hash))
}
