import { useEffect, useState } from "react";
import { api, getServerUrl, setServerUrl, type Project, type Settings } from "../api";

export function ProjectList({
  onOpen,
  settings,
  onSettingsChange,
}: {
  onOpen: (id: number) => void;
  settings: Settings | null;
  onSettingsChange: (s: Settings) => void;
}) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [filter, setFilter] = useState<string>("");
  const [showCreate, setShowCreate] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [showServer, setShowServer] = useState(false);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  async function load() {
    setLoading(true);
    try {
      setProjects(await api.listProjects(filter || undefined));
    } catch (e: unknown) {
      setErr((e instanceof Error ? e.message : String(e)));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, [filter]);

  return (
    <div className="project-list">
      <div className="toolbar">
        <select value={filter} onChange={(e) => setFilter(e.target.value)}>
          <option value="">全部</option>
          <option value="active">进行中</option>
          <option value="completed">已完成</option>
          <option value="archived_stoploss">止损归档</option>
        </select>
        <button onClick={() => setShowSettings(true)}>⚙ 设置</button>
        <button onClick={() => setShowServer(true)}>🔗 服务地址</button>
        <button className="primary" onClick={() => setShowCreate(true)}>
          + 新项目
        </button>
      </div>

      {err && <div className="error">{err}</div>}
      {loading && <div className="muted">加载中…</div>}

      <div className="project-grid">
        {projects.map((p) => (
          <ProjectCard key={p.id} project={p} onClick={() => onOpen(p.id)} />
        ))}
        {!loading && projects.length === 0 && (
          <div className="muted">还没有项目，点击「新项目」开始。</div>
        )}
      </div>

      {showCreate && (
        <ProjectForm
          onClose={() => setShowCreate(false)}
          onSaved={() => {
            setShowCreate(false);
            load();
          }}
        />
      )}

      {showSettings && settings && (
        <SettingsModal
          settings={settings}
          onClose={() => setShowSettings(false)}
          onSaved={(s) => {
            onSettingsChange(s);
            setShowSettings(false);
          }}
        />
      )}

      {showServer && (
        <ServerModal
          onClose={() => setShowServer(false)}
          onSaved={() => {
            setShowServer(false);
            load();
          }}
        />
      )}
    </div>
  );
}

function ProjectCard({ project: p, onClick }: { project: Project; onClick: () => void }) {
  const [stats, setStats] = useState<{ consumed: number; budget: number } | null>(null);
  useEffect(() => {
    api.stats(p.id).then((s) => setStats({ consumed: s.consumed_tomatoes, budget: s.budget_tomatoes }));
  }, [p.id]);

  const statusMap: Record<string, { label: string; cls: string }> = {
    active: { label: "进行中", cls: "st-active" },
    completed: { label: "已完成", cls: "st-done" },
    archived_stoploss: { label: "止损", cls: "st-stop" },
  };
  const st = statusMap[p.status];
  return (
    <div className={`project-card ${st.cls}`} onClick={onClick}>
      <div className="card-top">
        <h3>{p.title}</h3>
        <span className={`badge ${st.cls}`}>{st.label}</span>
      </div>
      <div className="card-budget">
        {stats ? (
          <>
            🍅 {fmtTomato(stats.consumed)} / {stats.budget}
            <div className="progress">
              <div
                className="progress-bar"
                style={{ width: `${stats.budget ? (Math.min(stats.consumed / stats.budget, 1) * 100).toFixed(1) : 0}%` }}
              />
            </div>
          </>
        ) : (
          <span className="muted">…</span>
        )}
      </div>
      {p.necessity && <p className="card-nec">{p.necessity.slice(0, 60)}</p>}
    </div>
  );
}

