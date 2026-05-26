package instagrapi

import (
	"strconv"
)

// TrackInfoByCanonicalID retrieves track info by music canonical ID.
func (c *Client) TrackInfoByCanonicalID(musicCanonicalID string) (*Track, error) {
	data := map[string]any{
		"tab_type":           "clips",
		"referrer_media_id":  "",
		"_uuid":              c.UUID,
		"music_canonical_id": musicCanonicalID,
	}

	result, err := c.privateRequest(
		"clips/music/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	musicInfo := navigateJSON(result, "metadata", "music_info", "music_asset_info")
	if musicInfo == nil {
		return nil, &TrackNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	}

	musicMap, ok := musicInfo.(map[string]any)
	if !ok {
		return nil, &TrackNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Invalid track data"}}}}
	}

	track := &Track{}
	if id, ok := musicMap["id"].(string); ok {
		track.ID = id
	}
	if title, ok := musicMap["title"].(string); ok {
		track.Title = title
	}
	if displayArtist, ok := musicMap["display_artist"].(string); ok {
		track.DisplayArtist = displayArtist
	}
	if artistID, ok := musicMap["artist_id"].(string); ok {
		track.ArtistID = artistID
	}
	if audioAssetID, ok := musicMap["audio_asset_id"].(string); ok {
		track.AudioAssetID = audioAssetID
	}
	if audioClusterID, ok := musicMap["audio_cluster_id"].(string); ok {
		track.AUDIOClusterID = audioClusterID
	}
	if duration, ok := musicMap["duration_in_ms"].(float64); ok {
		track.Duration = duration / 1000.0 // Convert to seconds
	}
	if albumCoverURL, ok := musicMap["album_cover_art_url"].(string); ok {
		track.AlbumCoverURL = albumCoverURL
	}
	if slug, ok := musicMap["slug"].(string); ok {
		track.Slug = slug
	}

	return track, nil
}

// TrackInfoByID retrieves track info by audio cluster ID.
func (c *Client) TrackInfoByID(trackID string) (map[string]any, error) {
	data := map[string]any{
		"audio_cluster_id":              trackID,
		"original_sound_audio_asset_id": trackID,
	}

	result, err := c.privateRequest(
		"clips/music/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// TrackStreamInfoByID retrieves stream clips pivot page for a track.
func (c *Client) TrackStreamInfoByID(trackID string) (map[string]any, error) {
	data := map[string]any{
		"pivot_page_type": "audio",
		"music_page": map[string]any{
			"tab_type":         "clips",
			"audio_asset_id":   trackID,
			"audio_cluster_id": trackID,
		},
		"_uuid": c.UUID,
	}

	result, err := c.privateRequest(
		"clips/stream_clips_pivot_page/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// MusicSearch searches for music tracks.
func (c *Client) MusicSearch(query string) ([]*Track, error) {
	result, err := c.privateRequest(
		"music/audio_global_search/",
		nil,
		map[string]string{
			"query":             query,
			"browse_session_id": c.generateUUID("", ""),
		},
	)
	if err != nil {
		return nil, err
	}

	itemsArr := navigateJSON(result, "items")
	if itemsArr == nil {
		return nil, nil
	}

	itemsList, ok := itemsArr.([]any)
	if !ok {
		return nil, nil
	}

	var tracks []*Track
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		trackData := navigateJSON(itemMap, "track")
		if trackData == nil {
			continue
		}

		trackMap, ok := trackData.(map[string]any)
		if !ok {
			continue
		}

		track := &Track{}
		if id, ok := trackMap["id"].(string); ok {
			track.ID = id
		}
		if title, ok := trackMap["title"].(string); ok {
			track.Title = title
		}
		if displayArtist, ok := trackMap["display_artist"].(string); ok {
			track.DisplayArtist = displayArtist
		}
		if audioAssetID, ok := trackMap["audio_asset_id"].(string); ok {
			track.AudioAssetID = audioAssetID
		}
		if audioClusterID, ok := trackMap["audio_cluster_id"].(string); ok {
			track.AUDIOClusterID = audioClusterID
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// MusicTrending retrieves trending music.
func (c *Client) MusicTrending(product string) (map[string]any, error) {
	if product == "" {
		product = "feed_post"
	}

	data := map[string]any{
		"product": product,
		"_uuid":   c.UUID,
	}

	result, err := c.privateRequest(
		"music/trending/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// MusicTopTrends retrieves top trending music.
func (c *Client) MusicTopTrends(product string, pageSize int) (map[string]any, error) {
	if product == "" {
		product = "music_in_feed"
	}

	data := map[string]any{
		"product":   product,
		"_uuid":     c.UUID,
		"page_size": strconv.Itoa(pageSize),
	}

	result, err := c.privateRequest(
		"music/top_trends/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// MusicBookmark bookmarks a music track.
func (c *Client) MusicBookmark(originalAudioID string) error {
	data := map[string]any{
		"original_audio_id":      originalAudioID,
		"_uuid":                  c.UUID,
		"surface_requested_from": "audio_aggregation_page",
	}

	result, err := c.privateRequest(
		"music/bookmark_music/",
		data,
		map[string]string{},
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

// MusicVerifyOriginalAudioTitle validates an original audio title.
func (c *Client) MusicVerifyOriginalAudioTitle(audioName string) bool {
	data := map[string]any{
		"original_audio_name": audioName,
		"_uuid":               c.UUID,
	}

	result, err := c.privateRequest(
		"music/verify_original_audio_title/",
		data,
		map[string]string{},
	)
	if err != nil {
		return false
	}

	isValid, _ := result["is_valid"].(bool)
	return isValid
}
