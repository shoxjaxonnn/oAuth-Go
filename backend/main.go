package main

import (
	"log"
	"net/http"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/config"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/db"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/server"
)

func main() {
	cfg := config.Load()

	database, err := db.Open("data.sqlite")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	handler := server.New(cfg, database, nil)

	log.Printf("backend listening on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}
