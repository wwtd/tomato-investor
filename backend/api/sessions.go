package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"tomato-investor/model"
)

type SessionHandler struct {
	DB *sql.DB
}

func (h *SessionHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listByProject)
	r.Post("/", h.start)
	r.Route("/{sid}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Post("/pause", h.pause)
		r.Post("/resume", h.resume)
		r.Post("/comment", h.comment)
		r.Post("/voice", h.voice)
		r.Post("/end", h.end)
	})
	return r
}

// SessionOpsRouter 暴露 /sessions/{sid} 操作（不依赖 project 路由上下文）
func (h *SessionHandler) SessionOpsRouter() chi.Router {
	r := chi.NewRouter()
	r.Route("/{sid}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Post("/pause", h.pause)
		r.Post("/resume", h.resume)
		r.Post("/comment", h.comment)
		r.Post("/voice", h.voice)
		r.Post("/end", h.end)
	})
	return r
}

func (h *SessionHandler) listByProject(w http.ResponseWriter, r *http.Request) {
	pid, ok := idParam(w, r)
	if !ok {
		return
	}
	rows, err := h.DB.Query(`SELECT id, project_id, status, started_at, ended_at, consumed_tomato, note
		FROM sessions WHERE project_id=? ORDER BY started_at DESC`, pid)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	defer rows.Close()
	var out []model.Session
	for rows.Next() {
		out = append(out, scanSessionRow(rows))
	}
	if out == nil {
		out = []model.Session{}
	}
	writeJSON(w, out)
}

func (h *SessionHandler) start(w http.ResponseWriter, r *http.Request) {
	pid, ok := idParam(w, r)
	if !ok {
		return
	}
	// 检查是否已有 running/paused 会话
	var cnt int
	h.DB.QueryRow(`SELECT COUNT(*) FROM sessions WHERE project_id=? AND status IN ('running','paused')`, pid).Scan(&cnt)
	if cnt > 0 {
		writeErr(w, 400, "project already has an active session; end it first")
		return
	}
	res, err := h.DB.Exec(`INSERT INTO sessions(project_id, status, started_at) VALUES(?, 'running', datetime('now'))`, pid)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	id, _ := res.LastInsertId()
	h.fetchAndWrite(w, id)
}

func (h *SessionHandler) get(w http.ResponseWriter, r *http.Request) {
	sid, err := parseURLInt(w, r, "sid")
	if err != nil {
		return
	}
	h.fetchAndWrite(w, sid)
}

func (h *SessionHandler) pause(w http.ResponseWriter, r *http.Request) {
	sid, err := parseURLInt(w, r, "sid")
	if err != nil {
		return
	}
	var req model.CommentReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if !h.mustStatus(w, sid, "running") {
		return
	}
	if _, err := h.DB.Exec(`UPDATE sessions SET status='paused' WHERE id=?`, sid); err != nil {
		writeErr(w, 500, err)
		return
	}
	payload := "{}"
	if req.Text != "" {
		b, _ := json.Marshal(map[string]string{"comment": req.Text})
		payload = string(b)
	}
	h.DB.Exec(`INSERT INTO session_events(session_id, type, payload) VALUES(?, 'pause', ?)`, sid, payload)
	h.fetchAndWrite(w, sid)
}

func (h *SessionHandler) resume(w http.ResponseWriter, r *http.Request) {
	sid, err := parseURLInt(w, r, "sid")
	if err != nil {
		return
	}
	if !h.mustStatus(w, sid, "paused") {
		return
	}
	if _, err := h.DB.Exec(`UPDATE sessions SET status='running' WHERE id=?`, sid); err != nil {
		writeErr(w, 500, err)
		return
	}
	h.DB.Exec(`INSERT INTO session_events(session_id, type) VALUES(?, 'resume')`, sid)
	h.fetchAndWrite(w, sid)
}

func (h *SessionHandler) comment(w http.ResponseWriter, r *http.Request) {
	sid, err := parseURLInt(w, r, "sid")
	if err != nil {
		return
	}
	var req model.CommentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, err)
		return
	}
	if req.Text == "" {
		writeErr(w, 400, "text required")
		return
	}
	// 仅运行或暂停态可以加 comment
	var status string
	h.DB.QueryRow(`SELECT status FROM sessions WHERE id=?`, sid).Scan(&status)
	if status != "running" && status != "paused" {
		writeErr(w, 400, "session not active")
		return
	}
	b, _ := json.Marshal(map[string]string{"comment": req.Text})
	res, err := h.DB.Exec(`INSERT INTO session_events(session_id, type, payload) VALUES(?, 'comment', ?)`, sid, string(b))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	eid, _ := res.LastInsertId()
	writeJSON(w, map[string]any{"id": eid, "ok": true})
}

func (h *SessionHandler) end(w http.ResponseWriter, r *http.Request) {
	sid, err := parseURLInt(w, r, "sid")
	if err != nil {
		return
	}
	var req model.EndSessionReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	var status string
	h.DB.QueryRow(`SELECT status FROM sessions WHERE id=?`, sid).Scan(&status)
	if status == "ended" {
		writeErr(w, 400, "session already ended")
		return
	}
	if status == "" {
		writeErr(w, 404, "session not found")
		return
	}
	consume := 1
	if !req.ConsumeTomato {
		consume = 0
	}
	// 若中途暂停未恢复，补一个 pause 事件已在；end 时直接 ended
	if _, err := h.DB.Exec(`UPDATE sessions SET status='ended', ended_at=datetime('now'), consumed_tomato=?, note=? WHERE id=?`,
		consume, req.Note, sid); err != nil {
		writeErr(w, 500, err)
		return
	}
	h.fetchAndWrite(w, sid)
}

