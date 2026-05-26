package instagrapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSaveSession tests saving a session to a JSON file.
func TestSaveSession(t *testing.T) {
	c := NewClient()
	c.SessionID = "test_session_12345"
	c.UserID = 987654321
	c.Username = "testuser"
	c.Token = "test_token_value"
	c.Mid = "test_mid_value"

	// Set up some session data
	c.Settings.Cookies = map[string]string{
		"sessionid":  c.SessionID,
		"ds_user_id": "987654321",
	}
	c.AuthData = map[string]any{
		"sessionid":                      c.SessionID,
		"ds_user_id":                     "987654321",
		"should_use_header_over_cookies": true,
	}

	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "session.json")

	err := c.SaveSession(sessionPath)
	if err != nil {
		t.Fatalf("SaveSession error: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("Failed to read saved session file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Saved session file is empty")
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Saved session is not valid JSON: %v", err)
	}

	// Check that key fields are present
	if _, ok := parsed["session_id"]; !ok {
		t.Error("Missing 'session_id' in saved session")
	}
	if _, ok := parsed["user_id"]; !ok {
		t.Error("Missing 'user_id' in saved session")
	}
	if _, ok := parsed["username"]; !ok {
		t.Error("Missing 'username' in saved session")
	}

	t.Logf("Session saved successfully to %s (%d bytes)", sessionPath, len(data))
}

// TestLoadSession tests loading a session from a JSON file.
func TestLoadSession(t *testing.T) {
	c := NewClient()

	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "session.json")

	// Create a session file manually
	sessionData := SessionData{
		SessionID:            "loaded_session_67890",
		UserID:               123456789,
		Username:             "loadtestuser",
		Token:                "loaded_token_value",
		Mid:                  "loaded_mid_value",
		BloksVersioningID:    "test_bloks_version",
		RequestTimeout:       5.0,
		PublicRequestRetries: 10,
		Cookies: map[string]string{
			"sessionid":  "loaded_session_67890",
			"ds_user_id": "123456789",
			"csrftoken":  "test_csrf",
		},
		AuthData: map[string]any{
			"sessionid":                      "loaded_session_67890",
			"ds_user_id":                     "123456789",
			"should_use_header_over_cookies": true,
		},
		TimezoneOffset: -7200, // UTC+2
		Locale:         "ru_RU",
		Country:        "RU",
		CountryCode:    7,
		UserAgent:      "TestAgent/1.0",
		PushDisabled:   true,
	}

	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test session data: %v", err)
	}
	if err := os.WriteFile(sessionPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test session file: %v", err)
	}

	// Load the session
	err = c.LoadSession(sessionPath)
	if err != nil {
		t.Fatalf("LoadSession error: %v", err)
	}

	// Verify loaded values
	if c.SessionID != "loaded_session_67890" {
		t.Errorf("SessionID = %q; want 'loaded_session_67890'", c.SessionID)
	}
	if c.UserID != 123456789 {
		t.Errorf("UserID = %d; want 123456789", c.UserID)
	}
	if c.Username != "loadtestuser" {
		t.Errorf("Username = %q; want 'loadtestuser'", c.Username)
	}
	if c.Token != "loaded_token_value" {
		t.Errorf("Token = %q; want 'loaded_token_value'", c.Token)
	}
	if c.Mid != "loaded_mid_value" {
		t.Errorf("Mid = %q; want 'loaded_mid_value'", c.Mid)
	}
	if c.Settings.Locale != "ru_RU" {
		t.Errorf("Locale = %q; want 'ru_RU'", c.Settings.Locale)
	}
	if c.Settings.TimezoneOffset != -7200 {
		t.Errorf("TimezoneOffset = %d; want -7200", c.Settings.TimezoneOffset)
	}
	if c.Settings.UserAgent != "TestAgent/1.0" {
		t.Errorf("UserAgent = %q; want 'TestAgent/1.0'", c.Settings.UserAgent)
	}

	// Verify maps are initialized (not nil)
	if c.AuthData == nil {
		t.Error("AuthData should not be nil after loading")
	}
	if c.Settings.Cookies == nil {
		t.Error("Cookies should not be nil after loading")
	}

	t.Log("Session loaded successfully")
}

// TestLoadSessionNonExistent tests loading a non-existent session file.
func TestLoadSessionNonExistent(t *testing.T) {
	c := NewClient()
	err := c.LoadSession("/nonexistent/path/session.json")
	if err == nil {
		t.Error("Expected error when loading non-existent session file, got nil")
	}
}

