import { useEffect, useState } from "react";
import { api, type Settings } from "./api";
import { ProjectList } from "./components/ProjectList";
import { ProjectDetail } from "./components/ProjectDetail";
import "./App.css";

export default function App() {
  const [view, setView] = useState<"list" | "detail">("list");
  const [currentId, setCurrentId] = useState<number | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);

  useEffect(() => {
    api.getSettings().then(setSettings).catch(console.error);
  }, []);

  function openProject(id: number) {
    setCurrentId(id);
    setView("detail");
  }

  function back() {
    setView("list");
    setCurrentId(null);
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1 onClick={back} style={{ cursor: "pointer" }}>
          🍅 番茄投资人
        </h1>
        {settings && (
          <span className="tomato-config">
            1 番茄 = {settings.tomato_minutes} 分钟
          </span>
        )}
      </header>
      {view === "list" && (
        <ProjectList onOpen={openProject} settings={settings} onSettingsChange={setSettings} />
      )}
      {view === "detail" && currentId && (
        <ProjectDetail id={currentId} onBack={back} />
      )}
    </div>
  );
}
