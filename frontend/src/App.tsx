import { useEffect, useState } from "react";
import { apiGet, logout as apiLogout, saveToken } from "./api";
import { GoogleLoginButton } from "./components/GoogleLoginButton";
import { FacebookLoginButton } from "./components/FacebookLoginButton";
import { TelegramOtpForm } from "./components/TelegramOtpForm";
import { FirebaseEmailForm } from "./components/FirebaseEmailForm";
import { Dashboard } from "./components/Dashboard";

interface Me {
  sub: string;
  provider: string;
  email?: string;
  name?: string;
}

function consumeCallbackToken() {
  if (window.location.pathname !== "/auth/callback") return;
  const params = new URLSearchParams(window.location.search);
  const token = params.get("token");
  if (token) saveToken(token);
  window.history.replaceState({}, "", "/");
}

export default function App() {
  const [me, setMe] = useState<Me | null>(null);
  const [loading, setLoading] = useState(true);

  async function refreshMe() {
    try {
      // authedFetch tries a cookie-based silent refresh on 401, so this
      // also covers "no access token yet but the refresh cookie is valid".
      const { user } = await apiGet<{ user: Me }>("/auth/me");
      setMe(user);
    } catch {
      setMe(null);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    consumeCallbackToken();
    refreshMe();
  }, []);

  async function logout() {
    await apiLogout();
    setMe(null);
  }

  if (loading) {
    return (
      <div className="loading">
        <span className="spinner" /> Yuklanmoqda...
      </div>
    );
  }

  if (me) {
    return <Dashboard me={me} onLogout={logout} />;
  }

  return (
    <div className="auth-page">
      <div className="auth-brand">
        <div className="auth-brand-mark">⚡</div>
        <span className="auth-brand-name">oAuth</span>
      </div>

      <div className="auth-card">
        <h1>Hisobga kirish</h1>
        <p className="auth-subtitle">4 ta usuldan birini tanlang</p>

        <div className="auth-section btn-group">
          <GoogleLoginButton />
          <FacebookLoginButton />
        </div>

        <div className="auth-section">
          <p className="auth-section-label">Telegram OTP</p>
          <TelegramOtpForm onLoggedIn={refreshMe} />
        </div>

        <div className="auth-section">
          <p className="auth-section-label">Email / Parol</p>
          <FirebaseEmailForm onLoggedIn={refreshMe} />
        </div>
      </div>
    </div>
  );
}
