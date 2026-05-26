package instagrapi

import (
	"fmt"
	"strconv"
)

// AccountInfo retrieves the authenticated user's account info.
func (c *Client) AccountInfo() (*Account, error) {
	result, err := c.privateRequest(
		"accounts/current_user/?edit=true",
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	userData := navigateJSON(result, "user")
	if userData == nil {
		return nil, &ClientError{Message: "No user data in response"}
	}

	return c.extractAccountFromMap(userData.(map[string]any))
}

// AccountEdit edits the authenticated user's profile.
func (c *Client) AccountEdit(fullName string, biography string, externalURL string) (*Account, error) {
	data := map[string]any{
		"_uid":         strconv.FormatInt(c.UserID, 10),
		"_uuid":        c.UUID,
		"full_name":    fullName,
		"biography":    biography,
		"external_url": externalURL,
	}

	result, err := c.privateRequest(
		"accounts/edit_profile/",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	userData := navigateJSON(result, "user")
	if userData == nil {
		return nil, &ClientError{Message: "No user data in response"}
	}

	return c.extractAccountFromMap(userData.(map[string]any))
}

// AccountSetPrivate sets the account to private.
func (c *Client) AccountSetPrivate() error {
	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		"accounts/set_private/",
		data,
		nil,
	)
	return err
}

// AccountSetPublic sets the account to public.
func (c *Client) AccountSetPublic() error {
	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		"accounts/set_public/",
		data,
		nil,
	)
	return err
}

// AccountChangePassword changes the account password.
func (c *Client) AccountChangePassword(oldPassword string, newPassword string) error {
	encOldPwd := c.passwordEncrypt(oldPassword)
	encNewPwd := c.passwordEncrypt(newPassword)

	data := map[string]any{
		"enc_old_password":  encOldPwd,
		"enc_new_password1": encNewPwd,
		"enc_new_password2": encNewPwd,
		"_uid":              strconv.FormatInt(c.UserID, 10),
		"_uuid":             c.UUID,
	}

	_, err := c.privateRequest(
		"accounts/change_password/",
		data,
		nil,
	)
	return err
}

// AccountSecurityInfo retrieves account security information.
func (c *Client) AccountSecurityInfo() (*AccountSecurityInfo, error) {
	result, err := c.privateRequest(
		"accounts/account_security_info/",
		c.withDefaultData(map[string]any{}),
		nil,
	)
	if err != nil {
		return nil, err
	}

	info := &AccountSecurityInfo{}
	if isPhoneConfirmed, ok := result["is_phone_confirmed"].(bool); ok {
		info.IsPhoneConfirmed = isPhoneConfirmed
	}
	if isTwoFactorEnabled, ok := result["is_two_factor_enabled"].(bool); ok {
		info.IsTwoFactorEnabled = isTwoFactorEnabled
	}
	if isTOTPTwoFactorEnabled, ok := result["is_totp_two_factor_enabled"].(bool); ok {
		info.IsTOTPTwoFactorEnabled = isTOTPTwoFactorEnabled
	}
	if backupCodes, ok := result["backup_codes"].([]any); ok {
		for _, code := range backupCodes {
			if s, ok := code.(string); ok {
				info.BackupCodes = append(info.BackupCodes, s)
			}
		}
	}

	return info, nil
}

// AccountSendPasswordReset sends a password reset link to the account email.
func (c *Client) AccountSendPasswordReset(identifier string) error {
	csrfToken := c.TokenValue()

	result, err := c.publicRequest("POST", "https://www.instagram.com/accounts/account_recovery_send_ajax/", map[string]any{
		"email_or_username": identifier,
	}, map[string]string{
		"x-requested-with": "XMLHttpRequest",
		"x-csrftoken":      csrfToken,
	})

	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: fmt.Sprintf("Password reset failed: %v", result)}
	}

	return nil
}

// AccountSetBiography sets the account biography.
func (c *Client) AccountSetBiography(biography string) error {
	data := map[string]any{
		"logged_in_uids": jsonMarshalCompact([]string{strconv.FormatInt(c.UserID, 10)}),
		"raw_text":       biography,
	}

	_, err := c.privateRequest(
		"accounts/set_biography/",
		c.withDefaultData(data),
		nil,
	)
	return err
}

// AccountSetExternalURL sets the external URL in the bio.
func (c *Client) AccountSetExternalURL(url string) error {
	data := map[string]any{
		"updated_links": jsonMarshalCompact([]map[string]string{{
			"url":       url,
			"title":     "",
			"link_type": "external",
		}}),
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		"accounts/update_bio_links/",
		map[string]any{
			"signed_body": "SIGNATURE." + urlEncode(jsonMarshalCompact(data)),
		},
		nil,
	)
	return err
}

// withDefaultData adds default authentication data to a map.
func (c *Client) withDefaultData(data map[string]any) map[string]any {
	if data["_uid"] == "" {
		data["_uid"] = strconv.FormatInt(c.UserID, 10)
	}
	if data["_uuid"] == "" {
		data["_uuid"] = c.UUID
	}
	return data
}

// extractAccountFromMap extracts an Account from raw API response data.
func (c *Client) extractAccountFromMap(data map[string]any) (*Account, error) {
	account := &Account{}

	if pk, ok := data["pk"].(string); ok {
		account.PK = pk
	} else if pkFloat, ok := data["pk"].(float64); ok {
		account.PK = strconv.FormatInt(int64(pkFloat), 10)
	}
	if username, ok := data["username"].(string); ok {
		account.Username = username
	}
	if fullName, ok := data["full_name"].(string); ok {
		account.FullName = fullName
	}
	if biography, ok := data["biography"].(string); ok {
		account.Biography = biography
	}
	if profilePicURL, ok := data["profile_pic_url"].(string); ok {
		account.ProfilePicURL = profilePicURL
	}
	if isPrivate, ok := data["is_private"].(bool); ok {
		account.IsPrivate = isPrivate
	}
	if isVerified, ok := data["is_verified"].(bool); ok {
		account.IsVerified = isVerified
	}
	if email, ok := data["email"].(string); ok {
		account.Email = email
	}
	if phoneNumber, ok := data["phone_number"].(string); ok {
		account.Phone_number = phoneNumber
	}
	if mediaCount, ok := data["media_count"].(float64); ok {
		account.MediaCount = int(mediaCount)
	}
	if followerCount, ok := data["follower_count"].(float64); ok {
		account.FollowerCount = int(followerCount)
	}
	if followingCount, ok := data["following_count"].(float64); ok {
		account.FollowingCount = int(followingCount)
	}

	return account, nil
}
