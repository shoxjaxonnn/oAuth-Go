const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:4000";

// Access token: short-lived (15m), kept in memory/localStorage, sent as a
// Bearer header. Refresh token: long-lived, httpOnly cookie set by the
// backend — this code never reads or writes it directly.
export function saveToken(token: string) {
  localStorage.setItem("token", token);
}

export function getToken(): string | null {
  return localStorage.getItem("token");
}

export function clearToken() {
  localStorage.removeItem("token");
}

async function refreshAccessToken(): Promise<string | null> {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) return null;
  const { token } = (await res.json()) as { token: string };
  saveToken(token);
  return token;
}

async function authedFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const run = (token: string | null) =>
    fetch(`${API_BASE}${path}`, {
      ...init,
      credentials: "include",
      headers: {
        ...(init.headers ?? {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });

  let res = await run(getToken());
  if (res.status === 401) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      res = await run(newToken);
    } else {
      clearToken();
    }
  }
  return res;
}

export async function apiGet<T>(path: string): Promise<T> {
  const res = await authedFetch(path);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const res = await authedFetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error ?? "request failed");
  return data as T;
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE}/auth/logout`, { method: "POST", credentials: "include" });
  clearToken();
}

export { API_BASE, refreshAccessToken };
