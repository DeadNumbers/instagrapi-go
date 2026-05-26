package instagrapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client is the main Instagram API client. It embeds all mixin functionality.
type Client struct {
	// Session state (set after login)
	SessionID string
	UserID    int64
	Username  string
	Token     string // CSRF token
	Mid       string // from ig-set-x-mid response header

	// Device identifiers
	UUID            string
	PhoneID         string
	AndroidDeviceID string
	AdvertisingID   string

	// Configuration
	Settings        ClientSettings
	PublicClient    *http.Client
	PrivateClient   *http.Client
	Logger          Logger
	delayRange      [2]float64
	lastJSON        map[string]any
	lastResponse    *http.Response
	privateReqCount int

	// Authorization
	AuthData map[string]any // parsed from ig-set-authorization header

	// Timeline seen posts tracking
	timelineSeenPosts []string
}

// ClientSettings holds all configurable settings for the client.
type ClientSettings struct {
	Cookies                   map[string]string
	AuthorizationData         map[string]any
	LastLogin                 float64 // Unix timestamp
	TimezoneOffset            int     // seconds from UTC
	TimezoneName              string
	PushDisabled              bool
	Locale                    string
	Country                   string
	CountryCode               int
	UserAgent                 string
	DeviceSettings            map[string]any
	UUIDs                     map[string]string
	BloksVersioningID         string
	TLSVerify                 bool
	RequestTimeout            float64 // seconds for individual requests
	PublicRequestRetries      int
	PublicRequestRetryWait    time.Duration
	SessionRetryTotal         int
	SessionRetryBackoffFactor float64
	SessionRetryStatuses      []int
	IGURUR                    string
	IGWWWClaim                string
}

// Logger is the logging interface.
type Logger interface {
	Info(msg string)
	Debug(msg string)
	Error(msg string)
	Warn(msg string)
}

// DefaultLogger implements Logger using the standard log package.
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string)  { log.Println("[INFO]", msg) }
func (l *DefaultLogger) Debug(msg string) { log.Println("[DEBUG]", msg) }
func (l *DefaultLogger) Error(msg string) { log.Println("[ERROR]", msg) }
func (l *DefaultLogger) Warn(msg string)  { log.Println("[WARN]", msg) }

// NewClient creates a new Instagram client with default settings.
func NewClient() *Client {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	c := &Client{
		Logger: &DefaultLogger{},
		Settings: ClientSettings{
			TimezoneOffset:            -14400, // New York, GMT-4
			Locale:                    "en_US",
			Country:                   "US",
			CountryCode:               1,
			PushDisabled:              true,
			TLSVerify:                 true,
			RequestTimeout:            1.0,
			PublicRequestRetries:      3,
			SessionRetryTotal:         3,
			SessionRetryBackoffFactor: 2.0,
		},
		UUID:            fmt.Sprintf("%x", r.Int63()),
		PhoneID:         fmt.Sprintf("%x", r.Int63()),
		AndroidDeviceID: fmt.Sprintf("android-%x", r.Int63()),
		AdvertisingID:   fmt.Sprintf("%x", r.Int63()),
	}

	c.PublicClient = c.newHTTPClient(false)
	c.PrivateClient = c.newHTTPClient(true)

	return c
}

// newHTTPClient creates an HTTP client with retry logic.
func (c *Client) newHTTPClient(isPrivate bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !c.Settings.TLSVerify,
		},
	}

	if isPrivate && c.Settings.SessionRetryTotal > 0 {
		statuses := c.Settings.SessionRetryStatuses
		if len(statuses) == 0 {
			statuses = []int{429, 500, 502, 503, 504}
		}
		tr.MaxIdleConnsPerHost = 1
	}

	return &http.Client{
		Transport: tr,
		Timeout:   time.Duration(c.Settings.RequestTimeout * 1e9),
		Jar:       nil, // cookies managed manually via headers
	}
}

// setProxy sets the proxy for both public and private clients.
func (c *Client) setProxy(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		parsed, err = url.Parse("http://" + proxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
	}

	proxy := http.ProxyFromEnvironment
	if parsed.Host != "" {
		proxy = http.ProxyURL(parsed)
	}

	for _, client := range []*http.Client{c.PublicClient, c.PrivateClient} {
		if tr, ok := client.Transport.(*http.Transport); ok {
			tr.Proxy = proxy
		}
	}

	return nil
}

