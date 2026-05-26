package instagrapi

import (
	"encoding/json"
	"fmt"
	"os"
)

// SessionData represents a serializable Instagram session.
type SessionData struct {
	Cookies              map[string]string `json:"cookies"`
	AuthData             map[string]any    `json:"authorization_data"`
	LastLogin            float64           `json:"last_login,omitempty"`
	TimezoneOffset       int               `json:"timezone_offset,omitempty"`
	TimezoneName         string            `json:"timezone_name,omitempty"`
	PushDisabled         bool              `json:"push_disabled,omitempty"`
	Locale               string            `json:"locale,omitempty"`
	Country              string            `json:"country,omitempty"`
	CountryCode          int               `json:"country_code,omitempty"`
	UserAgent            string            `json:"user_agent,omitempty"`
	DeviceSettings       map[string]any    `json:"device_settings,omitempty"`
	UUIDs                map[string]string `json:"uuids,omitempty"`
	Mid                  string            `json:"mid,omitempty"`
	SessionID            string            `json:"session_id,omitempty"`
	UserID               int64             `json:"user_id,omitempty"`
	Username             string            `json:"username,omitempty"`
	Token                string            `json:"token,omitempty"`
	BloksVersioningID    string            `json:"bloks_versioning_id,omitempty"`
	RequestTimeout       float64           `json:"request_timeout,omitempty"`
	PublicRequestRetries int               `json:"public_request_retries,omitempty"`
}

// SaveSession saves the current client session to a JSON file.
func (c *Client) SaveSession(path string) error {
	session := &SessionData{
		Cookies:              c.Settings.Cookies,
		AuthData:             c.AuthData,
		LastLogin:            c.Settings.LastLogin,
		TimezoneOffset:       c.Settings.TimezoneOffset,
		TimezoneName:         c.Settings.TimezoneName,
		PushDisabled:         c.Settings.PushDisabled,
		Locale:               c.Settings.Locale,
		Country:              c.Settings.Country,
		CountryCode:          c.Settings.CountryCode,
		UserAgent:            c.Settings.UserAgent,
		DeviceSettings:       c.Settings.DeviceSettings,
		UUIDs:                c.Settings.UUIDs,
		Mid:                  c.Mid,
		SessionID:            c.SessionID,
		UserID:               c.UserID,
		Username:             c.Username,
		Token:                c.Token,
		BloksVersioningID:    c.Settings.BloksVersioningID,
		RequestTimeout:       c.Settings.RequestTimeout,
		PublicRequestRetries: c.Settings.PublicRequestRetries,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	c.Logger.Info(fmt.Sprintf("Session saved to %s", path))
	return nil
}

// LoadSession loads a session from a JSON file and restores it to the client.
func (c *Client) LoadSession(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	c.Settings.Cookies = session.Cookies
	c.AuthData = session.AuthData
	c.Settings.LastLogin = session.LastLogin
	c.Settings.TimezoneOffset = session.TimezoneOffset
	c.Settings.TimezoneName = session.TimezoneName
	c.Settings.PushDisabled = session.PushDisabled
	c.Settings.Locale = session.Locale
	c.Settings.Country = session.Country
	c.Settings.CountryCode = session.CountryCode
	c.Settings.UserAgent = session.UserAgent
	c.Settings.DeviceSettings = session.DeviceSettings
	c.Settings.UUIDs = session.UUIDs
	c.Mid = session.Mid
	c.SessionID = session.SessionID
	c.UserID = session.UserID
	c.Username = session.Username
	c.Token = session.Token
	c.Settings.BloksVersioningID = session.BloksVersioningID
	c.Settings.RequestTimeout = session.RequestTimeout
	c.Settings.PublicRequestRetries = session.PublicRequestRetries

	if c.AuthData == nil {
		c.AuthData = make(map[string]any)
	}
	if c.Settings.Cookies == nil {
		c.Settings.Cookies = make(map[string]string)
	}
	if c.Settings.UUIDs == nil {
		c.Settings.UUIDs = make(map[string]string)
	}

	c.Logger.Info(fmt.Sprintf("Session loaded for user %s (ID: %d)", c.Username, c.UserID))
	return nil
}

// SessionExists checks if a session file exists at the given path.
func SessionExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
