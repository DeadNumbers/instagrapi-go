package instagrapi

import (
	"fmt"
	"strconv"
)

// GetUserStories retrieves stories for a list of user IDs.
func (c *Client) GetUserStories(userIDs []int64) ([]*Story, error) {
	var allStories []*Story

	for _, userID := range userIDs {
		stories, err := c.getUserStoriesV1(userID)
		if err != nil {
			c.Logger.Debug(fmt.Sprintf("GetUserStories for %d: %v", userID, err))
			continue
		}
		allStories = append(allStories, stories...)
	}

	return allStories, nil
}

// GetUserStoryArchive retrieves archived stories for the authenticated user.
func (c *Client) GetUserStoryArchive() ([]*Story, error) {
	var allStories []*Story
	dayCursor := ""

	for {
		result, err := c.privateRequest(
			"users/story_archive/",
			nil,
			map[string]string{
				"phone_id":      c.PhoneID,
				"battery_level": strconv.Itoa(50 + randInt(50)),
				"origin_type":   "1",
				"tag_name":      "",
				"tab_type":      "archives",
			},
		)
		if err != nil {
			break
		}

		reelsArr := navigateJSON(result, "reels")
		if reelsArr == nil {
			break
		}

		reelsMap, ok := reelsArr.(map[string]any)
		if !ok {
			break
		}

		for _, reelKey := range []string{"archive_reel"} {
			reelData := navigateJSON(reelsMap, reelKey)
			if reelData == nil {
				continue
			}

			reelList, ok := reelData.([]any)
			if !ok {
				continue
			}

			for _, item := range reelList {
				itemMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				story, err := extractStoryFromMap(itemMap)
				if err != nil {
					continue
				}
				allStories = append(allStories, story)
			}
		}

		moreAvailable, _ := navigateJSON(result, "more_available").(bool)
		if !moreAvailable || dayCursor == "" {
			break
		}

		dayCursor = "" // For simplicity, stop after first page
	}

	return allStories, nil
}

// StoryDelete deletes a story.
func (c *Client) StoryDelete(storyID any) error {
	var pk string
	switch v := storyID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.StoryDelete(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid story ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/", pk),
		map[string]any{
			"media_type": "v",
			"_uid":       strconv.FormatInt(c.UserID, 10),
			"_uuid":      c.UUID,
		},
		nil,
	)
	return err
}

// StoryView marks a story as seen.
func (c *Client) StoryView(reelID string, itemIDs []string, indices []int) error {
	tapIndices := make([]any, len(indices))
	for i, idx := range indices {
		tapIndices[i] = float64(idx)
	}

	data := map[string]any{
		"supported_capabilities_new": supportedCapabilities,
		"target_user_id":             "",
		"source_of_sequence_id":      "profile",
		"reel_id":                    reelID,
		"view_indices":               tapIndices,
		"item_ids":                   itemIDs,
	}

	if len(indices) > 0 {
		data["view_stories_item_ids"] = itemIDs[:min(len(itemIDs), len(indices))]
	}

	_, err := c.privateRequest(
		"feed/view/",
		data,
		map[string]string{
			"phone_id":                c.PhoneID,
			"battery_level":           strconv.Itoa(50 + randInt(50)),
			"is_charging":             "1",
			"reason":                  "foreground",
			"reel_open_on_reel_press": "true",
		},
	)
	return err
}

// StoryLike likes a story.
func (c *Client) StoryLike(storyID any) error {
	var pk string
	switch v := storyID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.StoryLike(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid story ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/like/", pk),
		map[string]any{
			"container":    "feed",
			"is_reel_item": "true",
			"_uid":         strconv.FormatInt(c.UserID, 10),
			"_uuid":        c.UUID,
		},
		nil,
	)
	return err
}

// StoryUnlike unlikes a story.
func (c *Client) StoryUnlike(storyID any) error {
	var pk string
	switch v := storyID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.StoryUnlike(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid story ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/unlike/", pk),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// StoryViewers retrieves viewers of a story.
func (c *Client) StoryViewers(storyID any) ([]*UserShort, error) {
	var pk string
	switch v := storyID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.StoryViewers(parsed)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid story ID type"}
	}

	result, err := c.privateRequest(
		fmt.Sprintf("media/%s/list_reel_media_viewer/", pk),
		map[string]any{
			"target_user_ids": "[]",
			"_uid":            strconv.FormatInt(c.UserID, 10),
			"_uuid":           c.UUID,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	viewersArr := navigateJSON(result, "users")
	if viewersArr == nil {
		return nil, nil
	}

	viewersList, ok := viewersArr.([]any)
	if !ok {
		return nil, nil
	}

	resultUsers := make([]*UserShort, 0, len(viewersList))
	for _, u := range viewersList {
		userMap, ok := u.(map[string]any)
		if !ok {
			continue
		}
		user, err := extractUserShortFromMap(userMap)
		if err != nil {
			continue
		}
		resultUsers = append(resultUsers, user)
	}

	return resultUsers, nil
}

// getUserStoriesV1 retrieves stories for a single user via v1 API.
func (c *Client) getUserStoriesV1(userID int64) ([]*Story, error) {
	userIDStr := strconv.FormatInt(userID, 10)
	result, err := c.privateRequest(
		fmt.Sprintf("feed/users/%d/story/", userID),
		map[string]any{
			"phone_id":      c.PhoneID,
			"battery_level": strconv.Itoa(50 + randInt(50)),
			"is_charging":   "1",
			"is_dark_mode":  "1",
			"will_sound_on": "1",
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	reelData := navigateJSON(result, "reels", userIDStr)
	if reelData == nil {
		return nil, nil
	}

	reelMap, ok := reelData.(map[string]any)
	if !ok {
		return nil, nil
	}

	var stories []*Story
	itemsArr := navigateJSON(reelMap, "items")
	if itemsArr != nil {
		if itemList, ok := itemsArr.([]any); ok {
			for _, item := range itemList {
				itemMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				story, err := extractStoryFromMap(itemMap)
				if err != nil {
					continue
				}
				stories = append(stories, story)
			}
		}
	}

	return stories, nil
}

// extractStoryFromMap extracts a Story from raw API response data.
func extractStoryFromMap(data map[string]any) (*Story, error) {
	story := &Story{}

	if pk, ok := data["pk"].(string); ok {
		story.PK = pk
	} else if pkFloat, ok := data["pk"].(float64); ok {
		story.PK = strconv.FormatInt(int64(pkFloat), 10)
	}

	if id, ok := data["id"].(string); ok {
		story.ID = id
	}
	if takenAt, ok := data["taken_at"].(float64); ok {
		story.TakenAt = int64(takenAt)
	}
	if mediaType, ok := data["media_type"].(float64); ok {
		story.MediaType = int(mediaType)
	}
	if imageURL, ok := data["image_url"].(string); ok {
		story.ImageURL = imageURL
	}
	if videoURL, ok := data["video_url"].(string); ok {
		story.VideoURL = videoURL
	}
	if videoDuration, ok := data["video_duration"].(float64); ok {
		story.VideoDuration = videoDuration
	}
	if captionText, ok := data["caption_text"].(string); ok {
		story.CaptionText = captionText
	}

	// Extract user info
	if userData, ok := data["user"].(map[string]any); ok {
		user, err := extractUserShortFromMap(userData)
		if err == nil && user != nil {
			story.User = *user
		}
	}

	return story, nil
}
