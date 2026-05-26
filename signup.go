package instagrapi

import (
	"fmt"
	"strconv"
	"time"
)

// SignUp registers a new Instagram account.
func (c *Client) SignUp(
	username string,
	password string,
	email string,
	phoneNumber string,
	fullName string,
	year int,
	month int,
	day int,
) (*UserShort, error) {
	if email == "" && phoneNumber == "" {
		return nil, &ClientError{Message: "Use email or phone_number for signup"}
	}

	_, err := c.GetSignupConfig()
	if err != nil {
		return nil, err
	}

	kwargs := map[string]any{
		"username":  username,
		"password":  password,
		"full_name": fullName,
		"year":      year,
		"month":     month,
		"day":       day,
	}

	if email != "" {
		check, err := c.CheckEmail(email)
		if err != nil {
			return nil, err
		}
		valid, _ := check["valid"].(bool)
		if !valid {
			errorTitle, _ := check["error_title"].(string)
			return nil, &EmailInvalidError{ClientError: ClientError{Message: fmt.Sprintf("Email not valid: %s", errorTitle)}}
		}
		available, _ := check["available"].(bool)
		if !available {
			feedbackMsg, _ := check["feedback_message"].(string)
			return nil, &EmailNotAvailableError{ClientError: ClientError{Message: fmt.Sprintf("Email not available: %s", feedbackMsg)}}
		}

		sent, err := c.SendVerifyEmail(email)
		if err != nil {
			return nil, err
		}
		emailSent, _ := sent["email_sent"].(bool)
		if !emailSent {
			return nil, &EmailVerificationSendError{ClientError: ClientError{Message: "Failed to send verification email"}}
		}

		if year > 0 && month > 0 && day > 0 {
			ageCheck, err := c.CheckAgeEligibility(year, month, day)
			if err != nil {
				return nil, err
			}
			eligible, _ := ageCheck["eligible"].(bool)
			if !eligible {
				return nil, &AgeEligibilityError{ClientError: ClientError{Message: "Account not eligible based on age criteria"}}
			}
		}

		var code string
		for attempt := 1; attempt <= 10; attempt++ {
			code = c.ChallengeCodeOrRaised(username)
			if code != "" {
				break
			}
			time.Sleep(time.Duration(int(c.Settings.TimezoneOffset)/100*attempt) * time.Second)
		}

		signupCode, err := c.CheckConfirmationCode(email, code)
		if err != nil {
			return nil, err
		}
		signupCodeStr, _ := signupCode["signup_code"].(string)
		kwargs["email"] = email
		kwargs["signup_code"] = signupCodeStr
	}

	if phoneNumber != "" && email == "" {
		kwargs["phone_number"] = phoneNumber
		check, err := c.CheckPhoneNumber(phoneNumber)
		if err != nil {
			return nil, err
		}
		valid, _ := check["valid"].(bool)
		status, _ := check["status"].(string)
		if status != "ok" && !valid {
			return nil, &ClientError{Message: fmt.Sprintf("Phone number not valid")}
		}

		sms, err := c.SendSignupSMSCode(phoneNumber)
		if err != nil {
			return nil, err
		}
		smsStatus, _ := sms["status"].(string)
		if smsStatus != "ok" {
			return nil, &ClientError{Message: fmt.Sprintf("Error when verify phone number")}
		}

		var code string
		if verificationCode, ok := sms["verification_code"].(string); ok && verificationCode != "" {
			code = verificationCode
		} else {
			for attempt := 1; attempt <= 10; attempt++ {
				code = c.ChallengeCodeOrRaised(username)
				if code != "" {
					break
				}
				time.Sleep(time.Duration(attempt) * time.Second)
			}
		}
		kwargs["phone_code"] = code
	}

	var data map[string]any
	for retries := 0; retries < 3; retries++ {
		data, err = c.AccountsCreate(kwargs)
		if err != nil {
			return nil, err
		}
		message, _ := data["message"].(string)
		if message != "challenge_required" {
			break
		}
		challengeData, ok := data["challenge"].(map[string]any)
		if !ok {
			break
		}
		c.ChallengeFlow(challengeData, phoneNumber, username)
		kwargs["suggestedUsername"] = ""
		kwargs["sn_result"] = "MLA"
	}

	createdUser, ok := data["created_user"].(map[string]any)
	if !ok {
		return nil, &ClientError{Message: "No created user in response"}
	}

	user, err := extractUserShortFromMap(createdUser)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetSignupConfig retrieves signup configuration.
func (c *Client) GetSignupConfig() (map[string]any, error) {
	return c.privateRequest(
		"consent/get_signup_config/",
		nil,
		map[string]string{
			"guid":                  c.UUID,
			"main_account_selected": "false",
		},
	)
}

// CheckEmail checks if an email is valid and available.
func (c *Client) CheckEmail(email string) (map[string]any, error) {
	return c.privateRequest(
		"users/check_email/",
		map[string]any{
			"android_device_id": c.AndroidDeviceID,
			"login_nonce_map":   "{}",
			"login_nonces":      "[]",
			"email":             email,
			"qe_id":             c.generateUUID("", ""),
			"waterfall_id":      c.UUID,
		},
		nil,
	)
}

// CheckPhoneNumber checks if a phone number is valid.
func (c *Client) CheckPhoneNumber(phoneNumber string) (map[string]any, error) {
	return c.privateRequest(
		"accounts/check_phone_number/",
		map[string]any{
			"phone_id":        c.PhoneID,
			"login_nonce_map": "{}",
			"phone_number":    phoneNumber,
			"guid":            c.UUID,
			"device_id":       c.AndroidDeviceID,
			"prefill_shown":   "False",
		},
		nil,
	)
}

// SendSignupSMSCode sends a signup SMS code.
func (c *Client) SendSignupSMSCode(phoneNumber string) (map[string]any, error) {
	return c.privateRequest(
		"accounts/send_signup_sms_code/",
		map[string]any{
			"phone_id":           c.PhoneID,
			"phone_number":       phoneNumber,
			"guid":               c.UUID,
			"device_id":          c.AndroidDeviceID,
			"android_build_type": "release",
			"waterfall_id":       c.UUID,
		},
		nil,
	)
}

// SendVerifyEmail sends a verification email.
func (c *Client) SendVerifyEmail(email string) (map[string]any, error) {
	return c.privateRequest(
		"accounts/send_verify_email/",
		map[string]any{
			"phone_id":          c.PhoneID,
			"device_id":         c.AndroidDeviceID,
			"email":             email,
			"waterfall_id":      c.UUID,
			"auto_confirm_only": "false",
		},
		nil,
	)
}

// CheckConfirmationCode verifies the confirmation code from email.
func (c *Client) CheckConfirmationCode(email string, code string) (map[string]any, error) {
	return c.privateRequest(
		"accounts/check_confirmation_code/",
		map[string]any{
			"code":         code,
			"device_id":    c.AndroidDeviceID,
			"email":        email,
			"waterfall_id": c.UUID,
		},
		nil,
	)
}

// CheckAgeEligibility checks if the user is eligible based on age.
func (c *Client) CheckAgeEligibility(year int, month int, day int) (map[string]any, error) {
	return c.privateRequest(
		"consent/check_age_eligibility/",
		map[string]any{
			"_csrftoken": c.TokenValue(),
			"day":        day,
			"year":       year,
			"month":      month,
		},
		nil,
	)
}

// AccountsCreate creates an account with the given parameters.
func (c *Client) AccountsCreate(kwargs map[string]any) (map[string]any, error) {
	email, _ := kwargs["email"].(string)
	phoneNumber, _ := kwargs["phone_number"].(string)

	if email == "" && phoneNumber == "" {
		return nil, &ClientError{Message: "Use email or phone_number for signup"}
	}

	data := map[string]any{
		"jazoest":                                strconv.Itoa(22341),
		"tos_version":                            "row",
		"suggestedUsername":                      "",
		"sn_result":                              "",
		"do_not_auto_login_if_credentials_match": "false",
		"phone_id":                               c.PhoneID,
		"enc_password":                           c.passwordEncrypt(kwargs["password"].(string)),
		"username":                               fmt.Sprintf("%v", kwargs["username"]),
		"first_name":                             fmt.Sprintf("%v", kwargs["full_name"]),
		"adid":                                   c.AdvertisingID,
		"guid":                                   c.UUID,
		"_uuid":                                  c.UUID,
		"one_tap_opt_in":                         "true",
	}

	if year, ok := kwargs["year"].(int); ok {
		data["year"] = year
	}
	if month, ok := kwargs["month"].(int); ok {
		data["month"] = month
	}
	if day, ok := kwargs["day"].(int); ok {
		data["day"] = day
	}

	for k, v := range kwargs {
		if k == "email" || k == "signup_code" || k == "phone_number" || k == "phone_code" {
			continue
		}
		data[k] = v
	}

	var endpoint string

	if email != "" && phoneNumber == "" {
		endpoint = "accounts/create/"
		data["email"] = email
		signupCode, _ := kwargs["signup_code"].(string)
		data["force_sign_up_code"] = signupCode
	} else {
		endpoint = "accounts/create_validated/"
		data["phone_number"] = phoneNumber
		phoneCode, _ := kwargs["phone_code"].(string)
		data["verification_code"] = phoneCode
		data["force_sign_up_code"] = ""
		data["has_sms_consent"] = "true"
	}

	return c.privateRequest(endpoint, data, nil)
}

// ChallengeFlow handles the signup challenge flow.
func (c *Client) ChallengeFlow(challenge map[string]any, phoneNumber string, username string) bool {
	for i := 0; i < 10; i++ {
		status, _ := challenge["status"].(string)
		if status == "ok" {
			return true
		}

		message, _ := challenge["message"].(string)
		if message == "challenge_required" {
			challenge = c.ChallengeCaptchaFlow(challenge)
			continue
		}

		challengeType, _ := challenge["challengeType"].(string)
		if challengeType == "SubmitPhoneNumberForm" {
			if phoneNumber == "" {
				return false
			}
			challenge = c.ChallengeSubmitPhoneNumberFlow(challenge, phoneNumber)
			continue
		}

		if challengeType == "VerifySMSCodeFormForSMSCaptcha" {
			var securityCode string
			for attempt := 0; attempt < 10; attempt++ {
				securityCode = c.ChallengeCodeOrRaised(username)
				if securityCode != "" {
					break
				}
			}
			if securityCode == "" {
				return false
			}
			challenge = c.ChallengeVerifySMSCaptchaFlow(challenge, securityCode)
			continue
		}

		return false
	}
	return false
}

// ChallengeCaptchaFlow handles captcha during signup.
func (c *Client) ChallengeCaptchaFlow(challenge map[string]any) map[string]any {
	apiPath, _ := challenge["api_path"].(string)
	fields, _ := challenge["fields"].(map[string]any)
	siteKey, _ := fields["sitekey"].(string)

	if siteKey == "" || apiPath == "" {
		return challenge
	}

	captchaDetails := map[string]any{
		"site_key":           siteKey,
		"challenge_type":     "",
		"raw_challenge_json": challenge,
		"page_url":           "https://www.instagram.com/accounts/emailsignup/",
	}

	token, err := c.CaptchaResolve(captchaDetails)
	if err != nil {
		return challenge
	}

	result, _ := c.privateRequest(
		apiPath,
		map[string]any{
			"g-recaptcha-response": token,
		},
		nil,
	)
	return result
}

// ChallengeSubmitPhoneNumberFlow submits phone number during signup challenge flow.
func (c *Client) ChallengeSubmitPhoneNumberFlow(challenge map[string]any, phoneNumber string) map[string]any {
	navigation, _ := challenge["navigation"].(map[string]any)
	apiPath, _ := navigation["forward"].(string)
	if apiPath == "" {
		return challenge
	}

	result, _ := c.privateRequest(
		apiPath,
		map[string]any{
			"phone_number":      phoneNumber,
			"challenge_context": "",
		},
		nil,
	)
	return result
}

// ChallengeVerifySMSCaptchaFlow verifies SMS code during signup challenge flow.
func (c *Client) ChallengeVerifySMSCaptchaFlow(challenge map[string]any, securityCode string) map[string]any {
	navigation, _ := challenge["navigation"].(map[string]any)
	apiPath, _ := navigation["forward"].(string)
	if apiPath == "" {
		return challenge
	}

	result, _ := c.privateRequest(
		apiPath,
		map[string]any{
			"security_code":     securityCode,
			"challenge_context": "",
		},
		nil,
	)
	return result
}
