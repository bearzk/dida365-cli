# OAuth2 Authentication Flow Design

**Date:** 2026-02-21
**Status:** Approved
**Author:** Claude Code

## Overview

Implement proper OAuth2 authentication flow for Dida365 CLI to replace the current manual `auth configure` command. The new flow will automatically handle browser-based authorization, token exchange, and refresh token management.

## Problem Statement

Current implementation requires users to manually obtain `access_token` through external means (browser, curl), which is:
- Not user-friendly
- Error-prone
- Not standard OAuth2 practice
- Doesn't support token refresh

## Goals

1. Implement standard OAuth2 authorization code flow
2. Support both TickTick and Dida365 services
3. Provide configurable redirect URI port
4. Store and manage refresh tokens
5. Maintain backward compatibility with existing configs
6. Keep minimal external dependencies (stdlib-first approach)

## Architecture

### New Package: `internal/oauth/`

**Files:**
- `oauth.go` - OAuth2 flow orchestration
- `server.go` - Local HTTP callback server
- `browser.go` - Cross-platform browser opening

### Modified Files

**`cmd/auth.go`**
- Remove `auth configure` command
- Add `auth login` command with OAuth flow
- Add `auth refresh` command for token refresh
- Update `auth status` to show token expiry and refresh capability

**`internal/config/config.go`**
- Add `RefreshToken string` field
- Add `TokenExpiry time.Time` field

## Design Decisions

### 1. Service Type Support
**Decision:** Support both TickTick and Dida365 services
**Rationale:** Minimal additional complexity, better user experience
**Implementation:** `--service` flag (values: `ticktick`, `dida365`, default: `dida365`)

### 2. Redirect URI Port
**Decision:** Configurable port via `--port` flag (default: 8080)
**Rationale:** Gives users flexibility if default port is busy
**Implementation:** `--port` flag, must match what's registered in developer portal

### 3. Token Refresh Strategy
**Decision:** Manual refresh via explicit `auth refresh` command
**Rationale:** Gives users control, simpler initial implementation, can add auto-refresh later
**Implementation:** Store refresh_token and expiry, require explicit command to refresh

### 4. Backward Compatibility
**Decision:** Replace `auth configure` entirely
**Rationale:** One clear authentication path, OAuth is the standard approach
**Implementation:** Existing configs continue to work, missing refresh_token handled gracefully

### 5. External Dependencies
**Decision:** Minimal dependencies - stdlib-first approach
**Rationale:** Keep binary small, maintain current project style, OAuth flow is straightforward
**Implementation:** Use `net/http`, `os/exec`, standard HTTP client

## Data Flow

### `auth login` Command Flow

```
User runs: dida365 auth login --client-id xxx --client-secret yyy --service dida365 --port 8080

1. Validate inputs (client-id, client-secret required)
2. Start callback server on http://localhost:8080/callback
3. Generate random state (32-byte hex for CSRF protection)
4. Build authorization URL:
   https://{service}.com/oauth/authorize?
     client_id=xxx&
     redirect_uri=http://localhost:{port}/callback&
     response_type=code&
     state={random_state}&
     scope=tasks:read tasks:write

5. Open browser → user sees authorization page
6. User authorizes → redirect to callback with ?code=AUTH_CODE&state={random_state}
7. Server validates state matches
8. Extract auth code from query parameter
9. Exchange code for tokens:
   POST https://{service}.com/oauth/token
   {
     "client_id": "xxx",
     "client_secret": "yyy",
     "code": "AUTH_CODE",
     "grant_type": "authorization_code",
     "redirect_uri": "http://localhost:{port}/callback"
   }

10. Receive: access_token, refresh_token, expires_in
11. Calculate expiry: time.Now().Add(expires_in * time.Second)
12. Save to config: all credentials + tokens + expiry
13. Output JSON:
    {
      "authenticated": true,
      "expires_at": "2026-02-21T14:30:00Z",
      "config_path": "~/.dida365/config.json"
    }
14. Shutdown callback server
```

### `auth refresh` Command Flow

```
User runs: dida365 auth refresh

1. Load config → get client_id, client_secret, refresh_token
2. Validate refresh token exists
3. Exchange refresh token:
   POST https://{service}.com/oauth/token
   {
     "client_id": "xxx",
     "client_secret": "yyy",
     "refresh_token": "REFRESH_TOKEN",
     "grant_type": "refresh_token"
   }

4. Receive new tokens: access_token, refresh_token (new), expires_in
5. Update config with new tokens and expiry
6. Output JSON:
   {
     "refreshed": true,
     "expires_at": "2026-02-21T16:30:00Z"
   }
```

## Components

### OAuth Flow Orchestrator (`internal/oauth/oauth.go`)

**Functions:**
- `StartFlow(clientID, clientSecret, redirectURI, service string) (*TokenResponse, error)`
- `RefreshToken(clientID, clientSecret, refreshToken, service string) (*TokenResponse, error)`

**Responsibilities:**
- Generate random state for CSRF protection
- Build authorization URL
- Start callback server
- Open browser
- Wait for callback with timeout (5 minutes)
- Exchange code/refresh_token for tokens
- Return token response

### Callback Server (`internal/oauth/server.go`)

