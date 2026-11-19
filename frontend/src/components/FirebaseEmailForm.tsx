import { useState } from "react";
import { apiPost, saveToken } from "../api";

export function FirebaseEmailForm({ onLoggedIn }: { onLoggedIn: () => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function submit(path: "register" | "login") {
    setError("");
    try {
      const { token } = await apiPost<{ token: string }>(`/auth/firebase/${path}`, {
        email,
        password,
      });
      saveToken(token);
      onLoggedIn();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <div>
      <input
        className="field"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <input
        className="field"
        placeholder="Parol"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <div className="btn-group">
        <button className="btn" onClick={() => submit("register")}>
          Ro'yxatdan o'tish
        </button>
        <button className="btn btn-primary" onClick={() => submit("login")}>
          Kirish
        </button>
      </div>
      {error && <p className="field-error">{error}</p>}
    </div>
  );
}
