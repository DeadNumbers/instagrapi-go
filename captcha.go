package instagrapi

// CaptchaHandler is a function type for solving captcha challenges.
// It takes challenge details and returns the solved captcha token.
type CaptchaHandler func(map[string]any) string

// SetCaptchaHandler sets a custom handler function for solving captcha challenges.
func (c *Client) SetCaptchaHandler(handler CaptchaHandler) {
	c.captchaHandler = handler
}

// CaptchaResolve resolves a captcha challenge using the registered handler.
func (c *Client) CaptchaResolve(challengeDetails map[string]any) (string, error) {
	if c.captchaHandler == nil {
		return "", &CaptchaChallengeRequired{
			ClientError: ClientError{
				Message: "No captcha handler is configured. Use client.SetCaptchaHandler() to set one.",
			},
			ChallengeDetails: challengeDetails,
		}
	}

	siteKey, _ := challengeDetails["site_key"].(string)
	pageURL, _ := challengeDetails["page_url"].(string)
	challengeType, _ := challengeDetails["challenge_type"].(string)
	rawJSON, _ := challengeDetails["raw_challenge_json"].(map[string]any)

	detailsToPass := map[string]any{
		"site_key":           siteKey,
		"page_url":           pageURL,
		"challenge_type":     challengeType,
		"raw_challenge_json": rawJSON,
	}

	token := c.captchaHandler(detailsToPass)
	if token == "" {
		return "", &CaptchaChallengeRequired{
			ClientError: ClientError{
				Message: "Captcha handler ran but did not return a valid token string.",
			},
			ChallengeDetails: challengeDetails,
		}
	}

	return token, nil
}
