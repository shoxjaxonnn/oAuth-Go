package server_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/config"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/db"
	"github.com/shoxjaxonnn/oAuth-Go/backend/internal/server"
)

// fakeTransport stands in for every outbound call this backend makes
// (Google/Facebook/Telegram/Firebase), so the test suite never touches the
// real network. It always answers 200 with an empty JSON object.
type fakeTransport struct{}

func (fakeTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
		Header:     make(http.Header),
	}, nil
}

func newTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	cfg := &config.Config{
		Port:        "0",
		JWTSecret:   "test-secret",
		FrontendURL: "http://localhost:5173",
	}

	client := &http.Client{Transport: fakeTransport{}}
	ts := httptest.NewServer(server.New(cfg, database, client))
	t.Cleanup(ts.Close)
	return ts, database
}

func postJSON(t *testing.T, url string, body map[string]any) *http.Response {
	t.Helper()
	payload, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestHealth(t *testing.T) {
	ts, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMe_WithoutTokenIsUnauthorized(t *testing.T) {
	ts, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/auth/me")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMe_WithGarbageTokenIsUnauthorized(t *testing.T) {
	ts, _ := newTestServer(t)

	req, _ := http.NewRequest("GET", ts.URL+"/auth/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestFirebaseRegister_RequiresEmailAndPassword(t *testing.T) {
	ts, _ := newTestServer(t)

	resp := postJSON(t, ts.URL+"/auth/firebase/register", map[string]any{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestTelegramVerifyOTP_RequiresChatIDAndCode(t *testing.T) {
	ts, _ := newTestServer(t)

	resp := postJSON(t, ts.URL+"/auth/telegram/verify-otp", map[string]any{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestTelegramVerifyOTP_RejectsWrongCode(t *testing.T) {
	ts, _ := newTestServer(t)

	resp := postJSON(t, ts.URL+"/auth/telegram/verify-otp", map[string]any{
		"chatId": "never-sent", "code": "000000",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_WithoutCookieIsUnauthorized(t *testing.T) {
	ts, _ := newTestServer(t)

	resp, err := http.Post(ts.URL+"/auth/refresh", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogout_AlwaysSucceeds(t *testing.T) {
	ts, _ := newTestServer(t)

	resp, err := http.Post(ts.URL+"/auth/logout", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDocs_ServesSwaggerUIAndSpec(t *testing.T) {
	ts, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/docs")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /docs, got %d", resp.StatusCode)
	}

	specResp, err := http.Get(ts.URL + "/docs/openapi.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer specResp.Body.Close()
	if specResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /docs/openapi.json, got %d", specResp.StatusCode)
	}
}

func TestRateLimiting_BlocksRepeatedOTPRequests(t *testing.T) {
	ts, _ := newTestServer(t)

	var got429 bool
	for i := 0; i < 6; i++ {
		resp := postJSON(t, ts.URL+"/auth/telegram/send-otp", map[string]any{"chatId": "rate-limit-test"})
		if resp.StatusCode == http.StatusTooManyRequests {
			got429 = true
		}
		resp.Body.Close()
	}

	if !got429 {
		t.Fatal("expected at least one request to be rate-limited with 429")
	}
}