// buildBaseHeaders returns the base headers for private API requests.
func (c *Client) buildBaseHeaders() map[string]string {
	headers := make(map[string]string)
	locale := strings.Replace(c.Settings.Locale, "-", "_", 1)

	headers["X-IG-App-Locale"] = locale
	headers["X-IG-Device-Locale"] = locale
	headers["X-IG-Mapped-Locale"] = locale
	headers["X-Pigeon-Session-Id"] = c.generateUUID("UFS-", "-1")
	headers["X-Pigeon-Rawclienttime"] = strconv.FormatFloat(float64(time.Now().Unix()), 'f', -1, 64)
	bandwidthSpeed := 2500 + randInt(500)
	headers["X-IG-Bandwidth-Speed-KBPS"] = strconv.Itoa(bandwidthSpeed) + ".0"
	totalBytes := 5000000 + randInt(85000000)
	headers["X-IG-Bandwidth-TotalBytes-B"] = strconv.Itoa(totalBytes)
	totalTimeMS := 2000 + randInt(7000)
	headers["X-IG-Bandwidth-TotalTime-MS"] = strconv.Itoa(totalTimeMS)
	headers["X-IG-App-Startup-Country"] = strings.ToUpper(c.Settings.Country)
	headers["X-Bloks-Version-Id"] = c.Settings.BloksVersioningID
	headers["X-IG-WWW-Claim"] = "0"
	headers["X-Bloks-Is-Layout-RTL"] = "false"
	headers["X-Bloks-Is-Panorama-Enabled"] = "true"
	headers["X-IG-Device-ID"] = c.UUID
	headers["X-IG-Family-Device-ID"] = c.PhoneID
	headers["X-IG-Android-ID"] = c.AndroidDeviceID
	headers["X-IG-Timezone-Offset"] = strconv.Itoa(c.Settings.TimezoneOffset)
	headers["X-IG-Connection-Type"] = "WIFI"
	headers["X-IG-Capabilities"] = "3brTv10="
	headers["X-IG-App-ID"] = "567067343352427"
	headers["Priority"] = "u=3"
	headers["User-Agent"] = c.Settings.UserAgent
	headers["Accept-Language"] = "en-US,en;q=0.9"
	headers["X-MID"] = c.Mid
	headers["Accept-Encoding"] = "gzip, deflate"
	headers["Host"] = apiDomain
	headers["X-FB-HTTP-Engine"] = "Tigon/MNS/TCP"
	headers["X-Tigon-Is-Retry"] = "False"
	headers["X-Zero-Balance"] = "INIT"
	headers["Connection"] = "keep-alive"

	if c.UserID > 0 {
		nextYear := time.Now().Add(365 * 24 * time.Hour).Unix()
		headers["IG-U-DS-USER-ID"] = strconv.FormatInt(c.UserID, 10)
		headers["IG-U-RUR"] = fmt.Sprintf("RVA,%d,%d:01f7f627f9ae4ce2874b2e04463efdb184340968b1b006fa88cb4cc69a942a04201e544c", c.UserID, nextYear)
	}

	if c.Settings.IGURUR != "" {
		headers["IG-U-RUR"] = c.Settings.IGURUR
	}
	if c.Settings.IGWWWClaim != "" {
		headers["X-IG-WWW-Claim"] = c.Settings.IGWWWClaim
	}

	return headers
}

// authorizationHeader returns the Authorization header value.
func (c *Client) authorizationHeader() string {
	if len(c.AuthData) == 0 {
		return ""
	}
	sid, _ := c.AuthData["sessionid"].(string)
	dsUID, _ := c.AuthData["ds_user_id"].(string)
	if sid == "" || dsUID == "" {
		return ""
	}
	return fmt.Sprintf("Instagram %s:%s", dsUID, sid)
}

// generateUUID generates a UUID with optional prefix and suffix.
func (c *Client) generateUUID(prefix string, suffix string) string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC4122
	return prefix + fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]) + suffix
}

// generateMutationToken generates a mutation token for GraphQL.
func (c *Client) generateMutationToken() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	result := make([]byte, 30)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// randInt returns a random int in [0, max).
func randInt(max int) int {
	return rand.Intn(max)
}

// urlEncode URL-encodes a string (URL-encoded form data).
func urlEncode(s string) string {
	return url.QueryEscape(s)
}

// jsonMarshal marshals to JSON with Instagram-style compact formatting.
func jsonMarshal(v any) (string, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "") // no indentation but use custom separators
	// Use default encoding which uses ", " between key-value pairs
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	result := buf.String()
	// Remove trailing newline and trim spaces
	result = strings.TrimSpace(result)
	return result, nil
}

