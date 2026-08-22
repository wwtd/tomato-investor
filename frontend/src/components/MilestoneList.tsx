import { useState } from "react";
import { api, type Milestone } from "../api";

export function MilestoneList({
  projectId,
  milestones,
  onChange,
}: {
  projectId: number;
  milestones: Milestone[];
  onChange: () => void;
}) {
  const [newTitle, setNewTitle] = useState("");
  const [adding, setAdding] = useState(false);

  async function add() {
    if (!newTitle.trim()) return;
    setAdding(true);
    try {
      await api.createMilestone(projectId, newTitle.trim());
      setNewTitle("");
      onChange();
    } finally {
      setAdding(false);
    }
  }

  async function toggle(m: Milestone) {
    await api.updateMilestone(m.id, { done: !m.done });
    onChange();
  }

  async function del(m: Milestone) {
    if (!confirm("删除里程碑？")) return;
    await api.deleteMilestone(m.id);
    onChange();
  }

  return (
    <section className="milestones">
      <h4>里程碑</h4>
      <div className="milestone-list">
        {milestones.map((m) => (
          <div key={m.id} className="milestone-item">
            <input type="checkbox" checked={m.done} onChange={() => toggle(m)} />
            <span className={m.done ? "done" : ""}>{m.title}</span>
            <button className="icon-btn" onClick={() => del(m)}>✕</button>
          </div>
        ))}
        {milestones.length === 0 && <div className="muted">暂无里程碑</div>}
      </div>
      <div className="milestone-add">
        <input
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          placeholder="新里程碑…"
          onKeyDown={(e) => e.key === "Enter" && add()}
        />
        <button onClick={add} disabled={adding || !newTitle.trim()}>
          添加
        </button>
      </div>
    </section>
  );
}
