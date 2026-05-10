import { useState } from "react";
import "./App.css";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8081";

function App() {
  const [message, setMessage] = useState("");
  const [reply, setReply] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function sendMessage(e: React.FormEvent) {
    e.preventDefault();

    if (!message.trim()) return;

    setLoading(true);
    setError("");
    setReply("");

    try {
      const res = await fetch(`${API_BASE_URL}/api/chat`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          message,
        }),
      });

      if (!res.ok) {
        throw new Error("API request failed");
      }

      const data = await res.json();
      setReply(data.reply);
    } catch (err) {
      console.error(err);
      setError("エラーが発生しました。OllamaまたはGoサーバーが起動しているか確認してください。");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="container">
      <h1>Ollama + Qwen3:4B Chat</h1>

      <form onSubmit={sendMessage} className="chat-form">
        <textarea
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="Qwenに聞きたいことを入力してください"
          rows={5}
        />

        <button type="submit" disabled={loading}>
          {loading ? "送信中..." : "送信"}
        </button>
      </form>

      {error && <div className="error">{error}</div>}

      {reply && (
        <section className="reply-box">
          <h2>回答</h2>
          <p>{reply}</p>
        </section>
      )}
    </main>
  );
}

export default App;