// jsonMarshalCompact marshals to compact JSON (no spaces).
func jsonMarshalCompact(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// navigateJSON navigates through nested maps/slices using a path of keys/indices.
func navigateJSON(data any, path ...any) any {
	cur := data
	for _, p := range path {
		switch k := p.(type) {
		case int:
			if arr, ok := cur.([]any); ok && k >= 0 && k < len(arr) {
				cur = arr[k]
			} else {
				return nil
			}
		case string:
			if m, ok := cur.(map[string]any); ok {
				cur = m[k]
			} else {
				return nil
			}
		default:
			return nil
		}
		if cur == nil {
			return nil
		}
	}
	return cur
}

// randomDelay sleeps for a random duration within the configured delay range.
func (c *Client) randomDelay() {
	if c.delayRange[0] > 0 && c.delayRange[1] > c.delayRange[0] {
		d := time.Duration(float64(time.Second) * (c.delayRange[0] + rand.Float64()*(c.delayRange[1]-c.delayRange[0])))
		if d > 0 {
			time.Sleep(d)
		}
	}
}

// requestLog logs a private API request.
func (c *Client) requestLog(resp *http.Response, method string, reqURL string) {
	appVer := ""
	if devSet, ok := any(c.Settings.DeviceSettings).(map[string]any); ok {
		if av, ok := devSet["app_version"].(string); ok {
			appVer = av
		}
	}
	model := ""
	if devSet2, ok := any(c.Settings.DeviceSettings).(map[string]any); ok {
		if m, ok := devSet2["model"].(string); ok {
			model = m
		}
	}

	c.Logger.Debug(fmt.Sprintf("%s [%d] %s %s (%s, %s)",
		c.Username, resp.StatusCode, method, reqURL, appVer, model))
}

// doRequest executes an HTTP request and returns the response.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, err
	}
	c.lastResponse = resp
	return resp, nil
}

// parseAuthorization parses the ig-set-authorization header value.
func (c *Client) parseAuthorization(authHeader string) map[string]any {
	result := make(map[string]any)
	if authHeader == "" {
		return result
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Instagram" {
		return result
	}
	kvPairs := strings.Split(parts[1], ":")
	for _, pair := range kvPairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

// saveSession saves the current session to a file.
func (c *Client) saveSession(path string) error {
	session := map[string]any{
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

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// loadSession loads a session from a file.
func (c *Client) loadSession(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session map[string]any
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	if cookies, ok := session["cookies"].(map[string]any); ok {
		c.Settings.Cookies = make(map[string]string)
		for k, v := range cookies {
			if s, ok := v.(string); ok {
				c.Settings.Cookies[k] = s
			}
		}
	}

	if authData, ok := session["authorization_data"].(map[string]any); ok {
		c.AuthData = authData
	}

	if lastLogin, ok := session["last_login"].(float64); ok {
		c.Settings.LastLogin = lastLogin
	}
	if tzOffset, ok := session["timezone_offset"].(float64); ok {
		c.Settings.TimezoneOffset = int(tzOffset)
	}
	if tzName, ok := session["timezone_name"].(string); ok {
		c.Settings.TimezoneName = tzName
	}
	if pushDisabled, ok := session["push_disabled"].(bool); ok {
		c.Settings.PushDisabled = pushDisabled
	}
	if locale, ok := session["locale"].(string); ok {
		c.Settings.Locale = locale
	}
	if country, ok := session["country"].(string); ok {
		c.Settings.Country = country
	}
	if countryCode, ok := session["country_code"].(float64); ok {
		c.Settings.CountryCode = int(countryCode)
	}
	if userAgent, ok := session["user_agent"].(string); ok {
		c.Settings.UserAgent = userAgent
	}
	if deviceSettings, ok := session["device_settings"].(map[string]any); ok {
		c.Settings.DeviceSettings = deviceSettings
	}
	if uuids, ok := session["uuids"].(map[string]any); ok {
		c.Settings.UUIDs = make(map[string]string)
		for k, v := range uuids {
			if s, ok := v.(string); ok {
				c.Settings.UUIDs[k] = s
			}
		}
	}
	if mid, ok := session["mid"].(string); ok {
		c.Mid = mid
	}

	return nil
}

// setDelayRange sets the delay range between requests.
func (c *Client) setDelayRange(min, max float64) {
	c.delayRange[0] = min
	c.delayRange[1] = max
}
