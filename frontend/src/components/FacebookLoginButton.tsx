import { API_BASE } from "../api";

export function FacebookLoginButton() {
  async function handleClick() {
    const res = await fetch(`${API_BASE}/auth/facebook/url`);
    const { url } = await res.json();
    window.location.href = url;
  }

  return (
    <button className="btn btn-facebook" onClick={handleClick}>
      <svg className="btn-icon" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
        <path d="M22 12.06C22 6.5 17.52 2 12 2S2 6.5 2 12.06c0 5 3.66 9.15 8.44 9.94v-7.03H7.9v-2.91h2.54V9.85c0-2.5 1.49-3.89 3.77-3.89 1.1 0 2.24.2 2.24.2v2.46h-1.26c-1.24 0-1.63.77-1.63 1.56v1.88h2.78l-.44 2.91h-2.34V22c4.78-.79 8.44-4.94 8.44-9.94Z" />
      </svg>
      Facebook
    </button>
  );
}
