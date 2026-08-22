package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// Open 打开/创建 SQLite 数据库，执行 schema，保证单用户存在。
// dbPath 为空则用 ./data/tomato.db
func Open(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "data/tomato.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	d, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	// sqlite 写串行化，避免锁问题
	if _, err := d.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		d.Close()
		return nil, err
	}
	if _, err := d.Exec(schemaSQL); err != nil {
		d.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	// 确保单用户存在
	if _, err := d.Exec(`INSERT OR IGNORE INTO users(id, owner_id, tomato_minutes) VALUES(1,'me',480)`); err != nil {
		d.Close()
		return nil, fmt.Errorf("seed user: %w", err)
	}
	return d, nil
}
