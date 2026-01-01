package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// CallbackResult holds the result of an OAuth callback.
type CallbackResult struct {
	Code  string
	State string
	Error string
	Desc  string
}

// CallbackServer handles the OAuth callback from the browser.
type CallbackServer struct {
	port      int
	server    *http.Server
	resultCh  chan CallbackResult
	cancelCh  chan struct{}
	mu        sync.Mutex
	running   bool
	listener  net.Listener
}

// NewCallbackServer creates a new callback server.
// If port is 0, a random available port will be used.
func NewCallbackServer(port int) *CallbackServer {
	return &CallbackServer{
		port:     port,
		resultCh: make(chan CallbackResult, 1),
		cancelCh: make(chan struct{}),
	}
}

// Start starts the callback server and returns the actual port.
func (s *CallbackServer) Start() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return 0, errors.New("server already running")
	}

	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener

	// Get the actual port (useful when port was 0)
	actualPort := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	s.server = &http.Server{
		Handler: mux,
	}

	s.running = true

	go func() {
		_ = s.server.Serve(listener)
	}()

	return actualPort, nil
}

// handleCallback handles the OAuth callback request.
func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	result := CallbackResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Error: query.Get("error"),
		Desc:  query.Get("error_description"),
	}

	// Send result
	select {
	case s.resultCh <- result:
	default:
	}

	// Display message to user
	if result.Error != "" {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body>
<h1>Authentication Failed</h1>
<p>Error: %s</p>
<p>%s</p>
<p>You can close this window.</p>
</body>
</html>`, result.Error, result.Desc)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Authentication Successful</title></head>
<body>
<h1>Authentication Successful</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>`)
}

// WaitForCode waits for the OAuth callback and returns the code and state.
func (s *CallbackServer) WaitForCode(ctx context.Context) (code, state string, err error) {
	select {
	case <-ctx.Done():
		return "", "", ctx.Err()
	case <-s.cancelCh:
		return "", "", errors.New("authentication cancelled")
	case result := <-s.resultCh:
		if result.Error != "" {
			return "", "", fmt.Errorf("OAuth error: %s - %s", result.Error, result.Desc)
		}
		return result.Code, result.State, nil
	}
}

// Cancel cancels the wait for callback.
func (s *CallbackServer) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case s.cancelCh <- struct{}{}:
	default:
	}

	s.stopServer()
}

// Stop stops the callback server.
func (s *CallbackServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stopServer()
}

func (s *CallbackServer) stopServer() {
	if !s.running {
		return
	}

	s.running = false
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		_ = s.server.Shutdown(ctx)
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
}

// CallbackHandler is an http.Handler for the OAuth callback.
// This is exposed for testing.
type CallbackHandler struct {
	resultCh chan CallbackResult
}

// NewCallbackHandler creates a new callback handler.
func NewCallbackHandler() *CallbackHandler {
	return &CallbackHandler{
		resultCh: make(chan CallbackResult, 1),
	}
}

// ServeHTTP handles the callback request.
func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	result := CallbackResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Error: query.Get("error"),
		Desc:  query.Get("error_description"),
	}

	select {
	case h.resultCh <- result:
	default:
	}

	if result.Error != "" {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Authentication failed. You can close this window.")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "Authentication successful. You can close this window.")
}