function ProjectForm({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const [form, setForm] = useState({
    title: "",
    necessity: "",
    mvp_scope: "",
    stoploss_note: "",
    budget_tomatoes: 3,
  });
  const [err, setErr] = useState("");
  const [saving, setSaving] = useState(false);

  async function save() {
    if (!form.title.trim()) {
      setErr("标题必填");
      return;
    }
    setSaving(true);
    try {
      await api.createProject(form);
      onSaved();
    } catch (e: unknown) {
      setErr((e instanceof Error ? e.message : String(e)));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal title="新项目" onClose={onClose}>
      <Field label="标题">
        <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="项目名称" />
      </Field>
      <Field label="必要性（为什么做）">
        <textarea value={form.necessity} onChange={(e) => setForm({ ...form, necessity: e.target.value })} rows={3} />
      </Field>
      <Field label="最小集（MVP 范围）">
        <textarea value={form.mvp_scope} onChange={(e) => setForm({ ...form, mvp_scope: e.target.value })} rows={3} />
      </Field>
      <Field label="止损说明（什么情况归档）">
        <textarea value={form.stoploss_note} onChange={(e) => setForm({ ...form, stoploss_note: e.target.value })} rows={2} />
      </Field>
      <Field label="投资番茄数（预算）">
        <input
          type="number"
          min={1}
          value={form.budget_tomatoes}
          onChange={(e) => setForm({ ...form, budget_tomatoes: parseInt(e.target.value) || 1 })}
        />
      </Field>
      {err && <div className="error">{err}</div>}
      <div className="form-actions">
        <button onClick={onClose}>取消</button>
        <button className="primary" onClick={save} disabled={saving}>
          {saving ? "保存中…" : "创建"}
        </button>
      </div>
    </Modal>
  );
}

function SettingsModal({
  settings,
  onClose,
  onSaved,
}: {
  settings: Settings;
  onClose: () => void;
  onSaved: (s: Settings) => void;
}) {
  const [val, setVal] = useState(settings.tomato_minutes);
  const [err, setErr] = useState("");
  const [saving, setSaving] = useState(false);

  async function save() {
    if (val < 1) {
      setErr("必须 >= 1");
      return;
    }
    setSaving(true);
    try {
      const s = await api.updateSettings(val);
      onSaved(s);
    } catch (e: unknown) {
      setErr((e instanceof Error ? e.message : String(e)));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal title="设置" onClose={onClose}>
      <Field label="番茄时长（分钟）">
        <input type="number" min={1} value={val} onChange={(e) => setVal(parseInt(e.target.value) || 1)} />
        <span className="muted">默认 480 = 8 小时</span>
      </Field>
      {err && <div className="error">{err}</div>}
      <div className="form-actions">
        <button onClick={onClose}>取消</button>
        <button className="primary" onClick={save} disabled={saving}>
          {saving ? "保存中…" : "保存"}
        </button>
      </div>
    </Modal>
  );
}

function ServerModal({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const [url, setUrl] = useState(getServerUrl());
  const [err, setErr] = useState("");
  const [ok, setOk] = useState(false);
  const [testing, setTesting] = useState(false);

  async function test() {
    let candidate = url.trim().replace(/\/$/, "");
    if (!candidate) {
      setErr("请输入服务地址");
      return;
    }
    // Android http 需显式协议前缀
    if (!/^https?:\/\//i.test(candidate)) {
      candidate = "http://" + candidate;
    }
    setTesting(true);
    setErr("");
    setOk(false);
    try {
      // 10s 超时，避免卡住
      const ctrl = new AbortController();
      const timer = setTimeout(() => ctrl.abort(), 10000);
      // 刷新前临时尝试连接候选地址的后端 /api/settings
      const res = await fetch(`${candidate}/api/settings`, { method: "GET", signal: ctrl.signal });
      clearTimeout(timer);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      setOk(true);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      // 区分常见错误，便于定位
      if (msg.includes("abort") || msg.includes("timeout")) {
        setErr("连接超时（10s）。请确认手机能访问该地址、后端已启动、防火墙放行");
      } else if (msg.includes("Failed to fetch") || msg.includes("Network request failed")) {
        setErr("网络请求失败。若在 APK 内：可能是明文 http 被拦截，或地址格式不对（需 http:// 前缀）");
      } else {
        setErr(`连接失败: ${msg} — 请检查地址/后端是否启动`);
      }
    } finally {
      setTesting(false);
    }
  }

  function save() {
    setServerUrl(url);
    onSaved();
  }

  return (
    <Modal title="设置服务地址" onClose={onClose}>
      <Field label="后端服务地址（APK/远程用，如 http://192.168.31.102:7800）">
        <input
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="http://192.168.31.102:7800"
        />
        <span className="muted">留空 = 当前环境默认地址（Web 走代理，APK 用局域网后端）</span>
      </Field>
      {ok && <div className="ok">✅ 连接成功</div>}
      {err && <div className="error">{err}</div>}
      <div className="form-actions">
        <button onClick={test} disabled={testing}>
          {testing ? "测试中…" : "测试连接"}
        </button>
        <button onClick={onClose}>取消</button>
        <button className="primary" onClick={save}>
          保存
        </button>
      </div>
    </Modal>
  );
}

export function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: React.ReactNode }) {
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>{title}</h2>
        {children}
      </div>
    </div>
  );
}

export function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="field">
      <span className="field-label">{label}</span>
      {children}
    </label>
  );
}

// 番茄数显示，最多 2 位小数，去掉尾随 0（如 0.5 / 1 / 0.1）
function fmtTomato(t: number): string {
  const rounded = Math.round(t * 100) / 100;
  return String(rounded);
}
