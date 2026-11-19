interface Me {
  sub: string;
  provider: string;
  email?: string;
  name?: string;
}

function initials(name?: string, email?: string): string {
  if (name) return name.charAt(0).toUpperCase();
  if (email) return email.charAt(0).toUpperCase();
  return "?";
}

const FEATURES = [
  "Access token (15m) + rotating refresh token (30d, httpOnly cookie)",
  "Rate limiting on login / register / OTP endpoints",
  "Same email across providers links to one account",
  "Generic login error — no user enumeration",
];

export function Dashboard({ me, onLogout }: { me: Me; onLogout: () => void }) {
  const mark = initials(me.name, me.email);

  return (
    <div className="dash">
      <nav className="dash-nav">
        <div className="auth-brand">
          <div className="auth-brand-mark">⚡</div>
          <span className="auth-brand-name">oAuth</span>
        </div>
        <div className="dash-nav-right">
          <span className="badge">
            <span className="badge-dot" />
            Session active
          </span>
          <div className="avatar avatar-sm">{mark}</div>
          <button className="btn" onClick={onLogout} style={{ width: "auto", padding: "8px 16px" }}>
            Chiqish
          </button>
        </div>
      </nav>

      <main className="dash-body">
        <section className="dash-hero">
          <div className="avatar">{mark}</div>
          <div>
            <h1>Xush kelibsiz{me.name ? `, ${me.name}` : ""}!</h1>
            <p>Siz muvaffaqiyatli autentifikatsiyadan o'tdingiz.</p>
          </div>
        </section>

        <div className="stat-grid">
          <div className="stat-card">
            <div className="stat-label">Provider</div>
            <div className="stat-value">{me.provider}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">Email</div>
            <div className="stat-value">{me.email ?? "—"}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">User ID</div>
            <div className="stat-value" style={{ fontSize: 12.5, fontFamily: "monospace" }}>
              {me.sub}
            </div>
          </div>
        </div>

        <div className="panel">
          <h3>Bu sessiyani himoya qiladigan narsalar</h3>
          <ul className="feature-list">
            {FEATURES.map((f) => (
              <li key={f}>{f}</li>
            ))}
          </ul>
        </div>
      </main>
    </div>
  );
}
