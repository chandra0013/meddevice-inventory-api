package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/chandra0013/meddevice-inventory-api/internal/api"
	"github.com/chandra0013/meddevice-inventory-api/internal/store"
)

func main() {
	dsn := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/meddevices?sslmode=disable")
	apiKey := getenv("API_KEY", "dev-api-key")
	addr := getenv("ADDR", ":8080")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("connect database: %v", err)
	}

	deviceStore := store.NewPostgresStore(db)
	if err := deviceStore.Migrate(); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	server := api.NewServer(deviceStore, apiKey)
	log.Printf("medical device inventory API listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
