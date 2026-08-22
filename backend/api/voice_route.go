package api

import (
	"database/sql"
	"net/http"
	"os"
)

// VoiceDataDir 从环境变量 TOMATO_DATA_DIR 读取，默认 ./data
func VoiceDataDir() string {
	d := os.Getenv("TOMATO_DATA_DIR")
	if d == "" {
		d = "data"
	}
	return d
}

// voiceHandler 在 SessionHandler 中注册 /sessions/{sid}/voice
// 由于 voice.go 里的 Upload 是包级函数，这里提供一个方法包装
func (h *SessionHandler) voice(w http.ResponseWriter, r *http.Request) {
	Upload(h.DB, VoiceDataDir())(w, r)
}

var _ = sql.ErrNoRows
