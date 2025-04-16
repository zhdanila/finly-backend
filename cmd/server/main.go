package main

import (
	"finly-backend/internal/app/bootstrap"
	"fmt"
	"os"
)

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	fmt.Println("Starting Finly Backend...")
	fmt.Printf("HTTP Port: %s\n", httpPort)
	fmt.Printf("Database Username: %s\n", dbUsername)
	fmt.Println("Database Password:]", dbPassword)
	fmt.Printf("Database Host: %s\n", dbHost)
	fmt.Printf("Database Port: %s\n", dbPort)
	fmt.Printf("Database Name: %s\n", dbName)
	fmt.Printf("Database SSL Mode: %s\n", dbSSLMode)
	fmt.Printf("Redis Host: %s\n", redisHost)
	fmt.Printf("Redis Port: %s\n", redisPort)
	// Initialize the website application
	bootstrap.Website()
}
