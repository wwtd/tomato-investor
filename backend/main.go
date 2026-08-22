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
	// 默认绑定 0.0.0.0 而非 :7800 —— Go 的 ":port" 会绑成 IPv6-only(::)双栈，
	// 部分客户端(安卓/非 Linux)对 IPv4-mapped 处理不一致导致外部无法访问。
	// 显式 0.0.0.0 强制 IPv4 全地址监听，局域网/外网设备均可连。
	addr := flag.String("addr", "0.0.0.0:7800", "listen address")
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
