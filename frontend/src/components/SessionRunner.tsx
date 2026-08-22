import { useEffect, useRef, useState } from "react";
import { api, type Session, type SessionEvent, uploadVoice, voiceUrl } from "../api";

export function SessionRunner({
  session,
  onChange,
}: {
  session: Session;
  onChange: () => void;
}) {
  const [cur, setCur] = useState<Session>(session);
  const [comment, setComment] = useState("");
  const [endNote, setEndNote] = useState("");
  const [showEnd, setShowEnd] = useState(false);
  const [err, setErr] = useState("");
  const [elapsed, setElapsed] = useState(0);
  const refreshRef = useRef<() => void>(() => {});

  useEffect(() => {
    setCur(session);
  }, [session]);

  // 轮询当前会话状态，刷新运行时长 + 事件
  async function refresh() {
    try {
      const s = await api.getSession(session.id);
      setCur(s);
    } catch {
      /* ignore */
    }
  }
  refreshRef.current = refresh;

  // 运行中每 30 秒刷新一次
  useEffect(() => {
    if (cur.status === "ended") return;
    const t = setInterval(() => refreshRef.current(), 30000);
    return () => clearInterval(t);
  }, [cur.status]);

  // 本地计时器（秒）
  useEffect(() => {
    if (cur.status !== "running") return;
    const t = setInterval(() => setElapsed((e) => e + 1), 1000);
    return () => clearInterval(t);
  }, [cur.status]);

  async function action(path: "pause" | "resume", commentText?: string) {
    try {
      if (path === "pause") {
        const s = await api.pauseSession(cur.id, commentText);
        setCur(s);
        setComment("");
      } else {
        const s = await api.resumeSession(cur.id);
        setCur(s);
      }
      setElapsed(0);
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  async function sendComment() {
    if (!comment.trim()) return;
    try {
      await api.commentSession(cur.id, comment.trim());
      setComment("");
      refresh();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  async function endSession(consume: boolean) {
    try {
      await api.endSession(cur.id, endNote, consume);
      setShowEnd(false);
      onChange();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  const running = cur.status === "running";
  const events = cur.events || [];

  return (
    <section className={`session-runner ${cur.status}`}>
      <h4>番茄会话 #{cur.id}</h4>
      <div className="runner-status">
        <span className={`status-dot ${cur.status}`}>
          {cur.status === "running" ? "运行中" : cur.status === "paused" ? "已暂停" : "已结束"}
        </span>
        {running && <span className="timer">{fmtSec(elapsed)}</span>}
        <span className="muted">开始: {cur.started_at}</span>
      </div>

      {cur.status !== "ended" && (
        <div className="runner-controls">
          {running ? (
            <>
              <button className="warn" onClick={() => action("pause", comment)}>
                ⏸ 暂停{comment ? "（带备注）" : ""}
              </button>
            </>
          ) : (
            <button className="primary" onClick={() => action("resume")}>
              ▶ 恢复
            </button>
          )}
          <button className="danger" onClick={() => setShowEnd(true)}>
            ⏹ 结束
          </button>
        </div>
      )}

      {cur.status !== "ended" && (
        <div className="comment-box">
          <textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            placeholder="备注（暂停时记录情景，或运行中随手记）…"
            rows={2}
          />
          {running && (
            <button onClick={sendComment} disabled={!comment.trim()}>
              发送备注
            </button>
          )}
        </div>
      )}

      <VoiceRecorder sessionId={cur.id} disabled={cur.status === "ended"} onDone={refresh} />

      {showEnd && (
        <div className="end-panel">
          <h5>结束会话</h5>
          <textarea
            value={endNote}
            onChange={(e) => setEndNote(e.target.value)}
            placeholder="结束总结…"
            rows={2}
          />
          <label className="check-row">
            <input type="checkbox" defaultChecked /> 消耗 1 个番茄预算
          </label>
          <div className="form-actions">
            <button onClick={() => setShowEnd(false)}>取消</button>
            <button className="danger" onClick={() => endSession(true)}>
              结束并消耗番茄
            </button>
            <button onClick={() => endSession(false)}>结束不消耗</button>
          </div>
        </div>
      )}

      {events.length > 0 && (
        <div className="event-list">
          <h5>事件流</h5>
          {events.map((ev) => (
            <EventRow key={ev.id} ev={ev} />
          ))}
        </div>
      )}

      {err && <div className="error">{err}</div>}
    </section>
  );
}

function EventRow({ ev }: { ev: SessionEvent }) {
  const icon = { pause: "⏸", resume: "▶", comment: "💬", voice_note: "🎤" }[ev.type];
  let text = "";
  try {
    const p = JSON.parse(ev.payload);
    text = p.comment || "";
  } catch {
    /* ignore */
  }
  return (
    <div className="event-row">
      <span className="ev-icon">{icon}</span>
      <span className="muted ev-time">{ev.at}</span>
      {ev.type === "pause" && text && <span>暂停备注: {text}</span>}
      {ev.type === "comment" && <span>{text}</span>}
      {ev.type === "voice_note" && ev.voice_note && <VoicePlayback vn={ev.voice_note} />}
      {ev.type === "pause" && !text && <span>暂停</span>}
      {ev.type === "resume" && <span>恢复</span>}
    </div>
  );
}

function VoicePlayback({ vn }: { vn: NonNullable<SessionEvent["voice_note"]> }) {
  return (
    <span className="voice-playback">
      🎤 <audio controls src={voiceUrl(vn.file_path)} style={{ height: 28 }} />
      {vn.text && <span className="muted">· {vn.text}</span>}
    </span>
  );
}

function VoiceRecorder({
  sessionId,
  disabled,
  onDone,
}: {
  sessionId: number;
  disabled: boolean;
  onDone: () => void;
}) {
  const [recording, setRecording] = useState(false);
  const [text, setText] = useState("");
  const [err, setErr] = useState("");
  const [uploading, setUploading] = useState(false);
  const mediaRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const startRef = useRef<number>(0);
  const durRef = useRef<number>(0);

  async function start() {
    setErr("");
    try {
      if (!window.isSecureContext) {
        setErr("录音需要 HTTPS 或 localhost 环境，当前不是安全上下文（局域网 http 访问无法使用麦克风）");
        return;
      }
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        setErr("此浏览器不支持麦克风录制（缺少 navigator.mediaDevices）");
        return;
      }
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (typeof MediaRecorder === "undefined") {
        setErr("此浏览器不支持 MediaRecorder，无法录音");
        return;
      }
      const mr = new MediaRecorder(stream);
      chunksRef.current = [];
      mr.ondataavailable = (e) => {
        if (e.data.size > 0) chunksRef.current.push(e.data);
      };
      mr.onstop = async () => {
        stream.getTracks().forEach((t) => t.stop());
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        durRef.current = Date.now() - startRef.current;
        if (blob.size === 0) return;
        setUploading(true);
        try {
          await uploadVoice(sessionId, blob, text, durRef.current);
          setText("");
          onDone();
        } catch (e: unknown) {
          setErr(e instanceof Error ? e.message : String(e));
        } finally {
          setUploading(false);
        }
      };
      mr.start();
      startRef.current = Date.now();
      mediaRef.current = mr;
      setRecording(true);
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : String(e));
    }
  }

  function stop() {
    if (mediaRef.current && mediaRef.current.state !== "inactive") {
      mediaRef.current.stop();
    }
    setRecording(false);
  }

  return (
    <div className="voice-recorder">
      <input
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="语音备注文字（可选）…"
        disabled={disabled}
      />
      {!recording ? (
        <button onClick={start} disabled={disabled || uploading}>
          {uploading ? "上传中…" : "🎤 录音"}
        </button>
      ) : (
        <button className="danger" onClick={stop}>
          ⏹ 停止
        </button>
      )}
      {err && <div className="error">{err}</div>}
    </div>
  );
}

function fmtSec(s: number): string {
  const m = Math.floor(s / 60);
  const r = s % 60;
  return `${String(m).padStart(2, "0")}:${String(r).padStart(2, "0")}`;
}
