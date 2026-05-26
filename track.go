package instagrapi

import "strconv"

// MusicTrending retrieves trending music candidates.
func (c *Client) MusicTrending(product string) (map[string]any, error) {
	if product == "" {
		product = "feed_post"
	}

	return c.privateRequest(
		"music/trending/",
		map[string]any{
			"product": product,
			"_uuid":   c.UUID,
		},
		nil,
	)
}

// MusicTopTrends retrieves top trending music candidates.
func (c *Client) MusicTopTrends(product string, pageSize int) (map[string]any, error) {
	if product == "" {
		product = "music_in_feed"
	}

	return c.privateRequest(
		"music/top_trends/",
		map[string]any{
			"product":   product,
			"_uuid":     c.UUID,
			"page_size": strconv.Itoa(pageSize),
		},
		nil,
	)
}

// MusicSearchV2 searches music through the search endpoint.
func (c *Client) MusicSearchV2(query string, product string, fromTypeahead bool, searchSessionID string, browseSessionID string) (map[string]any, error) {
	if product == "" {
		product = "music_in_feed"
	}

	if searchSessionID == "" {
		searchSessionID = c.generateUUID("", "")
	}
	if browseSessionID == "" {
		browseSessionID = c.generateUUID("", "")
	}

	return c.privateRequest(
		"music/search_v2/",
		map[string]any{
			"from_typeahead":    strconv.FormatBool(fromTypeahead),
			"search_session_id": searchSessionID,
			"product":           product,
			"q":                 query,
			"_uuid":             c.UUID,
			"browse_session_id": browseSessionID,
		},
		nil,
	)
}

// MusicBookmark bookmarks an original audio track.
func (c *Client) MusicBookmark(originalAudioID string, surfaceRequestedFrom string) error {
	if surfaceRequestedFrom == "" {
		surfaceRequestedFrom = "audio_aggregation_page"
	}

	result, err := c.privateRequest(
		"music/bookmark_music/",
		map[string]any{
			"original_audio_id":      originalAudioID,
			"_uuid":                  c.UUID,
			"surface_requested_from": surfaceRequestedFrom,
		},
		nil,
	)
	if err != nil {
		return err
	}

	success, _ := result["success"].(bool)
	status, _ := result["status"].(string)
	if !success && status != "ok" {
		return &ClientError{Message: "Failed to bookmark music"}
	}

	return nil
}

// MusicClipsAudioBrowser retrieves music candidates for the Reels/Clips camera surface.
func (c *Client) MusicClipsAudioBrowser(product string, browseSessionID string) (map[string]any, error) {
	if product == "" {
		product = "story_camera_clips_v2"
	}

	if browseSessionID == "" {
		browseSessionID = c.generateUUID("", "")
	}

	return c.privateRequest(
		"music/clips_audio_browser/",
		map[string]any{
			"product":           product,
			"_uuid":             c.UUID,
			"browse_session_id": browseSessionID,
		},
		nil,
	)
}

// MusicVerifyOriginalAudioTitle validates an original audio title for Reels publishing.
func (c *Client) MusicVerifyOriginalAudioTitle(originalAudioName string) bool {
	result, err := c.privateRequest(
		"music/verify_original_audio_title/",
		map[string]any{
			"original_audio_name": originalAudioName,
			"_uuid":               c.UUID,
		},
		nil,
	)
	if err != nil {
		return false
	}

	isValid, _ := result["is_valid"].(bool)
	return isValid
}

// TrackInfoByID retrieves track information by ID.
func (c *Client) TrackInfoByID(trackID string, maxID string) (map[string]any, error) {
	data := map[string]any{
		"audio_cluster_id":              trackID,
		"original_sound_audio_asset_id": trackID,
	}

	if maxID != "" {
		data["max_id"] = maxID
	}

	return c.privateRequest("clips/music/", data, nil)
}
