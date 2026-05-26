package instagrapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// privateRequest sends a private API request to Instagram.
// endpoint: the API endpoint path (e.g., "media/12345/info/")
// data: POST body as map (will be JSON-encoded and signed)
// params: URL query parameters
func (c *Client) privateRequest(endpoint string, data map[string]any, params map[string]string) (map[string]any, error) {
	c.privateReqCount++

	// Apply delay if configured
	if c.delayRange[0] > 0 && c.delayRange[1] > c.delayRange[0] {
		time.Sleep(time.Duration(float64(time.Second) * (c.delayRange[0] + rand.Float64()*(c.delayRange[1]-c.delayRange[0]))))
	}

	headers := c.buildBaseHeaders()

	var bodyReader io.Reader
	if data != nil && len(data) > 0 {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
		headers["Content-Type"] = "application/json; charset=utf-8"
	}

	apiURL := fmt.Sprintf("https://%s/api/v2/%s", apiDomain, strings.TrimPrefix(endpoint, "/"))

	req, err := http.NewRequest("POST", apiURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if auth := c.authorizationHeader(); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, &ClientConnectionError{ClientError: ClientError{Message: err.Error()}}
	}
	defer resp.Body.Close()

	c.lastResponse = resp
	c.requestLog(resp, "POST", apiURL)

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if resp.StatusCode >= 400 {
		err := c.handlePrivateError(resp.StatusCode, bodyBytes)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, &ClientJSONDecodeError{ClientError: ClientError{Message: fmt.Sprintf("failed to parse JSON: %v", err)}}
	}

	c.lastJSON = result

	status, _ := result["status"].(string)
	if status == "fail" {
		msg, _ := result["message"].(string)
		return nil, &ClientError{Message: msg, Response: result}
	}

	return result, nil
}

// privateGet sends a GET request to the private API.
func (c *Client) privateGet(endpoint string, params map[string]string) (map[string]any, error) {
	c.privateReqCount++

	if c.delayRange[0] > 0 && c.delayRange[1] > c.delayRange[0] {
		time.Sleep(time.Duration(float64(time.Second) * (c.delayRange[0] + rand.Float64()*(c.delayRange[1]-c.delayRange[0]))))
	}

	headers := c.buildBaseHeaders()
	delete(headers, "Content-Type")

	apiURL := fmt.Sprintf("https://%s/api/v2/%s", apiDomain, strings.TrimPrefix(endpoint, "/"))

	if params != nil {
		q := strings.Builder{}
		first := true
		for k, v := range params {
			if first {
				q.WriteString("?")
				first = false
			} else {
				q.WriteString("&")
			}
			q.WriteString(k)
			q.WriteString("=")
			q.WriteString(v)
		}
		apiURL += q.String()
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if auth := c.authorizationHeader(); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, &ClientConnectionError{ClientError: ClientError{Message: err.Error()}}
	}
	defer resp.Body.Close()

	c.lastResponse = resp
	c.requestLog(resp, "GET", apiURL)

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if resp.StatusCode >= 400 {
		err := c.handlePrivateError(resp.StatusCode, bodyBytes)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, &ClientJSONDecodeError{ClientError: ClientError{Message: fmt.Sprintf("failed to parse JSON: %v", err)}}
	}

	c.lastJSON = result

	status, _ := result["status"].(string)
	if status == "fail" {
		msg, _ := result["message"].(string)
		return nil, &ClientError{Message: msg, Response: result}
	}

	return result, nil
}