func (h *SessionHandler) fetchAndWrite(w http.ResponseWriter, sid int64) {
	s, err := fetchSession(h.DB, sid)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, s)
}

func (h *SessionHandler) mustStatus(w http.ResponseWriter, sid int64, want string) bool {
	var status string
	err := h.DB.QueryRow(`SELECT status FROM sessions WHERE id=?`, sid).Scan(&status)
	if err == sql.ErrNoRows {
		writeErr(w, 404, "session not found")
		return false
	}
	if err != nil {
		writeErr(w, 500, err)
		return false
	}
	if status != want {
		writeErr(w, 400, fmt.Sprintf("session status is %s, need %s", status, want))
		return false
	}
	return true
}

func scanSessionRow(rows *sql.Rows) model.Session {
	var s model.Session
	var ended sql.NullString
	rows.Scan(&s.ID, &s.ProjectID, &s.Status, &s.StartedAt, &ended, &s.ConsumedTomato, &s.Note)
	if ended.Valid {
		s.EndedAt = &ended.String
	}
	return s
}

func fetchSession(d *sql.DB, sid int64) (model.Session, error) {
	var s model.Session
	var ended sql.NullString
	err := d.QueryRow(`SELECT id, project_id, status, started_at, ended_at, consumed_tomato, note FROM sessions WHERE id=?`, sid).
		Scan(&s.ID, &s.ProjectID, &s.Status, &s.StartedAt, &ended, &s.ConsumedTomato, &s.Note)
	if err != nil {
		return s, err
	}
	if ended.Valid {
		s.EndedAt = &ended.String
	}
	// events + voice_notes
	rows, err := d.Query(`SELECT se.id, se.session_id, se.type, se.payload, se.voice_file, se.at,
		vn.id, vn.session_event_id, vn.file_path, vn.mime, vn.duration_ms, vn.text, vn.created_at
		FROM session_events se
		LEFT JOIN voice_notes vn ON vn.session_event_id=se.id
		WHERE se.session_id=? ORDER BY se.id`, sid)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	for rows.Next() {
		var e model.SessionEvent
		var (
			vnID                                            sql.NullInt64
			vnEID                                           sql.NullInt64
			vnFile, vnMime, vnText, vnCreated               sql.NullString
			vnDur                                           sql.NullInt64
		)
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Type, &e.Payload, &e.VoiceFile, &e.At,
			&vnID, &vnEID, &vnFile, &vnMime, &vnDur, &vnText, &vnCreated); err != nil {
			return s, err
		}
		if vnID.Valid {
			e.VoiceNote = &model.VoiceNote{
				ID:             vnID.Int64,
				SessionEventID: vnEID.Int64,
				FilePath:       vnFile.String,
				Mime:           vnMime.String,
				DurationMs:     int(vnDur.Int64),
				Text:           vnText.String,
				CreatedAt:      vnCreated.String,
			}
		}
		s.Events = append(s.Events, e)
	}
	return s, nil
}

// ===== 时长计算 =====
// sessionMinutesByID 计算单会话运行分钟：事件流重建 running 段总和。
// 段定义：起始→pause，resume→pause，resume→end。
func sessionMinutesByID(d *sql.DB, sid int64, started string, ended sql.NullString, status string) int {
	rows, err := d.Query(`SELECT type, at FROM session_events WHERE session_id=? AND type IN ('pause','resume') ORDER BY id`, sid)
	if err != nil {
		return 0
	}
	defer rows.Close()
	type ev struct{ t, at string }
	var events []ev
	for rows.Next() {
		var e ev
		rows.Scan(&e.t, &e.at)
		events = append(events, e)
	}
	endAt := ""
	if ended.Valid {
		endAt = ended.String
	}
	if len(events) == 0 {
		if status == "ended" && endAt != "" {
			return durMin(started, endAt)
		}
		if status == "running" || status == "paused" {
			return durMin(started, nowUTC())
		}
		return 0
	}
	var total int
	running := true
	segStart := started
	for _, e := range events {
		if e.t == "pause" && running {
			total += durMin(segStart, e.at)
			running = false
		} else if e.t == "resume" {
			segStart = e.at
			running = true
		}
	}
	if running {
		end := endAt
		if end == "" {
			end = nowUTC()
		}
		total += durMin(segStart, end)
	}
	return total
}

func durMin(a, b string) int {
	const layout = "2006-01-02 15:04:05"
	ta, err1 := time.Parse(layout, a)
	tb, err2 := time.Parse(layout, b)
	if err1 != nil || err2 != nil {
		return 0
	}
	d := tb.Sub(ta)
	if d < 0 {
		return 0
	}
	return int(d.Minutes())
}

func nowUTC() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

func parseURLInt(w http.ResponseWriter, r *http.Request, key string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(chi.URLParam(r, key), "%d", &v)
	if err != nil {
		writeErr(w, 400, "invalid "+key)
		return 0, err
	}
	return v, nil
}
