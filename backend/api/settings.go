package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"tomato-investor/model"
)

type SettingsHandler struct {
	DB *sql.DB
}

func (h *SettingsHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.get)
	r.Put("/", h.update)
	return r
}

func (h *SettingsHandler) get(w http.ResponseWriter, r *http.Request) {
	var s model.Settings
	h.DB.QueryRow(`SELECT tomato_minutes FROM users WHERE id=1`).Scan(&s.TomatoMinutes)
	writeJSON(w, s)
}

func (h *SettingsHandler) update(w http.ResponseWriter, r *http.Request) {
	var req model.SettingsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.TomatoMinutes < 1 {
		writeErr(w, 400, "tomato_minutes must be >= 1")
		return
	}
	if _, err := h.DB.Exec(`UPDATE users SET tomato_minutes=? WHERE id=1`, req.TomatoMinutes); err != nil {
		writeErr(w, 500, err)
		return
	}
	var s model.Settings
	h.DB.QueryRow(`SELECT tomato_minutes FROM users WHERE id=1`).Scan(&s.TomatoMinutes)
	writeJSON(w, s)
}
