// API 客户端 + 类型

// 服务地址：默认相对路径 /api（Web 走 Vite 代理）。
// APK / 远程环境需配置绝对地址，存 localStorage。
// 约定：server_url 存 http://host:port（不带 /api），最终拼出 `${server}/api`。
const STORAGE_KEY = "server_url";

export function getBase(): string {
  // 显式配置的地址优先（APK / 远程）
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved) {
    return saved.replace(/\/$/, "") + "/api";
  }
  // 否则：原生环境（Capacitor）默认指向局域网后端；Web 用相对 /api 走代理
  const cap = (window as unknown as { Capacitor?: { isNativePlatform?: () => boolean } }).Capacitor;
  const isNative = typeof cap?.isNativePlatform === "function" && cap.isNativePlatform();
  return isNative ? "http://192.168.31.102:7800/api" : "/api";
}

export function setServerUrl(url: string) {
  localStorage.setItem(STORAGE_KEY, url.trim().replace(/\/$/, ""));
}

export function getServerUrl(): string {
  return localStorage.getItem(STORAGE_KEY) || "";
}

export interface Project {
  id: number;
  owner_id: string;
  title: string;
  necessity: string;
  mvp_scope: string;
  stoploss_note: string;
  budget_tomatoes: number;
  tomato_minutes: number | null;
  status: "active" | "completed" | "archived_stoploss";
  created_at: string;
  archived_at: string | null;
  milestones?: Milestone[];
  consumed_tomatoes: number;
  consumed_minutes: number;
}

export interface Milestone {
  id: number;
  project_id: number;
  title: string;
  done: boolean;
  order_idx: number;
  created_at: string;
}

export interface SessionEvent {
  id: number;
  session_id: number;
  type: "pause" | "resume" | "comment" | "voice_note";
  payload: string;
  voice_file: string | null;
  at: string;
  voice_note?: VoiceNote | null;
}

export interface VoiceNote {
  id: number;
  session_event_id: number;
  file_path: string;
  mime: string;
  duration_ms: number;
  text: string;
  created_at: string;
}

export interface Session {
  id: number;
  project_id: number;
  status: "running" | "paused" | "ended";
  started_at: string;
  ended_at: string | null;
  consumed_tomato: number;
  note: string;
  events?: SessionEvent[];
}

export interface Settings {
  tomato_minutes: number;
}

export interface Stats {
  budget_tomatoes: number;
  consumed_tomatoes: number;
  remaining_tomatoes: number;
  tomato_minutes: number;
  consumed_minutes: number;
  remaining_minutes: number;
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(getBase() + path, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;
  if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);
  return data as T;
}

export const api = {
  listProjects: (status?: string) =>
    req<Project[]>(`/projects${status ? `?status=${status}` : ""}`),
  getProject: (id: number) => req<Project>(`/projects/${id}`),
  createProject: (body: Partial<Project>) =>
    req<Project>(`/projects`, { method: "POST", body: JSON.stringify(body) }),
  updateProject: (id: number, body: Partial<Project>) =>
    req<Project>(`/projects/${id}`, { method: "PUT", body: JSON.stringify(body) }),
  setProjectStatus: (id: number, status: string) =>
    req<Project>(`/projects/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) }),
  deleteProject: (id: number) =>
    req<void>(`/projects/${id}`, { method: "DELETE" }),
  stats: (id: number) => req<Stats>(`/projects/${id}/stats`),

  createMilestone: (pid: number, title: string) =>
    req<Milestone>(`/projects/${pid}/milestones`, { method: "POST", body: JSON.stringify({ title }) }),
  updateMilestone: (mid: number, body: Partial<Milestone>) =>
    req<Milestone>(`/milestones/${mid}`, { method: "PATCH", body: JSON.stringify(body) }),
  deleteMilestone: (mid: number) =>
    req<void>(`/milestones/${mid}`, { method: "DELETE" }),

  listSessions: (pid: number) => req<Session[]>(`/projects/${pid}/sessions`),
  startSession: (pid: number) =>
    req<Session>(`/projects/${pid}/sessions`, { method: "POST" }),
  getSession: (sid: number) => req<Session>(`/sessions/${sid}`),
  pauseSession: (sid: number, comment?: string) =>
    req<Session>(`/sessions/${sid}/pause`, { method: "POST", body: JSON.stringify({ text: comment || "" }) }),
  resumeSession: (sid: number) =>
    req<Session>(`/sessions/${sid}/resume`, { method: "POST" }),
  commentSession: (sid: number, text: string) =>
    req<{ id: number; ok: boolean }>(`/sessions/${sid}/comment`, { method: "POST", body: JSON.stringify({ text }) }),
  endSession: (sid: number, note: string, consume: boolean) =>
    req<Session>(`/sessions/${sid}/end`, { method: "POST", body: JSON.stringify({ note, consume_tomato: consume }) }),

  getSettings: () => req<Settings>(`/settings`),
  updateSettings: (tomato_minutes: number) =>
    req<Settings>(`/settings`, { method: "PUT", body: JSON.stringify({ tomato_minutes }) }),
};

// 语音上传需 multipart，单独写
export async function uploadVoice(sid: number, file: Blob, text: string, durationMs: number) {
  const fd = new FormData();
  fd.append("file", file);
  fd.append("text", text);
  fd.append("duration_ms", String(durationMs));
  const res = await fetch(`${getBase()}/sessions/${sid}/voice`, { method: "POST", body: fd });
  if (!res.ok) {
    const t = await res.text();
    throw new Error(t || `HTTP ${res.status}`);
  }
  return res.json();
}

export const voiceUrl = (fn: string) => `${getBase()}/voice/${fn}`;
