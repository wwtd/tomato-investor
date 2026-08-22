package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type VoiceHandler struct {
	DB      *sql.DB
	DataDir string
}

// RegisterRoutes 把 voice 相关路由挂到已有 session 路由已是 /sessions/{sid}/voice，
// 这里仅提供静态文件服务 /api/voice/{filename}
func (h *VoiceHandler) FileRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{filename}", h.serveFile)
	return r
}

func (h *VoiceHandler) serveFile(w http.ResponseWriter, r *http.Request) {
	fn := chi.URLParam(r, "filename")
	// 防路径穿越
	clean := filepath.Clean(fn)
	if clean == ".." || len(clean) > 0 && clean[0] == '/' {
		writeErr(w, 400, "bad filename")
		return
	}
	full := filepath.Join(h.DataDir, "voice", clean)
	if _, err := os.Stat(full); err != nil {
		writeErr(w, 404, "not found")
		return
	}
	http.ServeFile(w, r, full)
}

// Upload 处理 /sessions/{sid}/voice multipart 上传
func Upload(db *sql.DB, dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid, err := parseURLInt(w, r, "sid")
		if err != nil {
			return
		}
		// 仅活跃会话可上传
		var status string
		db.QueryRow(`SELECT status FROM sessions WHERE id=?`, sid).Scan(&status)
		if status != "running" && status != "paused" {
			writeErr(w, 400, "session not active")
			return
		}
		// 限制 20MB
		r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			writeErr(w, 400, "upload too large or malformed: "+err.Error())
			return
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			writeErr(w, 400, "file field required")
			return
		}
		defer file.Close()
		text := r.FormValue("text")
		durationMs, _ := strconv.Atoi(r.FormValue("duration_ms"))

		// 生成唯一文件名
		ext := filepath.Ext(hdr.Filename)
		if ext == "" {
			ext = ".m4a"
		}
		buf := make([]byte, 8)
		rand.Read(buf)
		fn := hex.EncodeToString(buf) + ext

		voiceDir := filepath.Join(dataDir, "voice")
		os.MkdirAll(voiceDir, 0o755)
		dst := filepath.Join(voiceDir, fn)
		out, err := os.Create(dst)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			writeErr(w, 500, err)
			return
		}
		out.Close()

		mime := hdr.Header.Get("Content-Type")
		if mime == "" {
			mime = "application/octet-stream"
		}
		// 创建 voice_note 事件
		payload, _ := json.Marshal(map[string]any{"file": fn, "duration_ms": durationMs})
		res, err := db.Exec(`INSERT INTO session_events(session_id, type, payload, voice_file) VALUES(?, 'voice_note', ?, ?)`,
			sid, string(payload), fn)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		eid, _ := res.LastInsertId()
		_, err = db.Exec(`INSERT INTO voice_notes(session_event_id, file_path, mime, duration_ms, text) VALUES(?, ?, ?, ?, ?)`,
			eid, fn, mime, durationMs, text)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, map[string]any{
			"ok":         true,
			"file":       fn,
			"event_id":   eid,
			"text":       text,
			"duration_ms": durationMs,
		})
	}
}
