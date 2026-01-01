package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
)

func TestCallbackServer_WaitForCode(t *testing.T) {
	server := auth.NewCallbackServer(0) // Use random available port

	// Start server
	port, err := server.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	if port <= 0 {
		t.Errorf("Start() returned invalid port: %d", port)
	}

	// Simulate callback in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		resp, err := http.Get("http://localhost:" + itoa(port) + "/callback?code=test-auth-code&state=test-state")
		if err != nil {
			t.Errorf("callback request failed: %v", err)
			return
		}
		resp.Body.Close()
	}()

	// Wait for code
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code, state, err := server.WaitForCode(ctx)
	if err != nil {
		t.Errorf("WaitForCode() error = %v", err)
	}
	if code != "test-auth-code" {
		t.Errorf("WaitForCode() code = %q, want %q", code, "test-auth-code")
	}
	if state != "test-state" {
		t.Errorf("WaitForCode() state = %q, want %q", state, "test-state")
	}
}

func TestCallbackServer_WaitForCode_Timeout(t *testing.T) {
	server := auth.NewCallbackServer(0)

	_, err := server.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	// Wait with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, _, err = server.WaitForCode(ctx)
	if err == nil {
		t.Error("WaitForCode() expected timeout error, got nil")
	}
}

func TestCallbackServer_WaitForCode_Error(t *testing.T) {
	server := auth.NewCallbackServer(0)

	port, err := server.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer server.Stop()

	// Simulate error callback in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		resp, err := http.Get("http://localhost:" + itoa(port) + "/callback?error=access_denied&error_description=User+denied+access")
		if err != nil {
			t.Errorf("callback request failed: %v", err)
			return
		}
		resp.Body.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err = server.WaitForCode(ctx)
	if err == nil {
		t.Error("WaitForCode() expected error for access_denied, got nil")
	}
}

func TestCallbackServer_Cancel(t *testing.T) {
	server := auth.NewCallbackServer(0)

	_, err := server.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Cancel after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		server.Cancel()
	}()

	ctx := context.Background()
	_, _, err = server.WaitForCode(ctx)
	if err == nil {
		t.Error("WaitForCode() expected cancellation error, got nil")
	}
}

func TestCallbackHandler_Success(t *testing.T) {
	handler := auth.NewCallbackHandler()

	req := httptest.NewRequest(http.MethodGet, "/callback?code=auth-code&state=state-value", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handler returned status %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("handler returned empty body")
	}
}

func TestCallbackHandler_Error(t *testing.T) {
	handler := auth.NewCallbackHandler()

	req := httptest.NewRequest(http.MethodGet, "/callback?error=access_denied", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should still return 200 to display message to user
	if w.Code != http.StatusOK {
		t.Errorf("handler returned status %d, want %d", w.Code, http.StatusOK)
	}
}

// itoa converts an integer to string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
