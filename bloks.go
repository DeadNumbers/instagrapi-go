package instagrapi

import (
	"strconv"
)

// BloksAsyncAction sends an async action request through Bloks.
func (c *Client) BloksAsyncAction(action string, bloksVersioningID string, extraData map[string]any) (map[string]any, error) {
	data := map[string]any{
		"action":              action,
		"bloks_versioning_id": bloksVersioningID,
		"_uid":                strconv.FormatInt(c.UserID, 10),
		"_uuid":               c.UUID,
	}

	for k, v := range extraData {
		data[k] = v
	}

	result, err := c.privateRequest(
		"bloks/aem/"+action+"/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// BloksTwoStepVerificationEntrypoint initiates the 2FA verification flow.
func (c *Client) BloksTwoStepVerificationEntrypoint(twoStepVerificationContext string) error {
	data := map[string]any{
		"two_step_verification_context": twoStepVerificationContext,
		"_uid":                          strconv.FormatInt(c.UserID, 10),
		"_uuid":                         c.UUID,
	}

	_, err := c.privateRequest(
		"bloks/aem/two_step_verification_entrypoint/",
		data,
		map[string]string{},
	)
	return err
}

// BloksTwoStepVerificationMethodPicker selects the 2FA method.
func (c *Client) BloksTwoStepVerificationMethodPicker(twoStepVerificationContext string) error {
	data := map[string]any{
		"two_step_verification_context": twoStepVerificationContext,
		"_uid":                          strconv.FormatInt(c.UserID, 10),
		"_uuid":                         c.UUID,
	}

	_, err := c.privateRequest(
		"bloks/aem/two_step_verification_method_picker/",
		data,
		map[string]string{},
	)
	return err
}

// BloksTwoStepVerificationSelectMethod selects a specific 2FA method.
func (c *Client) BloksTwoStepVerificationSelectMethod(twoStepVerificationContext string, selectedMethod string) error {
	data := map[string]any{
		"two_step_verification_context": twoStepVerificationContext,
		"selected_method":               selectedMethod,
		"_uid":                          strconv.FormatInt(c.UserID, 10),
		"_uuid":                         c.UUID,
	}

	_, err := c.privateRequest(
		"bloks/aem/two_step_verification_select_method/",
		data,
		map[string]string{},
	)
	return err
}

// BloksTwoStepVerificationEnterTOTPCode enters a TOTP code.
func (c *Client) BloksTwoStepVerificationEnterTOTPCode(twoStepVerificationContext string, totpCode string) error {
	data := map[string]any{
		"two_step_verification_context": twoStepVerificationContext,
		"totp_code":                     totpCode,
		"_uid":                          strconv.FormatInt(c.UserID, 10),
		"_uuid":                         c.UUID,
	}

	_, err := c.privateRequest(
		"bloks/aem/two_step_verification_enter_totp_code/",
		data,
		map[string]string{},
	)
	return err
}

// BloksTwoStepVerificationEnterBackupCode enters a backup code.
func (c *Client) BloksTwoStepVerificationEnterBackupCode(twoStepVerificationContext string, backupCode string) error {
	data := map[string]any{
		"two_step_verification_context": twoStepVerificationContext,
		"backup_code":                   backupCode,
		"_uid":                          strconv.FormatInt(c.UserID, 10),
		"_uuid":                         c.UUID,
	}

	_, err := c.privateRequest(
		"bloks/aem/two_step_verification_enter_backup_code/",
		data,
		map[string]string{},
	)
	return err
}

// BloksTwoStepVerificationVerifyCode verifies the 2FA code.
func (c *Client) BloksTwoStepVerificationVerifyCode(twoStepVerificationContext string, code string, challengeType string) (map[string]any, error) {
	data := map[string]any{
		"two_step_verification_context": twoStepVerificationContext,
		"code":                          code,
		"challenge_type":                challengeType,
		"_uid":                          strconv.FormatInt(c.UserID, 10),
		"_uuid":                         c.UUID,
	}

	result, err := c.privateRequest(
		"bloks/aem/two_step_verification_verify_code/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// BloksCaaLoginSendRequest sends a CAA login request.
func (c *Client) BloksCaaLoginSendRequest(password string, loginAttemptCount int) map[string]any {
	data := map[string]any{
		"username":            c.Username,
		"enc_password":        c.passwordEncrypt(password),
		"login_attempt_count": strconv.Itoa(loginAttemptCount),
		"_uid":                strconv.FormatInt(c.UserID, 10),
		"_uuid":               c.UUID,
	}

	result, err := c.privateRequest(
		"bloks/aem/caa_login_send_request/",
		data,
		map[string]string{},
	)
	if err != nil {
		return map[string]any{}
	}

	return result
}

// BloksExtractTwoStepVerificationContext extracts the two-step verification context from a response.
func (c *Client) BloksExtractTwoStepVerificationContext(response map[string]any) string {
	context, _ := navigateJSON(response, "two_step_verification_context").(string)
	if context == "" {
		return ""
	}

	// Try to find it in nested structures
	if value, ok := response["two_step_verification_context"].(string); ok {
		return value
	}

	return context
}

// BloksApplyLoginResponse applies the login response from Bloks.
func (c *Client) BloksApplyLoginResponse(response map[string]any) bool {
	loginPayload := navigateJSON(response, "login_payload")
	if loginPayload == nil {
		return false
	}

	payloadMap, ok := loginPayload.(map[string]any)
	if !ok {
		return false
	}

	sessionID, _ := payloadMap["sessionid"].(string)
	dsUID, _ := payloadMap["ds_user_id"].(string)

	if sessionID != "" {
		c.SessionID = sessionID
		if c.Settings.Cookies == nil {
			c.Settings.Cookies = make(map[string]string)
		}
		c.Settings.Cookies["sessionid"] = sessionID
	}

	if dsUID != "" {
		c.UserID, _ = strconv.ParseInt(dsUID, 10, 64)
		if c.Settings.Cookies == nil {
			c.Settings.Cookies = make(map[string]string)
		}
		c.Settings.Cookies["ds_user_id"] = dsUID
	}

	return true
}

// BloksChangePassword changes the password via Bloks.
func (c *Client) BloksChangePassword(oldPassword string, newPassword string) error {
	oldEnc := c.passwordEncrypt(oldPassword)
	newEnc := c.passwordEncrypt(newPassword)

	data := map[string]any{
		"enc_old_password":  oldEnc,
		"enc_new_password1": newEnc,
		"enc_new_password2": newEnc,
		"_uid":              strconv.FormatInt(c.UserID, 10),
		"_uuid":             c.UUID,
	}

	result, err := c.privateRequest(
		"bloks/aem/change_password/",
		data,
		map[string]string{},
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Bloks action failed"}
	}
	return nil
}

// BloksFxCalLinkReelsShare shares reels via Bloks.
func (c *Client) BloksFxCalLinkReelsShare(reelIDs []string) error {
	data := map[string]any{
		"reel_ids": jsonMarshalCompact(reelIDs),
		"_uid":     strconv.FormatInt(c.UserID, 10),
		"_uuid":    c.UUID,
	}

	_, err := c.privateRequest(
		"bloks/aem/fx_cal_link_reels_share/",
		data,
		map[string]string{},
	)
	return err
}
