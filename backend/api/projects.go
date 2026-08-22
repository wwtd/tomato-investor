package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"tomato-investor/model"
)

type ProjectHandler struct {
	DB *sql.DB
}

func (h *ProjectHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Patch("/status", h.status)
		r.Delete("/", h.del)
		r.Mount("/milestones", (&MilestoneHandler{DB: h.DB}).Router())
		r.Mount("/sessions", (&SessionHandler{DB: h.DB}).Router())
		r.Get("/stats", h.stats)
	})
	return r
}

func (h *ProjectHandler) list(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	rows, err := h.DB.Query(`SELECT id, owner_id, title, necessity, mvp_scope, stoploss_note,
		budget_tomatoes, tomato_minutes, status, created_at, archived_at FROM projects
		WHERE 1=1`+condStatus(status)+` ORDER BY created_at DESC`, )
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	defer rows.Close()
	var out []model.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		out = append(out, p)
	}
	if out == nil {
		out = []model.Project{}
	}
	writeJSON(w, out)
}

func condStatus(status string) string {
	if status == "" {
		return ""
	}
	return ` AND status='` + status + `'`
}

func (h *ProjectHandler) create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Title == "" {
		writeErr(w, 400, "title required")
		return
	}
	if req.BudgetTomatoes < 1 {
		req.BudgetTomatoes = 1
	}
	res, err := h.DB.Exec(`INSERT INTO projects(owner_id, title, necessity, mvp_scope, stoploss_note, budget_tomatoes, tomato_minutes)
		VALUES('me', ?, ?, ?, ?, ?, ?)`,
		req.Title, req.Necessity, req.MvpScope, req.StoplossNote, req.BudgetTomatoes, req.TomatoMinutes)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	id, _ := res.LastInsertId()
	h.fetchAndWrite(w, id)
}

func (h *ProjectHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	h.fetchAndWrite(w, id)
}

func (h *ProjectHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var req model.UpdateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Title != nil {
		if _, err := h.DB.Exec(`UPDATE projects SET title=? WHERE id=?`, *req.Title, id); err != nil {
			writeErr(w, 500, err)
			return
		}
	}
	if req.Necessity != nil {
		h.DB.Exec(`UPDATE projects SET necessity=? WHERE id=?`, *req.Necessity, id)
	}
	if req.MvpScope != nil {
		h.DB.Exec(`UPDATE projects SET mvp_scope=? WHERE id=?`, *req.MvpScope, id)
	}
	if req.StoplossNote != nil {
		h.DB.Exec(`UPDATE projects SET stoploss_note=? WHERE id=?`, *req.StoplossNote, id)
	}
	if req.BudgetTomatoes != nil {
		h.DB.Exec(`UPDATE projects SET budget_tomatoes=? WHERE id=?`, *req.BudgetTomatoes, id)
	}
	if req.TomatoMinutes != nil {
		h.DB.Exec(`UPDATE projects SET tomato_minutes=? WHERE id=?`, *req.TomatoMinutes, id)
	}
	h.fetchAndWrite(w, id)
}

func (h *ProjectHandler) status(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var req model.StatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Status != "completed" && req.Status != "archived_stoploss" {
		writeErr(w, 400, "status must be completed|archived_stoploss")
		return
	}
	if _, err := h.DB.Exec(`UPDATE projects SET status=?, archived_at=datetime('now') WHERE id=?`, req.Status, id); err != nil {
		writeErr(w, 500, err)
		return
	}
	h.fetchAndWrite(w, id)
}

func (h *ProjectHandler) del(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	var n int
	h.DB.QueryRow(`SELECT COUNT(*) FROM sessions WHERE project_id=?`, id).Scan(&n)
	if n > 0 {
		writeErr(w, 400, "cannot delete project with sessions; archive instead")
		return
	}
	h.DB.Exec(`DELETE FROM projects WHERE id=?`, id)
	w.WriteHeader(204)
}

