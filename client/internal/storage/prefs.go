package storage

import "fyne.io/fyne/v2"

const (
	prefServerURL  = "server_url"
	prefRememberMe = "remember_me"
	prefAuthToken  = "auth_token"
)

// Prefs wraps fyne application preferences for messenger settings.
type Prefs struct {
	p fyne.Preferences
}

// NewPrefs creates a preferences accessor for the given app.
func NewPrefs(app fyne.App) *Prefs {
	return &Prefs{p: app.Preferences()}
}

// ServerURL returns the saved server URL or empty string if unset.
func (p *Prefs) ServerURL() string {
	return p.p.String(prefServerURL)
}

// SetServerURL persists the server URL.
func (p *Prefs) SetServerURL(url string) {
	p.p.SetString(prefServerURL, url)
}

// RememberMe reports whether the user opted in to stay signed in.
func (p *Prefs) RememberMe() bool {
	return p.p.Bool(prefRememberMe)
}

// SetRememberMe persists the remember-me flag.
func (p *Prefs) SetRememberMe(remember bool) {
	p.p.SetBool(prefRememberMe, remember)
}

// AuthToken returns the saved JWT or empty string.
func (p *Prefs) AuthToken() string {
	return p.p.String(prefAuthToken)
}

// SetAuthToken persists the JWT for auto-login.
func (p *Prefs) SetAuthToken(token string) {
	p.p.SetString(prefAuthToken, token)
}

// ClearAuthSession removes saved credentials.
func (p *Prefs) ClearAuthSession() {
	p.SetRememberMe(false)
	p.SetAuthToken("")
}
