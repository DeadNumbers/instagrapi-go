package instagrapi

import "strconv"

// ExplorePageMediaInfo retrieves media information for a media item on the explore page.
func (c *Client) ExplorePageMediaInfo(mediaPK int64) (map[string]any, error) {
	result, err := c.privateRequest(
		"/v1/discover/media_metadata/",
		nil,
		map[string]string{
			"media_id": strconv.FormatInt(mediaPK, 10),
		},
	)
	if err != nil {
		return nil, err
	}

	mediaOrAd, ok := result["media_or_ad"].(map[string]any)
	if !ok {
		return nil, &ClientError{Message: "No media_or_ad in response"}
	}

	return mediaOrAd, nil
}
