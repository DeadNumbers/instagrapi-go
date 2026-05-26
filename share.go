package instagrapi

import (
	"encoding/base64"
	"strings"
)

// ShareInfo gets a Share object from a base64-encoded code.
func (c *Client) ShareInfo(code string) (*Share, error) {
	data, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return nil, &ClientError{Message: "Invalid share code"}
	}

	decoded := strings.ReplaceAll(string(data), "\x1d", "")
	parts := strings.Split(decoded, ":")
	if len(parts) < 2 {
		return nil, &ClientError{Message: "Invalid share code format"}
	}

	return &Share{
		Type: parts[0],
		PK:   parts[1],
	}, nil
}

// ShareInfoByURL gets a Share object from an Instagram URL.
func (c *Client) ShareInfoByURL(url string) (*Share, error) {
	code := c.ShareCodeFromURL(url)
	return c.ShareInfo(code)
}

// ShareCodeFromURL extracts the share code from an Instagram URL.
func (c *Client) ShareCodeFromURL(url string) string {
	parts := strings.Split(strings.TrimPrefix(url, "https://www.instagram.com/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}
