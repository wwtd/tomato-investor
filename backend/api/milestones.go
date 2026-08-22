package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"tomato-investor/model"
)

type MilestoneHandler struct {
	DB *sql.DB
}

func (h *MilestoneHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Route("/{mid}", func(r chi.Router) {
		r.Patch("/", h.update)
		r.Delete("/", h.del)
	})
	return r
}

// OpsRouter 暴露 /milestones/{mid} 操作（独立路径，不依赖 project 上下文）
func (h *MilestoneHandler) OpsRouter() chi.Router {
	r := chi.NewRouter()
	r.Route("/{mid}", func(r chi.Router) {
		r.Patch("/", h.update)
		r.Delete("/", h.del)
	})
	return r
}

func (h *MilestoneHandler) create(w http.ResponseWriter, r *http.Request) {
	pid, ok := idParam(w, r)
	if !ok {
		return
	}
	var req model.CreateMilestoneReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Title == "" {
		writeErr(w, 400, "title required")
		return
	}
	if req.OrderIdx == 0 {
		var maxIdx sql.NullInt64
		h.DB.QueryRow(`SELECT MAX(order_idx) FROM milestones WHERE project_id=?`, pid).Scan(&maxIdx)
		req.OrderIdx = int(maxIdx.Int64) + 1
	}
	res, err := h.DB.Exec(`INSERT INTO milestones(project_id, title, done, order_idx) VALUES(?, ?, 0, ?)`,
		pid, req.Title, req.OrderIdx)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	id, _ := res.LastInsertId()
	var m model.Milestone
	var done int
	h.DB.QueryRow(`SELECT id, project_id, title, done, order_idx, created_at FROM milestones WHERE id=?`, id).
		Scan(&m.ID, &m.ProjectID, &m.Title, &done, &m.OrderIdx, &m.CreatedAt)
	m.Done = done == 1
	writeJSON(w, m)
}

func (h *MilestoneHandler) update(w http.ResponseWriter, r *http.Request) {
	mid, err := parseURLInt(w, r, "mid")
	if err != nil {
		return
	}
	var req model.UpdateMilestoneReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Title != nil {
		h.DB.Exec(`UPDATE milestones SET title=? WHERE id=?`, *req.Title, mid)
	}
	if req.Done != nil {
		v := 0
		if *req.Done {
			v = 1
		}
		h.DB.Exec(`UPDATE milestones SET done=? WHERE id=?`, v, mid)
	}
	if req.OrderIdx != nil {
		h.DB.Exec(`UPDATE milestones SET order_idx=? WHERE id=?`, *req.OrderIdx, mid)
	}
	var m model.Milestone
	var done int
	h.DB.QueryRow(`SELECT id, project_id, title, done, order_idx, created_at FROM milestones WHERE id=?`, mid).
		Scan(&m.ID, &m.ProjectID, &m.Title, &done, &m.OrderIdx, &m.CreatedAt)
	m.Done = done == 1
	writeJSON(w, m)
}

func (h *MilestoneHandler) del(w http.ResponseWriter, r *http.Request) {
	mid, err := parseURLInt(w, r, "mid")
	if err != nil {
		return
	}
	h.DB.Exec(`DELETE FROM milestones WHERE id=?`, mid)
	w.WriteHeader(204)
}
