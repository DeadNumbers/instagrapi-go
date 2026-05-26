package instagrapi

import "strconv"

// StandaloneFundraiserInfoV1 retrieves fundraiser info for a user.
func (c *Client) StandaloneFundraiserInfoV1(userID int64) (map[string]any, error) {
	return c.privateRequest(
		"fundraiser/"+strconv.FormatInt(userID, 10)+"/standalone_fundraiser_info/",
		nil,
		map[string]string{},
	)
}
