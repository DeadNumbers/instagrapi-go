package instagrapi

import "strconv"

// ClipPin pins a reel to the Reels tab/profile Reels grid.
func (c *Client) ClipPin(mediaPK int64, revert bool) error {
	name := "pin"
	if revert {
		name = "unpin"
	}

	data := map[string]any{
		"post_id":      strconv.FormatInt(mediaPK, 10),
		"profile_grid": "clips",
	}

	result, err := c.privateRequest(
		"users/"+name+"_timeline_media/",
		data,
		nil,
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Clip pin action failed"}
	}

	return nil
}

// ClipUnpin unpins a reel from the Reels tab/profile Reels grid.
func (c *Client) ClipUnpin(mediaPK int64) error {
	return c.ClipPin(mediaPK, true)
}
