package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackHandler(t *testing.T) {
	t.Run("successful callback with valid code and state", func(t *testing.T) {
		expectedState := "test-state-123"
		expectedCode := "auth-code-456"
		resultChan := make(chan callbackResult, 1)

		handler := newCallbackHandler(expectedState, resultChan)

		req := httptest.NewRequest(http.MethodGet, "/?code="+expectedCode+"&state="+expectedState, nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check HTTP response
		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Authentication successful") {
			t.Error("expected success message in response body")
		}

		// Check result channel
		select {
		case result := <-resultChan:
			if result.err != nil {
				t.Errorf("expected no error, got %v", result.err)
			}
			if result.code != expectedCode {
				t.Errorf("expected code %q, got %q", expectedCode, result.code)
			}
		default:
			t.Error("expected result in channel, got nothing")
		}
	})

	t.Run("callback with error", func(t *testing.T) {
		expectedState := "test-state-123"
		resultChan := make(chan callbackResult, 1)

		handler := newCallbackHandler(expectedState, resultChan)

		req := httptest.NewRequest(http.MethodGet, "/?error=access_denied&error_description=User+denied&state="+expectedState, nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check HTTP response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Authentication failed") {
			t.Error("expected error message in response body")
		}
		if !strings.Contains(body, "access_denied") {
			t.Error("expected error code in response body")
		}

		// Check result channel
		select {
		case result := <-resultChan:
			if result.err == nil {
				t.Error("expected error, got nil")
			}
			if !strings.Contains(result.err.Error(), "access_denied") {
				t.Errorf("expected error to contain 'access_denied', got %v", result.err)
			}
			if result.code != "" {
				t.Errorf("expected empty code, got %q", result.code)
			}
		default:
			t.Error("expected result in channel, got nothing")
		}
	})

	t.Run("callback with invalid state", func(t *testing.T) {
		expectedState := "test-state-123"
		wrongState := "wrong-state-456"
		resultChan := make(chan callbackResult, 1)

		handler := newCallbackHandler(expectedState, resultChan)

		req := httptest.NewRequest(http.MethodGet, "/?code=some-code&state="+wrongState, nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Check HTTP response
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Authentication failed") {
			t.Error("expected error message in response body")
		}
		bodyLower := strings.ToLower(body)
		if !strings.Contains(bodyLower, "state") || !strings.Contains(bodyLower, "invalid") {
			t.Error("expected state validation error in response body")
		}

		// Check result channel
		select {
		case result := <-resultChan:
			if result.err == nil {
				t.Error("expected error, got nil")
			}
			if !strings.Contains(strings.ToLower(result.err.Error()), "state") {
				t.Errorf("expected error to mention state, got %v", result.err)
			}
			if result.code != "" {
				t.Errorf("expected empty code, got %q", result.code)
			}
		default:
			t.Error("expected result in channel, got nothing")
		}
	})
}

func TestGenerateState(t *testing.T) {
	t.Run("generates 64-character hex string", func(t *testing.T) {
		state := generateState()

		if len(state) != 64 {
			t.Errorf("expected state length 64, got %d", len(state))
		}

		// Verify it's valid hex
		for _, c := range state {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("expected hex characters only, got %q", state)
				break
			}
		}
	})

	t.Run("generates unique states", func(t *testing.T) {
		state1 := generateState()
		state2 := generateState()

		if state1 == state2 {
			t.Error("expected unique states, got duplicates")
		}
	})
}
