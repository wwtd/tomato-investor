import { useEffect, useState } from "react";
import { api, type Project, type Session, type Stats } from "../api";
import { MilestoneList } from "./MilestoneList";
import { SessionRunner } from "./SessionRunner";

export function ProjectDetail({
  id,
  onBack,
}: {
  id: number;
  onBack: () => void;
}) {
  const [project, setProject] = useState<Project | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [activeSession, setActiveSession] = useState<Session | null>(null);
  const [err, setErr] = useState("");

  async function loadAll() {
    try {
      const [p, s, ss] = await Promise.all([api.getProject(id), api.stats(id), api.listSessions(id)]);
      setProject(p);
      setStats(s);
      setSessions(ss);
      const act = ss.find((x) => x.status === "running" || x.status === "paused");
      if (act) {
        const full = await api.getSession(act.id);
        setActiveSession(full);
      } else {
        setActiveSession(null);
      }
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  useEffect(() => {
    loadAll();
  }, [id]);

  async function startSession() {
    try {
      const s = await api.startSession(id);
      setActiveSession(s);
      loadAll();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  async function changeStatus(status: string) {
    if (!confirm(`确认${status === "completed" ? "完成" : "止损归档"}这个项目？`)) return;
    try {
      await api.setProjectStatus(id, status);
      loadAll();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  if (err) return <div className="error">{err}</div>;
  if (!project) return <div className="muted">加载中…</div>;

  return (
    <div className="project-detail">
      <button className="back-btn" onClick={onBack}>
        ← 返回
      </button>

      <div className="detail-header">
        <div>
          <h2>{project.title}</h2>
          <span className={`badge st-${project.status === "archived_stoploss" ? "stop" : project.status === "completed" ? "done" : "active"}`}>
            {project.status === "active" ? "进行中" : project.status === "completed" ? "已完成" : "止损归档"}
          </span>
        </div>
        {project.status === "active" && (
          <div className="detail-actions">
            {!activeSession && <button className="primary" onClick={startSession}>开始番茄</button>}
            <button onClick={() => changeStatus("completed")}>完成</button>
            <button className="danger" onClick={() => changeStatus("archived_stoploss")}>止损归档</button>
          </div>
        )}
      </div>

      <div className="detail-grid">
        <section>
          <h4>投资预算</h4>
          {stats && (
            <div className="budget-box">
              <div className="budget-row">
                <span>🍅 已投入 / 预算</span>
                <strong>{fmtTomato(stats.consumed_tomatoes)} / {stats.budget_tomatoes}</strong>
              </div>
              <div className="progress">
                <div
                  className="progress-bar"
                  style={{ width: `${stats.budget_tomatoes ? (Math.min(stats.consumed_tomatoes / stats.budget_tomatoes, 1) * 100).toFixed(1) : 0}%` }}
                />
              </div>
              <div className="budget-row">
                <span>⏱ 已投入 / 预算</span>
                <strong>{fmtMin(stats.consumed_minutes)} / {fmtMin(stats.budget_tomatoes * stats.tomato_minutes)}</strong>
              </div>
              <div className="budget-row">
                <span>剩余</span>
                <strong className={stats.remaining_tomatoes <= 0 ? "danger" : ""}>
                  🍅 {fmtTomato(stats.remaining_tomatoes)} ({fmtMin(stats.remaining_minutes)})
                </strong>
              </div>
            </div>
          )}
        </section>

        <section>
          <h4>必要性</h4>
          <p>{project.necessity || "—"}</p>
          <h4>最小集</h4>
          <p>{project.mvp_scope || "—"}</p>
          <h4>止损说明</h4>
          <p>{project.stoploss_note || "—"}</p>
        </section>
      </div>

      <MilestoneList projectId={id} milestones={project.milestones || []} onChange={loadAll} />

      {activeSession && (
        <SessionRunner session={activeSession} onChange={loadAll} />
      )}

      <section>
        <h4>历史会话 ({sessions.length})</h4>
        <div className="session-history">
          {sessions.map((s) => (
            <div key={s.id} className={`session-item ${s.status}`}>
              <span>#{s.id}</span>
              <span>{s.status === "ended" ? "已结束" : s.status === "paused" ? "暂停中" : "运行中"}</span>
              <span className="muted">{s.started_at}</span>
              {s.consumed_tomato > 0 && <span>🍅 {fmtTomato(s.consumed_tomato)}</span>}
              {s.note && <span className="muted">· {s.note.slice(0, 30)}</span>}
            </div>
          ))}
          {sessions.length === 0 && <div className="muted">暂无会话</div>}
        </div>
      </section>
    </div>
  );
}

function fmtMin(m: number): string {
  if (m < 60) return `${m}分`;
  const h = Math.floor(m / 60);
  const r = m % 60;
  return r ? `${h}时${r}分` : `${h}时`;
}

// 番茄数显示，最多 2 位小数，去掉尾随 0（如 0.5 / 1 / 0.1）
function fmtTomato(t: number): string {
  const rounded = Math.round(t * 100) / 100;
  return String(rounded);
}

