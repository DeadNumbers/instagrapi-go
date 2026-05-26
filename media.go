package instagrapi

import (
	"fmt"
	"strconv"
)

// MediaInfo retrieves media information by media ID or PK.
func (c *Client) MediaInfo(mediaID any) (*Media, error) {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		// Try to parse as numeric PK first
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaInfo(parsed)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid media ID type"}
	}

	result, err := c.privateRequest(
		fmt.Sprintf("media/%s/info/", pk),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	// Check for items array (some endpoints return wrapped results)
	itemsArr := navigateJSON(result, "items")
	if itemsArr != nil {
		if itemList, ok := itemsArr.([]any); ok && len(itemList) > 0 {
			if itemMap, ok := itemList[0].(map[string]any); ok {
				result = itemMap
			}
		}
	}

	return extractMediaFromMap(result)
}

// MediaDelete deletes a media item.
func (c *Client) MediaDelete(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaDelete(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
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

// MediaEdit edits a media item's caption.
func (c *Client) MediaEdit(mediaID any, caption string) (*Media, error) {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaEdit(parsed, caption)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid media ID type"}
	}

	result, err := c.privateRequest(
		fmt.Sprintf("media/%s/edit_media/", pk),
		map[string]any{
			"caption_text": caption,
			"_uid":         strconv.FormatInt(c.UserID, 10),
			"_uuid":        c.UUID,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return extractMediaFromMap(result)
}

// MediaLike likes a media item.
func (c *Client) MediaLike(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaLike(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/like/", pk),
		map[string]any{
			"container": "feed",
			"source":    "feed",
			"_uid":      strconv.FormatInt(c.UserID, 10),
			"_uuid":     c.UUID,
		},
		nil,
	)
	return err
}

// MediaUnlike unlikes a media item.
func (c *Client) MediaUnlike(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaUnlike(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
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

// MediaArchive archives a media item.
func (c *Client) MediaArchive(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaArchive(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/archive/", pk),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// MediaUnarchive unarchives a media item.
func (c *Client) MediaUnarchive(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaUnarchive(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/unarchive/", pk),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// MediaPin pins a media item to the profile.
func (c *Client) MediaPin(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaPin(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/pin/", pk),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// MediaUnpin unpins a media item from the profile.
func (c *Client) MediaUnpin(mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaUnpin(parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/unpin/", pk),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// MediaOembed retrieves oembed information for a media item.
func (c *Client) MediaOembed(mediaID any) (*MediaOembed, error) {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaOembed(parsed)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid media ID type"}
	}

	result, err := c.privateRequest(
		fmt.Sprintf("media/%s/oembed/", pk),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	oembed := &MediaOembed{}
	if url, ok := result["url"].(string); ok {
		oembed.URL = url
	}
	if title, ok := result["title"].(string); ok {
		oembed.Title = title
	}
	if authorName, ok := result["author_name"].(string); ok {
		oembed.AuthorName = authorName
	}
	if thumbnailURL, ok := result["thumbnail_url"].(string); ok {
		oembed.ThumbnailURL = thumbnailURL
	}

	return oembed, nil
}

// GetUserMedias retrieves a user's media feed with pagination.
func (c *Client) GetUserMedias(userID int64, amount uint) ([]*Media, error) {
	var allMedias []*Media
	endCursor := ""
	count := 0
	maxCount := int(amount)

	for {
		if maxCount > 0 && count >= maxCount {
			break
		}

		params := map[string]string{
			"count": "12", // Instagram's default page size
		}
		if endCursor != "" {
			params["end_cursor"] = endCursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("users/%d/media/", userID),
			nil,
			params,
		)
		if err != nil {
			return allMedias, err
		}

		itemsArr := navigateJSON(result, "items")
		if itemsArr == nil {
			break
		}

		itemsList, ok := itemsArr.([]any)
		if !ok {
			break
		}

		for _, item := range itemsList {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			media, err := extractMediaFromMap(itemMap)
			if err != nil {
				continue
			}
			allMedias = append(allMedias, media)
			count++

			if maxCount > 0 && count >= maxCount {
				break
			}
		}

		// Check for pagination
		pagingInfo := navigateJSON(result, "paging_info")
		if pagingInfo == nil {
			break
		}

		moreAvailable, _ := navigateJSON(pagingInfo, "more_available").(bool)
		if !moreAvailable {
			break
		}

		endCursor, _ = navigateJSON(pagingInfo, "end_cursor").(string)
		if endCursor == "" {
			break
		}
	}

	return allMedias, nil
}

// GetUserClips retrieves a user's clips/reels.
func (c *Client) GetUserClips(userID int64, amount uint) ([]*Media, error) {
	var allMedias []*Media
	endCursor := ""
	count := 0
	maxCount := int(amount)

	for {
		if maxCount > 0 && count >= maxCount {
			break
		}

		params := map[string]string{
			"count": "12",
		}
		if endCursor != "" {
			params["end_cursor"] = endCursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("users/%d/clips/", userID),
			nil,
			params,
		)
		if err != nil {
			return allMedias, err
		}

		itemsArr := navigateJSON(result, "items")
		if itemsArr == nil {
			break
		}

		itemsList, ok := itemsArr.([]any)
		if !ok {
			break
		}

		for _, item := range itemsList {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			media, err := extractMediaFromMap(itemMap)
			if err != nil {
				continue
			}
			allMedias = append(allMedias, media)
			count++

			if maxCount > 0 && count >= maxCount {
				break
			}
		}

		pagingInfo := navigateJSON(result, "paging_info")
		if pagingInfo == nil {
			break
		}

		moreAvailable, _ := navigateJSON(pagingInfo, "more_available").(bool)
		if !moreAvailable {
			break
		}

		endCursor, _ = navigateJSON(pagingInfo, "end_cursor").(string)
		if endCursor == "" {
			break
		}
	}

	return allMedias, nil
}

// MediaCodeFromPK converts a numeric PK to an Instagram shortcode.
func (c *Client) MediaCodeFromPK(pk int64) string {
	codec := &InstagramIdCodec{}
	return codec.Encode(pk)
}

// MediaPKFromCode converts an Instagram shortcode to a numeric PK.
func (c *Client) MediaPKFromCode(code string) (int64, error) {
	codec := &InstagramIdCodec{}
	return codec.Decode(code), nil
}

// extractMediaFromMap extracts a Media from raw API response data.
func extractMediaFromMap(data map[string]any) (*Media, error) {
	media := &Media{}

	if pk, ok := data["pk"].(string); ok {
		media.PK = pk
	} else if pkFloat, ok := data["pk"].(float64); ok {
		media.PK = strconv.FormatInt(int64(pkFloat), 10)
	}

	if id, ok := data["id"].(string); ok {
		media.ID = id
	}
	if code, ok := data["code"].(string); ok {
		media.Code = code
	}
	if takenAt, ok := data["taken_at"].(float64); ok {
		media.TakenAt = int64(takenAt)
	}
	if mediaType, ok := data["media_type"].(float64); ok {
		media.MediaType = int(mediaType)
	}
	if captionText, ok := data["caption_text"].(string); ok {
		media.CaptionText = captionText
	}
	if thumbnailURL, ok := data["thumbnail_url"].(string); ok {
		media.ThumbnailURL = thumbnailURL
	}
	if imageURL, ok := data["image_url"].(string); ok {
		media.ImageURL = imageURL
	}
	if videoURL, ok := data["video_url"].(string); ok {
		media.VideoURL = videoURL
	}
	if commentCount, ok := data["comment_count"].(float64); ok {
		media.CommentCount = int(commentCount)
	}
	if likeCount, ok := data["like_count"].(float64); ok {
		media.LikeCount = int(likeCount)
	}
	if organicTrackingToken, ok := data["organic_tracking_token"].(string); ok {
		media.OrganicTrackingToken = organicTrackingToken
	}

	// Extract user info
	if userData, ok := data["user"].(map[string]any); ok {
		user, err := extractUserFromMap(userData)
		if err == nil && user != nil {
			media.User = *user
		}
	}

	return media, nil
}