**Implementation:**
- Single HTTP handler: `GET /callback`
- Validates state parameter
- Extracts auth code or error
- Shows HTML success/error page to user
- Signals completion via channel
- Auto-shutdown after receiving callback

**Error Handling:**
- Invalid state → show error page
- Missing code → show error page
- Authorization denied → extract error_description

### Browser Opener (`internal/oauth/browser.go`)

**Implementation:**
- Try platform-specific commands:
  - macOS: `open {url}`
  - Linux: `xdg-open {url}`
  - Windows: `cmd /c start {url}`
- Fallback: Print URL to terminal
- Non-blocking operation

### Token Exchange

**Authorization Code Exchange:**
```
POST https://{service}.com/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
code={AUTH_CODE}
client_id={CLIENT_ID}
client_secret={CLIENT_SECRET}
redirect_uri={REDIRECT_URI}
```

**Refresh Token Exchange:**
```
POST https://{service}.com/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
refresh_token={REFRESH_TOKEN}
client_id={CLIENT_ID}
client_secret={CLIENT_SECRET}
```

**Response:**
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 7200,
  "token_type": "Bearer"
}
```

## Error Handling

| Error | Exit Code | Message |
|-------|-----------|---------|
| Port already in use | 1 | "Port {port} is already in use. Specify different port with --port or free up port {port}" |
| Browser opening fails | N/A | Print URL, continue waiting for callback |
| User denies authorization | 2 | "Authorization denied by user" |
| State mismatch (CSRF) | 2 | "Invalid state parameter - possible CSRF attack" |
| Token exchange fails | 2 | "Failed to exchange code for token: {api_error}" |
| Callback timeout | 2 | "Authorization timeout - no response received within 5 minutes" |
| Config save fails | 1 | "Failed to save configuration: {error}" |
| No refresh token | 1 | "No refresh token found. Run 'dida365 auth login' first" |
| Refresh token expired | 2 | "Refresh token expired. Run 'dida365 auth login' to re-authenticate" |

## Config Schema Updates

**Before:**
```json
{
  "client_id": "xxx",
  "client_secret": "yyy",
  "access_token": "zzz",
  "base_url": "https://dida365.com"
}
```

**After:**
```json
{
  "client_id": "xxx",
  "client_secret": "yyy",
  "access_token": "zzz",
  "base_url": "https://dida365.com",
  "refresh_token": "www",
  "token_expiry": "2026-02-21T16:30:00Z"
}
```

**Backward Compatibility:**
- Old configs (missing refresh_token/token_expiry) remain valid
- `auth status` shows `"can_refresh": false` for old configs
- `auth refresh` errors gracefully for old configs

## Testing Strategy

### Unit Tests

**`internal/oauth/oauth_test.go`**
- State generation (randomness, length)
- URL building (correct parameters, encoding)
- Service URL mapping (ticktick vs dida365)

**`internal/oauth/browser_test.go`**
- Command selection for different platforms

**`internal/config/config_test.go`**
- Config with new fields marshals/unmarshals correctly
- Backward compatibility with old configs

### Integration Tests (Manual)

- Full login flow with real Dida365 credentials
- Full login flow with real TickTick credentials
- Token refresh flow
- Port conflict scenario
- Timeout scenario
- User denial scenario

### Not Tested (Low Value)

- Actual browser opening (platform-dependent)
- Real HTTP server lifecycle (stdlib-tested)

## Migration Path

### For Users

**Breaking Change:** `auth configure` command removed

**Impact:** Users must use `auth login` instead

**Existing Configs:**
- Continue to work (only adding fields)
- Missing refresh_token → `auth status` shows can't refresh
- Users can re-authenticate with `auth login` to get refresh capability

**README Updates:**
- Update "Getting Started" section
- Show `auth login` instead of `auth configure`
- Document `auth refresh` command
- Explain token expiry and refresh

## Implementation Estimates

| Component | Lines of Code | Complexity |
|-----------|---------------|------------|
| `internal/oauth/oauth.go` | ~150 | Medium |
| `internal/oauth/server.go` | ~100 | Low |
| `internal/oauth/browser.go` | ~50 | Low |
| `cmd/auth.go` modifications | ~200 | Medium |
| `internal/config/config.go` | ~10 | Low |
| Tests | ~150 | Low |
| **Total** | **~660** | **Medium** |

## Success Criteria

1. ✅ User can run `auth login` and authenticate via browser
2. ✅ Tokens automatically saved with expiry
3. ✅ User can refresh tokens with `auth refresh`
4. ✅ Both TickTick and Dida365 services supported
5. ✅ Configurable port for redirect URI
6. ✅ Clear error messages for all failure scenarios
7. ✅ Backward compatible with existing configs
8. ✅ All tests pass

## Future Enhancements (Out of Scope)

- Auto-refresh tokens before API calls
- Multiple profile support
- Token revocation command
- PKCE (Proof Key for Code Exchange) for additional security

## References

- [Dida365 API Documentation](https://developer.dida365.com/api)
- [TickTick API Documentation](https://developer.ticktick.com/docs)
- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [Reference Implementation](https://cyfine.github.io/TickTick-Dida365-API-Client/guides/authentication/)
