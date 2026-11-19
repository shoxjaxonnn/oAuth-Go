import { useEffect, useState } from "react";
import { apiGet, logout as apiLogout, saveToken } from "./api";
import { GoogleLoginButton } from "./components/GoogleLoginButton";
import { FacebookLoginButton } from "./components/FacebookLoginButton";
import { TelegramOtpForm } from "./components/TelegramOtpForm";
import { FirebaseEmailForm } from "./components/FirebaseEmailForm";

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

function initials(name?: string, email?: string): string {
  if (name) return name.charAt(0).toUpperCase();
  if (email) return email.charAt(0).toUpperCase();
  return "?";
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

  if (loading) return <div className="loading">Yuklanmoqda...</div>;

  if (me) {
    return (
      <div className="auth-page">
        <div className="auth-card profile-card">
          <div className="profile-avatar">{initials(me.name, me.email)}</div>
          <h1>Xush kelibsiz!</h1>
          <div className="profile-row">
            <span className="profile-label">Provider</span>
            <span className="profile-value">{me.provider}</span>
          </div>
          {me.email && (
            <div className="profile-row">
              <span className="profile-label">Email</span>
              <span className="profile-value">{me.email}</span>
            </div>
          )}
          {me.name && (
            <div className="profile-row">
              <span className="profile-label">Ism</span>
              <span className="profile-value">{me.name}</span>
            </div>
          )}
          <button className="btn btn-primary" style={{ marginTop: 20 }} onClick={logout}>
            Chiqish
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Hisobga kirish</h1>

        <div className="auth-section btn-group">
          <GoogleLoginButton />
          <FacebookLoginButton />
        </div>

        <hr className="auth-divider" />

        <div className="auth-section">
          <h2>Telegram OTP</h2>
          <TelegramOtpForm onLoggedIn={refreshMe} />
        </div>

        <hr className="auth-divider" />

        <div className="auth-section">
          <h2>Email / Parol</h2>
          <FirebaseEmailForm onLoggedIn={refreshMe} />
        </div>
      </div>
    </div>
  );
}
