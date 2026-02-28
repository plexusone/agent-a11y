// Package auth provides authentication handling for accessibility audits.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	vibium "github.com/agentplexus/vibium-go"
)

// AuthType represents the authentication method.
type AuthType string

const (
	AuthTypeForm   AuthType = "form"
	AuthTypeOAuth  AuthType = "oauth"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeCookie AuthType = "cookie"
)

// Config contains authentication configuration.
type Config struct {
	// Type of authentication
	Type AuthType `yaml:"type" json:"type"`

	// Login URL (for form auth)
	LoginURL string `yaml:"loginUrl,omitempty" json:"loginUrl,omitempty"`

	// Credentials
	Credentials map[string]string `yaml:"credentials,omitempty" json:"credentials,omitempty"`

	// Form selectors (for form auth)
	Selectors *FormSelectors `yaml:"selectors,omitempty" json:"selectors,omitempty"`

	// Headers (for bearer/basic auth)
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`

	// Cookies (for cookie auth)
	Cookies map[string]string `yaml:"cookies,omitempty" json:"cookies,omitempty"`

	// Success indicator (selector that indicates successful login)
	SuccessIndicator string `yaml:"successIndicator,omitempty" json:"successIndicator,omitempty"`

	// Timeout for login
	Timeout time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// FormSelectors contains CSS selectors for form-based authentication.
type FormSelectors struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Submit   string `yaml:"submit" json:"submit"`
	MFA      string `yaml:"mfa,omitempty" json:"mfa,omitempty"`
}

// Handler handles authentication for audits.
type Handler struct {
	vibe   *vibium.Vibe
	logger *slog.Logger
	config *Config
}

// NewHandler creates a new authentication handler.
func NewHandler(vibe *vibium.Vibe, logger *slog.Logger, config *Config) *Handler {
	return &Handler{
		vibe:   vibe,
		logger: logger,
		config: config,
	}
}

// Authenticate performs authentication based on the configured method.
func (h *Handler) Authenticate(ctx context.Context) error {
	if h.config == nil {
		return nil // No authentication configured
	}

	switch h.config.Type {
	case AuthTypeForm:
		return h.authenticateForm(ctx)
	case AuthTypeBasic, AuthTypeBearer:
		return h.authenticateHeaders(ctx)
	case AuthTypeCookie:
		return h.authenticateCookies(ctx)
	case AuthTypeOAuth:
		return h.authenticateOAuth(ctx)
	default:
		return fmt.Errorf("unsupported auth type: %s", h.config.Type)
	}
}

// authenticateForm performs form-based authentication.
func (h *Handler) authenticateForm(ctx context.Context) error {
	h.logger.Info("performing form authentication", "loginUrl", h.config.LoginURL)

	// Navigate to login page
	if err := h.vibe.Go(ctx, h.config.LoginURL); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Wait for login form
	if h.config.Selectors == nil {
		return fmt.Errorf("form selectors not configured")
	}

	// Wait for username field using Find with timeout
	_, err := h.vibe.Find(ctx, h.config.Selectors.Username, &vibium.FindOptions{
		Timeout: h.config.Timeout,
	})
	if err != nil {
		return fmt.Errorf("username field not found: %w", err)
	}

	// Fill username
	username, ok := h.config.Credentials["username"]
	if !ok {
		return fmt.Errorf("username credential not provided")
	}

	usernameEl, err := h.vibe.Find(ctx, h.config.Selectors.Username, nil)
	if err != nil {
		return fmt.Errorf("failed to find username field: %w", err)
	}
	if err := usernameEl.Fill(ctx, username, nil); err != nil {
		return fmt.Errorf("failed to fill username: %w", err)
	}

	// Fill password
	password, ok := h.config.Credentials["password"]
	if !ok {
		return fmt.Errorf("password credential not provided")
	}

	passwordEl, err := h.vibe.Find(ctx, h.config.Selectors.Password, nil)
	if err != nil {
		return fmt.Errorf("failed to find password field: %w", err)
	}
	if err := passwordEl.Fill(ctx, password, nil); err != nil {
		return fmt.Errorf("failed to fill password: %w", err)
	}

	// Click submit
	submitEl, err := h.vibe.Find(ctx, h.config.Selectors.Submit, nil)
	if err != nil {
		return fmt.Errorf("failed to find submit button: %w", err)
	}
	if err := submitEl.Click(ctx, nil); err != nil {
		return fmt.Errorf("failed to click submit: %w", err)
	}

	// Wait for success indicator
	if h.config.SuccessIndicator != "" {
		_, err := h.vibe.Find(ctx, h.config.SuccessIndicator, &vibium.FindOptions{
			Timeout: h.config.Timeout,
		})
		if err != nil {
			return fmt.Errorf("login failed: success indicator not found: %w", err)
		}
	} else {
		// Wait for navigation
		if err := h.vibe.WaitForLoad(ctx, "networkidle", h.config.Timeout); err != nil {
			return fmt.Errorf("login failed: page did not load: %w", err)
		}
	}

	h.logger.Info("form authentication successful")
	return nil
}

// authenticateHeaders sets authentication headers.
func (h *Handler) authenticateHeaders(ctx context.Context) error {
	h.logger.Info("setting authentication headers")

	headers := make(map[string]string)

	switch h.config.Type {
	case AuthTypeBearer:
		token, ok := h.config.Credentials["token"]
		if !ok {
			return fmt.Errorf("bearer token not provided")
		}
		headers["Authorization"] = "Bearer " + token

	case AuthTypeBasic:
		username, ok := h.config.Credentials["username"]
		if !ok {
			return fmt.Errorf("username not provided")
		}
		password, ok := h.config.Credentials["password"]
		if !ok {
			return fmt.Errorf("password not provided")
		}
		// Basic auth is handled differently - need to set on page context
		headers["Authorization"] = "Basic " + basicAuth(username, password)
	}

	// Add any custom headers
	for k, v := range h.config.Headers {
		headers[k] = v
	}

	// Set extra HTTP headers
	return h.vibe.SetExtraHTTPHeaders(ctx, headers)
}

// authenticateCookies sets authentication cookies.
func (h *Handler) authenticateCookies(ctx context.Context) error {
	h.logger.Info("setting authentication cookies")

	// Set cookies via JavaScript
	for name, value := range h.config.Cookies {
		script := fmt.Sprintf(`document.cookie = "%s=%s; path=/"`, name, value)
		if _, err := h.vibe.Evaluate(ctx, script); err != nil {
			return fmt.Errorf("failed to set cookie %s: %w", name, err)
		}
	}

	return nil
}

// authenticateOAuth performs OAuth authentication.
func (h *Handler) authenticateOAuth(ctx context.Context) error {
	// OAuth typically requires interactive login
	// This is a placeholder - full implementation would handle OAuth flows
	h.logger.Warn("OAuth authentication not fully implemented")
	return fmt.Errorf("OAuth authentication requires interactive login")
}

// HandleMFA handles multi-factor authentication.
func (h *Handler) HandleMFA(ctx context.Context, mfaCode string) error {
	if h.config.Selectors == nil || h.config.Selectors.MFA == "" {
		return fmt.Errorf("MFA selector not configured")
	}

	// Find MFA input
	mfaEl, err := h.vibe.Find(ctx, h.config.Selectors.MFA, nil)
	if err != nil {
		return fmt.Errorf("MFA input not found: %w", err)
	}

	// Fill MFA code
	if err := mfaEl.Fill(ctx, mfaCode, nil); err != nil {
		return fmt.Errorf("failed to fill MFA code: %w", err)
	}

	// Submit (usually MFA forms auto-submit or have a submit button)
	// Wait for success indicator
	if h.config.SuccessIndicator != "" {
		_, err := h.vibe.Find(ctx, h.config.SuccessIndicator, &vibium.FindOptions{
			Timeout: h.config.Timeout,
		})
		if err != nil {
			return fmt.Errorf("MFA verification failed: %w", err)
		}
	}

	return nil
}

// IsAuthenticated checks if the session is authenticated.
func (h *Handler) IsAuthenticated(ctx context.Context) bool {
	if h.config == nil || h.config.SuccessIndicator == "" {
		return false
	}

	el, err := h.vibe.Find(ctx, h.config.SuccessIndicator, nil)
	return err == nil && el != nil
}

// Cookie represents a browser cookie.
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
}

// GetSessionCookies returns current session cookies.
func (h *Handler) GetSessionCookies(ctx context.Context) ([]Cookie, error) {
	script := `
		return document.cookie.split(';').map(c => {
			const parts = c.trim().split('=');
			return { name: parts[0], value: parts.slice(1).join('=') };
		});
	`
	result, err := h.vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	cookies := make([]Cookie, 0)
	if arr, ok := result.([]any); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				cookie := Cookie{
					Name:  getString(m, "name"),
					Value: getString(m, "value"),
				}
				cookies = append(cookies, cookie)
			}
		}
	}

	return cookies, nil
}

// SaveSession saves the current session for reuse.
func (h *Handler) SaveSession(ctx context.Context) (*Session, error) {
	cookies, err := h.GetSessionCookies(ctx)
	if err != nil {
		return nil, err
	}

	localStorage, err := h.getLocalStorage(ctx)
	if err != nil {
		h.logger.Warn("failed to get localStorage", "error", err)
	}

	sessionStorage, err := h.getSessionStorage(ctx)
	if err != nil {
		h.logger.Warn("failed to get sessionStorage", "error", err)
	}

	return &Session{
		Cookies:        cookies,
		LocalStorage:   localStorage,
		SessionStorage: sessionStorage,
	}, nil
}

// RestoreSession restores a saved session.
func (h *Handler) RestoreSession(ctx context.Context, session *Session) error {
	if session == nil {
		return nil
	}

	// Restore cookies via JavaScript
	for _, cookie := range session.Cookies {
		script := fmt.Sprintf(`document.cookie = "%s=%s; path=/"`, cookie.Name, cookie.Value)
		if _, err := h.vibe.Evaluate(ctx, script); err != nil {
			return fmt.Errorf("failed to restore cookie %s: %w", cookie.Name, err)
		}
	}

	// Restore localStorage
	if len(session.LocalStorage) > 0 {
		if err := h.setLocalStorage(ctx, session.LocalStorage); err != nil {
			h.logger.Warn("failed to restore localStorage", "error", err)
		}
	}

	// Restore sessionStorage
	if len(session.SessionStorage) > 0 {
		if err := h.setSessionStorage(ctx, session.SessionStorage); err != nil {
			h.logger.Warn("failed to restore sessionStorage", "error", err)
		}
	}

	return nil
}

// Session contains saved session data.
type Session struct {
	Cookies        []Cookie          `json:"cookies"`
	LocalStorage   map[string]string `json:"localStorage"`
	SessionStorage map[string]string `json:"sessionStorage"`
}

// getLocalStorage retrieves localStorage data.
func (h *Handler) getLocalStorage(ctx context.Context) (map[string]string, error) {
	script := `
		const data = {};
		for (let i = 0; i < localStorage.length; i++) {
			const key = localStorage.key(i);
			data[key] = localStorage.getItem(key);
		}
		return data;
	`

	result, err := h.vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	storage := make(map[string]string)
	if m, ok := result.(map[string]any); ok {
		for k, v := range m {
			if s, ok := v.(string); ok {
				storage[k] = s
			}
		}
	}

	return storage, nil
}

// setLocalStorage sets localStorage data.
func (h *Handler) setLocalStorage(ctx context.Context, data map[string]string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	script := fmt.Sprintf(`
		const data = %s;
		for (const [key, value] of Object.entries(data)) {
			localStorage.setItem(key, value);
		}
	`, string(dataJSON))
	_, err = h.vibe.Evaluate(ctx, script)
	return err
}

// getSessionStorage retrieves sessionStorage data.
func (h *Handler) getSessionStorage(ctx context.Context) (map[string]string, error) {
	script := `
		const data = {};
		for (let i = 0; i < sessionStorage.length; i++) {
			const key = sessionStorage.key(i);
			data[key] = sessionStorage.getItem(key);
		}
		return data;
	`

	result, err := h.vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	storage := make(map[string]string)
	if m, ok := result.(map[string]any); ok {
		for k, v := range m {
			if s, ok := v.(string); ok {
				storage[k] = s
			}
		}
	}

	return storage, nil
}

// setSessionStorage sets sessionStorage data.
func (h *Handler) setSessionStorage(ctx context.Context, data map[string]string) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	script := fmt.Sprintf(`
		const data = %s;
		for (const [key, value] of Object.entries(data)) {
			sessionStorage.setItem(key, value);
		}
	`, string(dataJSON))
	_, err = h.vibe.Evaluate(ctx, script)
	return err
}

// basicAuth encodes credentials for basic auth.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64Encode([]byte(auth))
}

// base64Encode encodes data to base64.
func base64Encode(data []byte) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	result := make([]byte, 0, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		var n uint32
		switch len(data) - i {
		case 1:
			n = uint32(data[i]) << 16
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3f], '=', '=')
		case 2:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3f], base64Table[(n>>6)&0x3f], '=')
		default:
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
			result = append(result, base64Table[n>>18], base64Table[(n>>12)&0x3f], base64Table[(n>>6)&0x3f], base64Table[n&0x3f])
		}
	}

	return string(result)
}

// getString safely extracts a string from a map.
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
