package model

import "time"

// Project 项目
type Project struct {
	ID             int64   `json:"id"`
	OwnerID        string  `json:"owner_id"`
	Title          string  `json:"title"`
	Necessity      string  `json:"necessity"`
	MvpScope       string  `json:"mvp_scope"`
	StoplossNote   string  `json:"stoploss_note"`
	BudgetTomatoes int     `json:"budget_tomatoes"`
	TomatoMinutes  *int    `json:"tomato_minutes"` // nil = 用全局
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	ArchivedAt     *string `json:"archived_at"`

	// 聚合字段（详情用）
	Milestones       []Milestone `json:"milestones,omitempty"`
	ConsumedTomatoes float64     `json:"consumed_tomatoes"`
	ConsumedMinutes  int         `json:"consumed_minutes"`
}

type Milestone struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	OrderIdx  int    `json:"order_idx"`
	CreatedAt string `json:"created_at"`
}

// Session 番茄会话
type Session struct {
	ID             int64   `json:"id"`
	ProjectID      int64   `json:"project_id"`
	Status         string  `json:"status"`
	StartedAt      string  `json:"started_at"`
	EndedAt        *string `json:"ended_at"`
	ConsumedTomato float64 `json:"consumed_tomato"`
	Note           string  `json:"note"`

	Events []SessionEvent `json:"events,omitempty"`
}

type SessionEvent struct {
	ID        int64   `json:"id"`
	SessionID int64   `json:"session_id"`
	Type      string  `json:"type"` // pause|resume|comment|voice_note
	Payload   string  `json:"payload"`
	VoiceFile *string `json:"voice_file"`
	At        string  `json:"at"`

	VoiceNote *VoiceNote `json:"voice_note,omitempty"`
}

type VoiceNote struct {
	ID             int64  `json:"id"`
	SessionEventID int64  `json:"session_event_id"`
	FilePath       string `json:"file_path"`
	Mime           string `json:"mime"`
	DurationMs     int    `json:"duration_ms"`
	Text           string `json:"text"`
	CreatedAt      string `json:"created_at"`
}

type Settings struct {
	TomatoMinutes int `json:"tomato_minutes"`
}

// 请求体
type CreateProjectReq struct {
	Title          string `json:"title"`
	Necessity      string `json:"necessity"`
	MvpScope       string `json:"mvp_scope"`
	StoplossNote   string `json:"stoploss_note"`
	BudgetTomatoes int    `json:"budget_tomatoes"`
	TomatoMinutes  *int   `json:"tomato_minutes"`
}

type UpdateProjectReq struct {
	Title          *string `json:"title"`
	Necessity      *string `json:"necessity"`
	MvpScope       *string `json:"mvp_scope"`
	StoplossNote   *string `json:"stoploss_note"`
	BudgetTomatoes *int    `json:"budget_tomatoes"`
	TomatoMinutes  *int    `json:"tomato_minutes"`
}

type StatusReq struct {
	Status string `json:"status"` // completed | archived_stoploss
}

type CreateMilestoneReq struct {
	Title    string `json:"title"`
	OrderIdx int    `json:"order_idx"`
}

type UpdateMilestoneReq struct {
	Title    *string `json:"title"`
	Done     *bool   `json:"done"`
	OrderIdx *int    `json:"order_idx"`
}

type EndSessionReq struct {
	Note string `json:"note"`
}

type CommentReq struct {
	Text string `json:"text"`
}

type SettingsReq struct {
	TomatoMinutes int `json:"tomato_minutes"`
}

// 防止 time 未用告警（后续可能扩展）
var _ = time.Now
