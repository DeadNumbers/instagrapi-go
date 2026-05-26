package instagrapi

import (
	"fmt"
	"strconv"
)

// UserHighlights retrieves highlights for a user.
func (c *Client) UserHighlights(userID int64) ([]*Highlight, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("highlights/%d/highlights_tray/", userID),
		map[string]any{
			"phone_id":      c.PhoneID,
			"battery_level": float64(50 + randInt(50)),
			"is_charging":   float64(randInt(2)),
			"is_dark_mode":  float64(randInt(2)),
			"will_sound_on": float64(randInt(2)),
		},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	trayArr := navigateJSON(result, "tray")
	if trayArr == nil {
		return nil, nil
	}

	trayList, ok := trayArr.([]any)
	if !ok {
		return nil, nil
	}

	var highlights []*Highlight
	for _, item := range trayList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		highlight := &Highlight{}
		if pk, ok := itemMap["pk"].(string); ok {
			highlight.PK = pk
		} else if pkFloat, ok := itemMap["pk"].(float64); ok {
			highlight.PK = strconv.FormatInt(int64(pkFloat), 10)
		}
		if id, ok := itemMap["id"].(string); ok {
			highlight.ID = id
		}
		if title, ok := itemMap["title"].(string); ok {
			highlight.Title = title
		}
		if mediaCount, ok := itemMap["media_count"].(float64); ok {
			highlight.MediaCount = int(mediaCount)
		}

		// Extract cover media
		if coverMedia, ok := itemMap["cover_media"].(map[string]any); ok {
			highlight.CoverMedia.ThumbnailURL, _ = coverMedia["thumbnail_url"].(string)
		}

		// Extract user info
		if userData, ok := itemMap["user"].(map[string]any); ok {
			user, err := extractUserShortFromMap(userData)
			if err == nil && user != nil {
				highlight.User = *user
			}
		}

		highlights = append(highlights, highlight)
	}

	return highlights, nil
}

// HighlightInfo retrieves a single highlight by PK.
func (c *Client) HighlightInfo(highlightPK string) (*Highlight, error) {
	data := map[string]any{
		"source": "profile",
		"_uid":   strconv.FormatInt(c.UserID, 10),
		"_uuid":  c.UUID,
	}

	result, err := c.privateRequest(
		"feed/reels_media/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	reelKey := "highlight:" + highlightPK
	reelData := navigateJSON(result, "reels", reelKey)
	if reelData == nil {
		return nil, &HighlightNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	}

	highlight := &Highlight{}
	reelMap, ok := reelData.(map[string]any)
	if !ok {
		return nil, &HighlightNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Invalid highlight data"}}}}
	}

	if pk, ok := reelMap["pk"].(string); ok {
		highlight.PK = pk
	}
	if id, ok := reelMap["id"].(string); ok {
		highlight.ID = id
	}
	if title, ok := reelMap["title"].(string); ok {
		highlight.Title = title
	}

	return highlight, nil
}

// HighlightCreate creates a new highlight from story IDs.
func (c *Client) HighlightCreate(title string, storyIDs []string) (*Highlight, error) {
	data := map[string]any{
		"source": "self_profile",
		"_uid":   strconv.FormatInt(c.UserID, 10),
		"_uuid":  c.UUID,
	}

	result, err := c.privateRequest(
		"highlights/create_reel/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	highlight := &Highlight{}
	if reelData, ok := result["reel"].(map[string]any); ok {
		if pk, ok := reelData["pk"].(string); ok {
			highlight.PK = pk
		}
		if id, ok := reelData["id"].(string); ok {
			highlight.ID = id
		}
		if t, ok := reelData["title"].(string); ok {
			highlight.Title = t
		}
	}

	return highlight, nil
}

// HighlightDelete deletes a highlight.
func (c *Client) HighlightDelete(highlightPK string) error {
	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("highlights/highlight:%s/delete_reel/", highlightPK),
		data,
		map[string]string{},
	)
	return err
}

// HighlightEditTitle edits a highlight's title.
func (c *Client) HighlightEditTitle(highlightPK string, title string) error {
	data := map[string]any{
		"title": title,
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("highlights/highlight:%s/edit_reel/", highlightPK),
		data,
		map[string]string{},
	)
	return err
}

// HighlightAddStories adds stories to a highlight.
func (c *Client) HighlightAddStories(highlightPK string, storyIDs []string) error {
	data := map[string]any{
		"_uid":            strconv.FormatInt(c.UserID, 10),
		"_uuid":           c.UUID,
		"added_media_ids": jsonMarshalCompact(storyIDs),
	}

	_, err := c.privateRequest(
		fmt.Sprintf("highlights/highlight:%s/edit_reel/", highlightPK),
		data,
		map[string]string{},
	)
	return err
}

// HighlightRemoveStories removes stories from a highlight.
func (c *Client) HighlightRemoveStories(highlightPK string, storyIDs []string) error {
	data := map[string]any{
		"_uid":              strconv.FormatInt(c.UserID, 10),
		"_uuid":             c.UUID,
		"removed_media_ids": jsonMarshalCompact(storyIDs),
	}

	_, err := c.privateRequest(
		fmt.Sprintf("highlights/highlight:%s/edit_reel/", highlightPK),
		data,
		map[string]string{},
	)
	return err
}
