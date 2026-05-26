package instagrapi

import "strconv"

// FeaturedAccountsV1 retrieves featured accounts for a target user.
func (c *Client) FeaturedAccountsV1(targetUserID int64) (map[string]any, error) {
	return c.privateRequest(
		"multiple_accounts/get_featured_accounts/",
		nil,
		map[string]string{
			"target_user_id": strconv.FormatInt(targetUserID, 10),
		},
	)
}

// GetAccountFamilyV1 retrieves the account family information.
func (c *Client) GetAccountFamilyV1() (map[string]any, error) {
	return c.privateRequest(
		"multiple_accounts/get_account_family/",
		nil,
		map[string]string{},
	)
}
