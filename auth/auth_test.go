package auth

import (
	"encoding/json"
	"testing"
)

func TestAuthTypeConstants(t *testing.T) {
	tests := []struct {
		authType AuthType
		expected string
	}{
		{AuthTypeForm, "form"},
		{AuthTypeOAuth, "oauth"},
		{AuthTypeBasic, "basic"},
		{AuthTypeBearer, "bearer"},
		{AuthTypeCookie, "cookie"},
	}

	for _, tt := range tests {
		if string(tt.authType) != tt.expected {
			t.Errorf("AuthType %v: got %q, want %q", tt.authType, string(tt.authType), tt.expected)
		}
	}
}

func TestConfigJSONSerialization(t *testing.T) {
	cfg := Config{
		Type:     AuthTypeForm,
		LoginURL: "https://example.com/login",
		Credentials: map[string]string{
			"username": "testuser",
			"password": "testpass",
		},
		Selectors: &FormSelectors{
			Username: "#username",
			Password: "#password",
			Submit:   "#submit",
			MFA:      "#mfa",
		},
		Headers: map[string]string{
			"X-Custom": "value",
		},
		Cookies: map[string]string{
			"session": "abc123",
		},
		SuccessIndicator: ".dashboard",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal Config: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Config: %v", err)
	}

	if decoded.Type != cfg.Type {
		t.Errorf("Type mismatch: got %q, want %q", decoded.Type, cfg.Type)
	}
	if decoded.LoginURL != cfg.LoginURL {
		t.Errorf("LoginURL mismatch")
	}
	if decoded.Credentials["username"] != "testuser" {
		t.Errorf("Credentials[username] mismatch")
	}
	if decoded.Selectors == nil {
		t.Fatal("Selectors should not be nil")
	}
	if decoded.Selectors.Username != "#username" {
		t.Errorf("Selectors.Username mismatch")
	}
	if decoded.SuccessIndicator != ".dashboard" {
		t.Errorf("SuccessIndicator mismatch")
	}
}

func TestFormSelectorsJSONSerialization(t *testing.T) {
	selectors := FormSelectors{
		Username: "#user",
		Password: "#pass",
		Submit:   "button[type=submit]",
		MFA:      "#otp",
	}

	data, err := json.Marshal(selectors)
	if err != nil {
		t.Fatalf("Failed to marshal FormSelectors: %v", err)
	}

	var decoded FormSelectors
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal FormSelectors: %v", err)
	}

	if decoded.Username != selectors.Username {
		t.Errorf("Username mismatch")
	}
	if decoded.Password != selectors.Password {
		t.Errorf("Password mismatch")
	}
	if decoded.Submit != selectors.Submit {
		t.Errorf("Submit mismatch")
	}
	if decoded.MFA != selectors.MFA {
		t.Errorf("MFA mismatch")
	}
}

func TestCookieJSONSerialization(t *testing.T) {
	cookie := Cookie{
		Name:     "session",
		Value:    "abc123xyz",
		Domain:   ".example.com",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}

	data, err := json.Marshal(cookie)
	if err != nil {
		t.Fatalf("Failed to marshal Cookie: %v", err)
	}

	var decoded Cookie
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Cookie: %v", err)
	}

	if decoded.Name != cookie.Name {
		t.Errorf("Name mismatch")
	}
	if decoded.Value != cookie.Value {
		t.Errorf("Value mismatch")
	}
	if decoded.Domain != cookie.Domain {
		t.Errorf("Domain mismatch")
	}
	if decoded.Secure != cookie.Secure {
		t.Errorf("Secure mismatch")
	}
	if decoded.HttpOnly != cookie.HttpOnly {
		t.Errorf("HttpOnly mismatch")
	}
}

func TestSessionJSONSerialization(t *testing.T) {
	session := Session{
		Cookies: []Cookie{
			{Name: "session", Value: "abc"},
			{Name: "token", Value: "xyz"},
		},
		LocalStorage: map[string]string{
			"user":  "testuser",
			"theme": "dark",
		},
		SessionStorage: map[string]string{
			"tab": "home",
		},
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal Session: %v", err)
	}

	var decoded Session
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Session: %v", err)
	}

	if len(decoded.Cookies) != 2 {
		t.Errorf("Cookies count: got %d, want 2", len(decoded.Cookies))
	}
	if decoded.LocalStorage["user"] != "testuser" {
		t.Errorf("LocalStorage[user] mismatch")
	}
	if decoded.SessionStorage["tab"] != "home" {
		t.Errorf("SessionStorage[tab] mismatch")
	}
}

func TestBasicAuth(t *testing.T) {
	// Test basic auth encoding
	result := basicAuth("user", "pass")

	// "user:pass" base64 encoded is "dXNlcjpwYXNz"
	expected := "dXNlcjpwYXNz"
	if result != expected {
		t.Errorf("basicAuth(user, pass): got %q, want %q", result, expected)
	}
}

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "YQ=="},
		{"ab", "YWI="},
		{"abc", "YWJj"},
		{"Hello, World!", "SGVsbG8sIFdvcmxkIQ=="},
		{"user:password", "dXNlcjpwYXNzd29yZA=="},
	}

	for _, tt := range tests {
		result := base64Encode([]byte(tt.input))
		if result != tt.expected {
			t.Errorf("base64Encode(%q): got %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{
		"name":   "John",
		"count":  42,
		"flag":   true,
		"nested": map[string]any{"key": "value"},
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"name", "John"},
		{"missing", ""},
		{"count", ""},    // Not a string
		{"flag", ""},     // Not a string
		{"nested", ""},   // Not a string
	}

	for _, tt := range tests {
		result := getString(m, tt.key)
		if result != tt.expected {
			t.Errorf("getString(m, %q): got %q, want %q", tt.key, result, tt.expected)
		}
	}
}

func TestConfigWithNilSelectors(t *testing.T) {
	cfg := Config{
		Type:     AuthTypeBearer,
		LoginURL: "",
		Credentials: map[string]string{
			"token": "bearer-token-123",
		},
		Selectors: nil, // No selectors for bearer auth
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal Config: %v", err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Config: %v", err)
	}

	if decoded.Selectors != nil {
		t.Error("Selectors should be nil for bearer auth")
	}
}

func TestConfigYAMLTags(t *testing.T) {
	// Verify struct tags are properly set by checking JSON output
	cfg := Config{
		Type:             AuthTypeForm,
		LoginURL:         "https://example.com/login",
		SuccessIndicator: ".success",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Check that the JSON uses the correct field names
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, ok := raw["type"]; !ok {
		t.Error("Expected 'type' field in JSON")
	}
	if _, ok := raw["loginUrl"]; !ok {
		t.Error("Expected 'loginUrl' field in JSON")
	}
	if _, ok := raw["successIndicator"]; !ok {
		t.Error("Expected 'successIndicator' field in JSON")
	}
}

func TestEmptySession(t *testing.T) {
	session := Session{
		Cookies:        []Cookie{},
		LocalStorage:   map[string]string{},
		SessionStorage: map[string]string{},
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("Failed to marshal empty Session: %v", err)
	}

	var decoded Session
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Session: %v", err)
	}

	if decoded.Cookies == nil {
		t.Error("Cookies should not be nil")
	}
	if decoded.LocalStorage == nil {
		t.Error("LocalStorage should not be nil")
	}
	if decoded.SessionStorage == nil {
		t.Error("SessionStorage should not be nil")
	}
}