// TestLoadSessionInvalidJSON tests loading an invalid JSON file.
func TestLoadSessionInvalidJSON(t *testing.T) {
	c := NewClient()

	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(sessionPath, []byte("not valid json {{{"), 0600); err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	err := c.LoadSession(sessionPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON, got nil")
	}
}

// TestSessionExists tests the SessionExists helper function.
func TestSessionExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent file
	if SessionExists(filepath.Join(tmpDir, "nonexistent.json")) {
		t.Error("SessionExists should return false for non-existent file")
	}

	// Create a file and check again
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !SessionExists(testFile) {
		t.Error("SessionExists should return true for existing file")
	}
}

// TestSaveLoadRoundTrip tests saving and loading a session in one round-trip.
func TestSaveLoadRoundTrip(t *testing.T) {
	// Create original client with session data
	original := NewClient()
	original.SessionID = "roundtrip_session"
	original.UserID = 111222333
	original.Username = "roundtripuser"
	original.Token = "rt_token"
	original.Mid = "rt_mid"
	original.Settings.Locale = "en_US"
	original.Settings.Country = "US"
	original.Settings.TimezoneOffset = -14400
	original.Settings.UserAgent = "RoundTrip/1.0"
	original.Settings.Cookies = map[string]string{
		"sessionid": "roundtrip_session",
	}
	original.AuthData = map[string]any{
		"sessionid":  "roundtrip_session",
		"ds_user_id": "111222333",
	}

	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "roundtrip.json")

	// Save original session
	if err := original.SaveSession(sessionPath); err != nil {
		t.Fatalf("SaveSession error: %v", err)
	}

	// Create new client and load the session
	loaded := NewClient()
	if err := loaded.LoadSession(sessionPath); err != nil {
		t.Fatalf("LoadSession error: %v", err)
	}

	// Verify all fields match
	checks := []struct {
		name     string
		actual   any
		expected any
	}{
		{"SessionID", loaded.SessionID, original.SessionID},
		{"UserID", loaded.UserID, original.UserID},
		{"Username", loaded.Username, original.Username},
		{"Token", loaded.Token, original.Token},
		{"Mid", loaded.Mid, original.Mid},
		{"Locale", loaded.Settings.Locale, original.Settings.Locale},
		{"Country", loaded.Settings.Country, original.Settings.Country},
		{"TimezoneOffset", loaded.Settings.TimezoneOffset, original.Settings.TimezoneOffset},
		{"UserAgent", loaded.Settings.UserAgent, original.Settings.UserAgent},
	}

	for _, check := range checks {
		if check.actual != check.expected {
			t.Errorf("%s: got %v, want %v", check.name, check.actual, check.expected)
		}
	}

	// Verify cookies were loaded
	if loaded.Settings.Cookies["sessionid"] != "roundtrip_session" {
		t.Error("Cookie 'sessionid' not properly loaded")
	}

	t.Log("Round-trip save/load successful")
}

// TestSessionDataJSONSerialization tests that SessionData can be serialized/deserialized.
func TestSessionDataJSONSerialization(t *testing.T) {
	original := &SessionData{
		SessionID:            "json_test_session",
		UserID:               456789012,
		Username:             "jsontestuser",
		Cookies:              map[string]string{"sessionid": "json_test_session"},
		AuthData:             map[string]any{"sessionid": "json_test_session"},
		TimezoneOffset:       -5400,
		Locale:               "de_DE",
		Country:              "DE",
		CountryCode:          49,
		UserAgent:            "JSONTest/1.0",
		PushDisabled:         true,
		RequestTimeout:       3.5,
		PublicRequestRetries: 7,
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal SessionData: %v", err)
	}

	// Deserialize from JSON
	var restored SessionData
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal SessionData: %v", err)
	}

	// Verify all fields match
	checks := []struct {
		name     string
		actual   any
		expected any
	}{
		{"SessionID", restored.SessionID, original.SessionID},
		{"UserID", restored.UserID, original.UserID},
		{"Username", restored.Username, original.Username},
		{"TimezoneOffset", restored.TimezoneOffset, original.TimezoneOffset},
		{"Locale", restored.Locale, original.Locale},
		{"CountryCode", restored.CountryCode, original.CountryCode},
		{"UserAgent", restored.UserAgent, original.UserAgent},
		{"PushDisabled", restored.PushDisabled, original.PushDisabled},
		{"RequestTimeout", restored.RequestTimeout, original.RequestTimeout},
	}

	for _, check := range checks {
		if check.actual != check.expected {
			t.Errorf("%s: got %v, want %v", check.name, check.actual, check.expected)
		}
	}

	t.Log("SessionData JSON serialization successful")
}
