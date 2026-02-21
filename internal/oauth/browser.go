package oauth

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
)

// OpenBrowser opens the given URL in the default browser.
// If opening fails, it prints a fallback message with the URL.
func OpenBrowser(authURL string) error {
	cmd, fallback := getBrowserCommand(runtime.GOOS, authURL)

	if fallback {
		fmt.Printf("\nUnable to open browser automatically.\n")
		fmt.Printf("Please open this URL in your browser:\n\n%s\n\n", authURL)
		return nil
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("\nFailed to open browser: %v\n", err)
		fmt.Printf("Please open this URL in your browser:\n\n%s\n\n", authURL)
		return err
	}

	return nil
}

// getBrowserCommand returns the appropriate command to open a URL in the default browser
// for the given operating system. The bool return value indicates whether this is a fallback
// (true means the OS is unsupported and the caller should print the URL instead).
func getBrowserCommand(goos, url string) (*exec.Cmd, bool) {
	var cmd *exec.Cmd

	switch goos {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		// Return a dummy command that will fail, and set fallback to true
		return exec.Command(""), true
	}

	return cmd, false
}

// buildAuthURL constructs the OAuth authorization URL with the required parameters.
func buildAuthURL(service, clientID, redirectURI, state string) string {
	baseURL := getServiceAuthURL(service)

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("scope", "tasks:read tasks:write")

	return baseURL + "?" + params.Encode()
}

// getServiceAuthURL returns the OAuth authorization URL for the given service.
// Supported services: dida365 (default), ticktick.
func getServiceAuthURL(service string) string {
	switch strings.ToLower(service) {
	case "ticktick":
		return "https://ticktick.com/oauth/authorize"
	case "dida365":
		return "https://dida365.com/oauth/authorize"
	default:
		// Default to dida365
		return "https://dida365.com/oauth/authorize"
	}
}

// getServiceTokenURL returns the OAuth token URL for the given service.
// Supported services: dida365 (default), ticktick.
func getServiceTokenURL(service string) string {
	switch strings.ToLower(service) {
	case "ticktick":
		return "https://ticktick.com/oauth/token"
	case "dida365":
		return "https://dida365.com/oauth/token"
	default:
		// Default to dida365
		return "https://dida365.com/oauth/token"
	}
}