func (h *ProjectHandler) stats(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(w, r)
	if !ok {
		return
	}
	consumed, minutes, err := projectUsage(h.DB, id)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var budget, pmins, gmins int
	h.DB.QueryRow(`SELECT budget_tomatoes, COALESCE(tomato_minutes,0), (SELECT tomato_minutes FROM users WHERE id=1) FROM projects WHERE id=?`, id).Scan(&budget, &pmins, &gmins)
	tmins := pmins
	if tmins == 0 {
		tmins = gmins
	}
	writeJSON(w, map[string]int{
		"budget_tomatoes":   budget,
		"consumed_tomatoes": consumed,
		"remaining_tomatoes": budget - consumed,
		"tomato_minutes":    tmins,
		"consumed_minutes":  minutes,
		"remaining_minutes": budget*tmins - minutes,
	})
}

func (h *ProjectHandler) fetchAndWrite(w http.ResponseWriter, id int64) {
	p, err := fetchProject(h.DB, id)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, p)
}

// ===== helpers =====

func idParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, 400, "invalid id")
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err any) {
	w.WriteHeader(code)
	writeJSON(w, map[string]string{"error": errToString(err)})
}

func errToString(v any) string {
	if e, ok := v.(error); ok {
		return e.Error()
	}
	s, _ := v.(string)
	return s
}

// scanProject 从当前行扫描项目（不含 milestones / 聚合）
func scanProject(rows *sql.Rows) (model.Project, error) {
	var p model.Project
	var archived sql.NullString
	err := rows.Scan(&p.ID, &p.OwnerID, &p.Title, &p.Necessity, &p.MvpScope, &p.StoplossNote,
		&p.BudgetTomatoes, &p.TomatoMinutes, &p.Status, &p.CreatedAt, &archived)
	if archived.Valid {
		p.ArchivedAt = &archived.String
	}
	return p, err
}

func fetchProject(d *sql.DB, id int64) (model.Project, error) {
	var p model.Project
	var archived sql.NullString
	err := d.QueryRow(`SELECT id, owner_id, title, necessity, mvp_scope, stoploss_note,
		budget_tomatoes, tomato_minutes, status, created_at, archived_at FROM projects WHERE id=?`, id).
		Scan(&p.ID, &p.OwnerID, &p.Title, &p.Necessity, &p.MvpScope, &p.StoplossNote,
			&p.BudgetTomatoes, &p.TomatoMinutes, &p.Status, &p.CreatedAt, &archived)
	if err != nil {
		return p, err
	}
	if archived.Valid {
		p.ArchivedAt = &archived.String
	}
	// milestones
	ms, _ := fetchMilestones(d, id)
	p.Milestones = ms
	// 聚合
	p.ConsumedTomatoes, p.ConsumedMinutes, _ = projectUsage(d, id)
	return p, nil
}

func fetchMilestones(d *sql.DB, projectID int64) ([]model.Milestone, error) {
	rows, err := d.Query(`SELECT id, project_id, title, done, order_idx, created_at FROM milestones WHERE project_id=? ORDER BY order_idx, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Milestone
	for rows.Next() {
		var m model.Milestone
		var done int
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Title, &done, &m.OrderIdx, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Done = done == 1
		out = append(out, m)
	}
	return out, nil
}

// projectUsage 返回 (已消耗番茄数, 已投入分钟)
func projectUsage(d *sql.DB, projectID int64) (int, int, error) {
	// 已结束且 consumed_tomato=1 的会话数 = 已消耗番茄
	var consumed int
	if err := d.QueryRow(`SELECT COUNT(*) FROM sessions WHERE project_id=? AND consumed_tomato=1`, projectID).Scan(&consumed); err != nil {
		return 0, 0, err
	}
	minutes, err := projectMinutes(d, projectID)
	return consumed, minutes, err
}

// projectMinutes 计算项目所有会话的运行分钟数（running 段总和）
func projectMinutes(d *sql.DB, projectID int64) (int, error) {
	rows, err := d.Query(`SELECT id, started_at, ended_at, status FROM sessions WHERE project_id=?`, projectID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	type sess struct {
		id      int64
		started string
		ended   sql.NullString
		status  string
	}
	var sessions []sess
	for rows.Next() {
		var s sess
		if err := rows.Scan(&s.id, &s.started, &s.ended, &s.status); err != nil {
			return 0, err
		}
		sessions = append(sessions, s)
	}
	total := 0
	for _, s := range sessions {
		total += sessionMinutesByID(d, s.id, s.started, s.ended, s.status)
	}
	return total, nil
}
