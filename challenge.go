package instagrapi

import (
	"fmt"
)

// ChallengeResolve resolves a challenge response.
func (c *Client) ChallengeResolve(challenge map[string]any) error {
	apiPath, _ := navigateJSON(challenge, "api_path").(string)
	if apiPath == "" {
		return &ChallengeError{ClientError: ClientError{Message: "No API path in challenge"}}
	}

	challengeType, _ := navigateJSON(challenge, "challengeType").(string)
	switch challengeType {
	case "SubmitPhoneNumberForm":
		return c.challengeResolveContactForm(challenge, apiPath)
	default:
		return c.challengeResolveSimple(challenge, apiPath)
	}
}

// ChallengeCodeOrRaised gets the verification code from user input.
func (c *Client) ChallengeCodeOrRaised(username string) string {
	fmt.Printf("Enter verification code for %s: ", username)
	var code string
	fmt.Scanln(&code)
	return code
}

// ChallengeResolveContactForm handles contact form challenge.
func (c *Client) challengeResolveContactForm(challenge map[string]any, apiPath string) error {
	data := map[string]any{
		"method": "0", // SMS
	}

	result, err := c.privateRequest(
		apiPath,
		data,
		map[string]string{},
	)
	if err != nil {
		return err
	}

	nextAPIPath, _ := navigateJSON(result, "navigation", "forward").(string)
	if nextAPIPath == "" {
		return &ChallengeError{ClientError: ClientError{Message: "No forward path in challenge response"}}
	}

	code := c.ChallengeCodeOrRaised(c.Username)
	if code == "" {
		return &ChallengeRequired{ChallengeError: ChallengeError{ClientError: ClientError{Message: "No code provided"}}}
	}

	verifyData := map[string]any{
		"security_code": code,
	}

	_, err = c.privateRequest(
		nextAPIPath,
		verifyData,
		map[string]string{},
	)
	return err
}

// ChallengeResolveSimple handles simple challenge resolution.
func (c *Client) challengeResolveSimple(challenge map[string]any, apiPath string) error {
	code := c.ChallengeCodeOrRaised(c.Username)
	if code == "" {
		return &ChallengeRequired{ChallengeError: ChallengeError{ClientError: ClientError{Message: "No code provided"}}}
	}

	data := map[string]any{
		"security_code": code,
	}

	result, err := c.privateRequest(
		apiPath,
		data,
		map[string]string{},
	)
	if err != nil {
		return err
	}

	// Check if there's a next step
	nextAPIPath, _ := navigateJSON(result, "navigation", "forward").(string)
	if nextAPIPath == "" {
		return nil // Challenge resolved
	}

	// Follow the chain of challenges
	for i := 0; i < 5; i++ {
		code = c.ChallengeCodeOrRaised(c.Username)
		if code == "" {
			break
		}

		data["security_code"] = code
		result, err = c.privateRequest(
			nextAPIPath,
			data,
			map[string]string{},
		)
		if err != nil {
			return err
		}

		nextAPIPath, _ = navigateJSON(result, "navigation", "forward").(string)
		if nextAPIPath == "" {
			break
		}
	}

	return nil
}

// ChallengeCaptcha handles reCAPTCHA challenge.
func (c *Client) ChallengeCaptcha(challenge map[string]any, recaptchaResponse string) error {
	apiPath, _ := navigateJSON(challenge, "api_path").(string)
	if apiPath == "" {
		return &ChallengeError{ClientError: ClientError{Message: "No API path in captcha challenge"}}
	}

	data := map[string]any{
		"g-recaptcha-response": recaptchaResponse,
	}

	_, err := c.privateRequest(
		apiPath,
		data,
		map[string]string{},
	)
	return err
}

// ChallengeVerifySMS verifies SMS code.
func (c *Client) ChallengeVerifySMS(challenge map[string]any, securityCode string) error {
	apiPath, _ := navigateJSON(challenge, "navigation", "forward").(string)
	if apiPath == "" {
		return &ChallengeError{ClientError: ClientError{Message: "No forward path in SMS challenge"}}
	}

	data := map[string]any{
		"security_code": securityCode,
	}

	result, err := c.privateRequest(
		apiPath,
		data,
		map[string]string{},
	)
	if err != nil {
		return err
	}

	// Check for next challenge step
	nextAPIPath, _ := navigateJSON(result, "navigation", "forward").(string)
	if nextAPIPath == "" {
		return nil // Resolved
	}

	return &ChallengeRequired{ChallengeError: ChallengeError{ClientError: ClientError{Message: fmt.Sprintf("Next challenge at %s", nextAPIPath)}}}
}

// ChallengeSubmitPhoneNumber submits phone number for SMS verification.
func (c *Client) ChallengeSubmitPhoneNumber(challenge map[string]any, phoneNumber string) error {
	apiPath, _ := navigateJSON(challenge, "navigation", "forward").(string)
	if apiPath == "" {
		return &ChallengeError{ClientError: ClientError{Message: "No forward path in phone challenge"}}
	}

	data := map[string]any{
		"phone_number": phoneNumber,
	}

	result, err := c.privateRequest(
		apiPath,
		data,
		map[string]string{},
	)
	if err != nil {
		return err
	}

	challengeType, _ := navigateJSON(result, "challengeType").(string)
	switch challengeType {
	case "VerifySMSCodeFormForSMSCaptcha":
		_ = c.ChallengeCodeOrRaised(c.Username)
		return c.challengeResolveSimple(result, apiPath)
	default:
		return nil
	}
}
