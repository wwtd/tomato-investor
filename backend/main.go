package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"tomato-investor/api"
	"tomato-investor/db"
)

func main() {
	addr := flag.String("addr", ":7800", "listen address")
	dbPath := flag.String("db", "", "sqlite path (default ./data/tomato.db)")
	flag.Parse()

	dataDir := os.Getenv("TOMATO_DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	if *dbPath == "" {
		*dbPath = filepath.Join(dataDir, "tomato.db")
	}

	d, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer d.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Mount("/projects", (&api.ProjectHandler{DB: d}).Router())
		r.Mount("/sessions", (&api.SessionHandler{DB: d}).SessionOpsRouter())
		r.Mount("/milestones", (&api.MilestoneHandler{DB: d}).OpsRouter())
		r.Mount("/settings", (&api.SettingsHandler{DB: d}).Router())
		r.Mount("/voice", (&api.VoiceHandler{DB: d, DataDir: dataDir}).FileRouter())
	})

	log.Printf("tomato-investor listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal(err)
	}
}
