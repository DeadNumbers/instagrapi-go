package instagrapi

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Login logs in with username and password.
func (c *Client) Login(username, password string) error {
	c.Username = username

	if c.UserID > 0 && !c.shouldRelogin() {
		return nil // already logged in
	}

	// Pre-login flow
	c.preLoginFlow()

	// Encrypt password and login
	encPassword := c.passwordEncrypt(password)
	data := map[string]any{
		"jazoest":             GenerateJazoest(c.PhoneID),
		"country_codes":       fmt.Sprintf(`[{"country_code":"%d","source":["default"]}]`, c.Settings.CountryCode),
		"phone_id":            c.PhoneID,
		"enc_password":        encPassword,
		"username":            c.Username,
		"adid":                c.AdvertisingID,
		"guid":                c.UUID,
		"device_id":           c.AndroidDeviceID,
		"google_tokens":       "[]",
		"login_attempt_count": "0",
	}

	result, err := c.privateRequest("accounts/login/", data, nil)
	if err != nil {
		return err
	}

	// Parse authorization from response headers
	if c.lastResponse != nil {
		c.AuthData = c.parseAuthorization(c.lastResponse.Header.Get("ig-set-authorization"))
	}

	// Extract session ID and user ID
	if sessionID, ok := result["sessionid"].(string); ok {
		c.SessionID = sessionID
		c.Settings.Cookies["sessionid"] = sessionID
	}
	if dsUID, ok := result["ds_user_id"].(string); ok {
		c.UserID, _ = strconv.ParseInt(dsUID, 10, 64)
		c.Settings.Cookies["ds_user_id"] = dsUID
	}

	// Post-login flow
	c.postLoginFlow()
	c.Settings.LastLogin = float64(time.Now().Unix())

	return nil
}

// LoginBySessionID logs in using a session ID.
func (c *Client) LoginBySessionID(sessionID string) error {
	userMatch := regexp.MustCompile(`^\d+`).FindString(sessionID)
	if userMatch == "" || len(sessionID) <= 30 {
		return &ClientError{Message: "Invalid sessionid"}
	}

	userID, _ := strconv.ParseInt(userMatch, 10, 64)
	c.Settings.Cookies["sessionid"] = sessionID
	c.SessionID = sessionID

	c.AuthData = map[string]any{
		"ds_user_id":                     userMatch,
		"sessionid":                      sessionID,
		"should_use_header_over_cookies": true,
	}

	// Fetch user info to get username
	userInfo, err := c.UserInfoByPK(userID)
	if err != nil {
		return err
	}

	c.UserID = userID
	c.Username = userInfo.Username
	c.Settings.Cookies["ds_user_id"] = userMatch

	return nil
}

// Logout logs out the current user.
func (c *Client) Logout() error {
	_, err := c.privateRequest("accounts/logout/", map[string]any{
		"sessionid": c.SessionID,
	}, nil)
	c.UserID = 0
	c.Username = ""
	c.AuthData = make(map[string]any)
	return err
}

// IsLoggedIn checks if the client is currently logged in.
func (c *Client) IsLoggedIn() bool {
	return c.UserID > 0 && c.SessionID != ""
}

// SessionID returns the current session ID.
func (c *Client) SessionIDValue() string {
	return c.SessionID
}

// UserIDValue returns the current user ID.
func (c *Client) UserIDValue() int64 {
	return c.UserID
}

// UsernameValue returns the current username.
func (c *Client) UsernameValue() string {
	return c.Username
}

// TokenValue returns the CSRF token.
func (c *Client) TokenValue() string {
	if c.Token == "" {
		c.Token = GenToken(64, false)
	}
	return c.Token
}

// RankToken returns the rank token used for ranking operations.
func (c *Client) RankToken() string {
	return fmt.Sprintf("%d_%s", c.UserID, c.UUID)
}

// SetTimezoneOffset sets the timezone offset in seconds from UTC.
func (c *Client) SetTimezoneOffset(offsetSeconds int) {
	c.Settings.TimezoneOffset = offsetSeconds
}

// SetLocale sets the locale (e.g., "en_US").
func (c *Client) SetLocale(locale string) {
	c.Settings.Locale = locale
	parts := splitAtLast(locale, "_")
	if len(parts) == 2 {
		c.Settings.Country = parts[1]
	}
}