// handlePrivateError processes error responses from the private API.
func (c *Client) handlePrivateError(statusCode int, body []byte) error {
	var result map[string]any
	json.Unmarshal(body, &result) // ignore errors for error parsing

	message, _ := result["message"].(string)
	errorType, _ := result["error_type"].(string)

	switch statusCode {
	case 403:
		if message == "login_required" {
			return &LoginRequired{PrivateError: PrivateError{ClientError: ClientError{Message: message, Response: result}}}
		}
		return &ClientForbiddenError{ClientError: ClientError{Message: message, Code: statusCode, Response: result}}

	case 400:
		if errorType == "two_factor_required" || result["two_factor_info"] != nil {
			return &TwoFactorRequired{PrivateError: PrivateError{ClientError: ClientError{Message: "Two-factor authentication required", Response: result}}}
		}
		if message == "challenge_required" {
			return &ChallengeRequired{ChallengeError: ChallengeError{ClientError: ClientError{Message: message, Response: result}}}
		}
		if errorType == "rate_limit_error" {
			return &RateLimitError{PrivateError: PrivateError{ClientError: ClientError{Message: "Rate limit exceeded", Response: result}}}
		}
		if errorType == "sentry_block" {
			return &SentryBlock{PrivateError: PrivateError{ClientError: ClientError{Message: message, Response: result}}}
		}
		if message == "VideoTooLongException" {
			return &VideoTooLongException{PrivateError: PrivateError{ClientError: ClientError{Message: message, Response: result}}}
		}
		if strings.Contains(message, "Invalid target user") {
			return &InvalidTargetUser{PrivateError: PrivateError{ClientError: ClientError{Message: message, Response: result}}}
		}
		if strings.Contains(message, "Invalid media_id") {
			return &InvalidMediaId{PrivateError: PrivateError{ClientError: ClientError{Message: message, Response: result}}}
		}
		if strings.Contains(message, "Media is unavailable") || strings.Contains(message, "Media not found") {
			return &MediaUnavailable{PrivateError: PrivateError{ClientError: ClientError{Message: message, Response: result}}}
		}
		if strings.Contains(message, "unable to fetch followers") {
			return &UserNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found", Response: result}}}}
		}
		return &ClientBadRequestError{ClientError: ClientError{Message: message, Code: statusCode, Response: result}}

	case 429:
		return &ClientThrottledError{ClientError: ClientError{Message: "Too many requests", Code: statusCode, Response: result}}

	case 401:
		return &ClientUnauthorizedError{ClientError: ClientError{Message: "Unauthorized", Code: statusCode, Response: result}}

	case 404:
		if strings.TrimSpace(string(body)) == "Not Found" {
			return &ChallengeRequired{ChallengeError: ChallengeError{ClientError: ClientError{Message: "Masked challenge (404)", Response: result}}}
		}
		return &ClientNotFoundError{ClientError: ClientError{Message: "Endpoint not found", Code: statusCode, Response: result}}

	default:
		return &ClientError{Message: fmt.Sprintf("HTTP %d: %s", statusCode, message), Code: statusCode, Response: result}
	}
}

// publicRequest sends a public API request (no authentication).
func (c *Client) publicRequest(method string, urlStr string, data map[string]any, headers map[string]string) (map[string]any, error) {
	var bodyReader io.Reader
	if data != nil && len(data) > 0 {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.PublicClient.Do(req)
	if err != nil {
		return nil, &ClientConnectionError{ClientError: ClientError{Message: err.Error()}}
	}
	defer resp.Body.Close()

	c.lastResponse = resp

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if resp.StatusCode >= 400 {
		return nil, &ClientError{Message: fmt.Sprintf("HTTP %d", resp.StatusCode), Code: resp.StatusCode, Response: string(bodyBytes)}
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, &ClientJSONDecodeError{ClientError: ClientError{Message: fmt.Sprintf("failed to parse JSON: %v", err)}}
	}

	c.lastJSON = result
	return result, nil
}

// publicGraphqlRequest sends a GraphQL request to the public API.
func (c *Client) publicGraphqlRequest(variables map[string]any, queryHash string) (map[string]any, error) {
	data := map[string]any{
		"query_id":  "", // not needed for public
		"variables": variables,
		"doc_id":    queryHash,
	}

	headers := map[string]string{
		"User-Agent":      c.Settings.UserAgent,
		"Accept":          "application/json",
		"Accept-Encoding": "gzip, deflate",
		"Content-Type":    "application/json",
		"X-IG-App-ID":     "567067343352427",
	}

	return c.publicRequest("POST", fmt.Sprintf("https://%s/graphql/query/", apiDomain), data, headers)
}

// privateGraphqlRequest sends a GraphQL request to the private API.
func (c *Client) privateGraphqlRequest(data map[string]any) (map[string]any, error) {
	headers := c.buildBaseHeaders()
	headers["Content-Type"] = "application/json"

	apiURL := fmt.Sprintf("https://%s/api/v1/", apiDomain)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if auth := c.authorizationHeader(); auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, &ClientConnectionError{ClientError: ClientError{Message: err.Error()}}
	}
	defer resp.Body.Close()

	c.lastResponse = resp

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if resp.StatusCode >= 400 {
		err := c.handlePrivateError(resp.StatusCode, bodyBytes)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, &ClientJSONDecodeError{ClientError: ClientError{Message: fmt.Sprintf("failed to parse JSON: %v", err)}}
	}

	c.lastJSON = result
	return result, nil
}
