package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"time"
)

// callbackResult represents the result of an OAuth callback
type callbackResult struct {
	code string
	err  error
}

// startCallbackServer starts a local HTTP server to receive OAuth callbacks.
// It returns the server instance, a channel that will receive the callback result,
// and any error that occurred during setup.
func startCallbackServer(port int, state string) (*http.Server, chan callbackResult, error) {
	resultChan := make(chan callbackResult, 1)
	handler := newCallbackHandler(state, resultChan)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", handler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			resultChan <- callbackResult{err: fmt.Errorf("server error: %w", err)}
		}
	}()

	return server, resultChan, nil
}

// newCallbackHandler creates an HTTP handler for OAuth callbacks.
// It validates the state parameter (CSRF protection), checks for errors,
// extracts the authorization code, and sends the result through the channel.
func newCallbackHandler(expectedState string, resultChan chan<- callbackResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate HTTP method (OAuth 2.0 spec requires GET)
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Method Not Allowed</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .error { color: #d32f2f; }
        h1 { color: #d32f2f; }
    </style>
</head>
<body>
    <h1>Method Not Allowed</h1>
    <p class="error">OAuth callbacks must use GET method.</p>
    <p>You can close this window and return to the terminal.</p>
</body>
</html>`)
			return
		}

		// Check for OAuth error response
		if errCode := r.URL.Query().Get("error"); errCode != "" {
			errDesc := r.URL.Query().Get("error_description")
			if errDesc == "" {
				errDesc = errCode
			}

			resultChan <- callbackResult{
				err: fmt.Errorf("OAuth error: %s (%s)", errCode, errDesc),
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .error { color: #d32f2f; }
        h1 { color: #d32f2f; }
    </style>
</head>
<body>
    <h1>Authentication failed</h1>
    <p class="error">Error: %s</p>
    <p>%s</p>
    <p>You can close this window and return to the terminal.</p>
</body>
</html>`, html.EscapeString(errCode), html.EscapeString(errDesc))
			return
		}

		// Validate state parameter (CSRF protection)
		state := r.URL.Query().Get("state")
		if state != expectedState {
			resultChan <- callbackResult{
				err: fmt.Errorf("invalid state parameter: possible CSRF attack"),
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .error { color: #d32f2f; }
        h1 { color: #d32f2f; }
    </style>
</head>
<body>
    <h1>Authentication failed</h1>
    <p class="error">Invalid state parameter detected.</p>
    <p>This may indicate a security issue. Please try again.</p>
    <p>You can close this window and return to the terminal.</p>
</body>
</html>`)
			return
		}

		// Extract authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			resultChan <- callbackResult{
				err: fmt.Errorf("no authorization code received"),
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .error { color: #d32f2f; }
        h1 { color: #d32f2f; }
    </style>
</head>
<body>
    <h1>Authentication failed</h1>
    <p class="error">No authorization code received.</p>
    <p>You can close this window and return to the terminal.</p>
</body>
</html>`)
			return
		}

		// Success! Send the code
		resultChan <- callbackResult{code: code}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body { font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .success { color: #388e3c; }
        h1 { color: #388e3c; }
    </style>
</head>
<body>
    <h1>Authentication successful!</h1>
    <p class="success">You have successfully authenticated with Dida365.</p>
    <p>You can close this window and return to the terminal.</p>
</body>
</html>`)
	}
}

// generateState generates a cryptographically secure random state parameter
// for CSRF protection. Returns a 64-character hex string (32 random bytes).
func generateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// shutdownServer performs a graceful shutdown of the HTTP server
// with a 5-second timeout.
func shutdownServer(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