// SetCountry sets the country code (e.g., "US").
func (c *Client) SetCountry(country string) {
	c.Settings.Country = country
}

// SetPushDisabled disables push notifications in requests.
func (c *Client) SetPushDisabled(disabled bool) {
	c.Settings.PushDisabled = disabled
}

// SetUserAgent sets the User-Agent header value.
func (c *Client) SetUserAgent(ua string) {
	c.Settings.UserAgent = ua
}

// SetDeviceSettings sets the device profile settings.
func (c *Client) SetDeviceSettings(settings map[string]any) {
	c.Settings.DeviceSettings = settings
}

// GetSettings returns a copy of current settings.
func (c *Client) GetSettings() map[string]any {
	return map[string]any{
		"cookies":            c.Settings.Cookies,
		"authorization_data": c.AuthData,
		"last_login":         c.Settings.LastLogin,
		"timezone_offset":    c.Settings.TimezoneOffset,
		"timezone_name":      c.Settings.TimezoneName,
		"push_disabled":      c.Settings.PushDisabled,
		"locale":             c.Settings.Locale,
		"country":            c.Settings.Country,
		"country_code":       c.Settings.CountryCode,
		"user_agent":         c.Settings.UserAgent,
		"device_settings":    c.Settings.DeviceSettings,
		"uuids":              c.Settings.UUIDs,
		"mid":                c.Mid,
	}
}

// preLoginFlow performs pre-login emulation.
func (c *Client) preLoginFlow() {
	c.syncLauncher(false)
}

// postLoginFlow performs post-login initialization.
func (c *Client) postLoginFlow() {
	c.getReelsTrayFeed("cold_start")
	c.getTimelineFeed("cold_start_fetch")
}

// syncLauncher syncs launcher data with Instagram.
func (c *Client) syncLauncher(login bool) map[string]any {
	data := map[string]any{
		"id":                      c.UUID,
		"server_config_retrieval": "1",
	}
	if !login {
		data["_uid"] = strconv.FormatInt(c.UserID, 10)
		data["_uuid"] = c.UUID
		data["_csrftoken"] = c.TokenValue()
	}

	result, err := c.privateRequest("launcher/sync/", data, nil)
	if err != nil {
		c.Logger.Debug(fmt.Sprintf("syncLauncher error: %v", err))
	}
	return result
}

// getReelsTrayFeed gets the reels tray feed.
func (c *Client) getReelsTrayFeed(reason string) map[string]any {
	data := map[string]any{
		"payload": jsonMarshalCompact(map[string]any{
			"source": "cold_start",
			"app_pref": map[string]any{
				"reels_tray_expiry": "1970-01-01T00:00:00.000Z",
			},
		}),
		"reason": reason,
	}
	result, err := c.privateRequest("feed/reels_tray/", data, nil)
	if err != nil {
		c.Logger.Debug(fmt.Sprintf("getReelsTrayFeed error: %v", err))
	}
	return result
}

// getTimelineFeed gets the timeline feed.
func (c *Client) getTimelineFeed(reason string) map[string]any {
	data := map[string]any{
		"is_prefetch":     false,
		"feed_view_info":  "[]",
		"reason":          reason,
		"surface":         "hamburger",
		"is_carousel_bff": false,
		"card_name":       "feed_top",
	}

	params := map[string]string{
		"items_request_timestamp": strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		"method":                  "sync",
		"authenticity_token":      c.TokenValue(),
	}

	result, err := c.privateRequest("feed/timeline/", data, params)
	if err != nil {
		c.Logger.Debug(fmt.Sprintf("getTimelineFeed error: %v", err))
	}
	return result
}

// shouldRelogin checks if re-login is needed.
func (c *Client) shouldRelogin() bool {
	if c.Settings.LastLogin == 0 {
		return true
	}
	return time.Since(time.Unix(int64(c.Settings.LastLogin), 0)).Hours() > 24
}

// splitAtLast splits a string at the last occurrence of sep.
func splitAtLast(s, sep string) []string {
	idx := -1
	for i := len(s) - len(sep); i >= 0; i-- {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx < 0 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}
