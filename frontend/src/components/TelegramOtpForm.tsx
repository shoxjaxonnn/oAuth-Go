import { useState } from "react";
import { apiPost, saveToken } from "../api";

const TELEGRAM_BOT_USERNAME = import.meta.env.VITE_TELEGRAM_BOT_USERNAME ?? "";
const botLink = TELEGRAM_BOT_USERNAME ? `https://t.me/${TELEGRAM_BOT_USERNAME}` : null;

export function TelegramOtpForm({ onLoggedIn }: { onLoggedIn: () => void }) {
  const [chatId, setChatId] = useState("");
  const [code, setCode] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  async function sendOtp() {
    setError("");
    try {
      await apiPost("/auth/telegram/send-otp", { chatId });
      setSent(true);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  async function verifyOtp() {
    setError("");
    try {
      const { token } = await apiPost<{ token: string }>("/auth/telegram/verify-otp", {
        chatId,
        code,
      });
      saveToken(token);
      onLoggedIn();
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <div>
      <p className="field-hint">
        Avval{" "}
        {botLink ? (
          <a href={botLink} target="_blank" rel="noopener noreferrer">
            botga
          </a>
        ) : (
          "botga"
        )}{" "}
        /start yozing, keyin chat ID'ingizni kiriting.
      </p>
      <input
        className="field"
        placeholder="Telegram chat ID"
        value={chatId}
        onChange={(e) => setChatId(e.target.value)}
      />
      {!sent ? (
        <button className="btn btn-primary" onClick={sendOtp}>
          Kod yubor
        </button>
      ) : (
        <>
          <input
            className="field"
            placeholder="Kod"
            value={code}
            onChange={(e) => setCode(e.target.value)}
          />
          <button className="btn btn-primary" onClick={verifyOtp}>
            Tasdiqla
          </button>
        </>
      )}
      {error && <p className="field-error">{error}</p>}
    </div>
  );
}
