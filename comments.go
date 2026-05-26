package instagrapi

import (
	"fmt"
	"strconv"
)

// MediaComments retrieves comments on a media item.
func (c *Client) MediaComments(mediaID any, amount uint) ([]*Comment, error) {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaComments(parsed, amount)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid media ID type"}
	}

	var allComments []*Comment
	cursor := ""
	count := 0
	maxCount := int(amount)

	for {
		if maxCount > 0 && count >= maxCount {
			break
		}

		params := map[string]string{
			"can_support_threading": "true",
			"permalink_enabled":     "false",
		}
		if cursor != "" {
			params["cursor"] = cursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("media/%s/comments/", pk),
			nil,
			params,
		)
		if err != nil {
			return allComments, err
		}

		itemsArr := navigateJSON(result, "comments")
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
			comment, err := extractCommentFromMap(itemMap)
			if err != nil {
				continue
			}
			allComments = append(allComments, comment)
			count++

			if maxCount > 0 && count >= maxCount {
				break
			}
		}

		moreAvailable, _ := navigateJSON(result, "has_more").(bool)
		if !moreAvailable || len(itemsList) == 0 {
			break
		}

		cursor, _ = navigateJSON(result, "next_cursor").(string)
		if cursor == "" {
			break
		}
	}

	return allComments, nil
}

// MediaComment posts a comment on media.
func (c *Client) MediaComment(mediaID any, text string) (*Comment, error) {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaComment(parsed, text)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid media ID type"}
	}

	data := map[string]any{
		"feed_item_id": pk,
		"feed_type":    "media",
		"_uid":         strconv.FormatInt(c.UserID, 10),
		"_uuid":        c.UUID,
		"comment_text": text,
	}

	result, err := c.privateRequest(
		fmt.Sprintf("media/%s/comment/", pk),
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	commentData := navigateJSON(result, "comment")
	if commentData == nil {
		return nil, &ClientError{Message: "No comment in response"}
	}

	return extractCommentFromMap(commentData.(map[string]any))
}

// MediaCommentDelete deletes a comment from media.
func (c *Client) MediaCommentDelete(mediaID any, commentID string) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaCommentDelete(parsed, commentID)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/comment/%s/", pk, commentID),
		map[string]any{
			"is_carousel": "false",
			"_uid":        strconv.FormatInt(c.UserID, 10),
			"_uuid":       c.UUID,
		},
		nil,
	)
	return err
}

// MediaCommentLike likes a comment.
func (c *Client) MediaCommentLike(mediaID any, commentID string) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaCommentLike(parsed, commentID)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/comment/%s/like/", pk, commentID),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// MediaCommentUnlike unlikes a comment.
func (c *Client) MediaCommentUnlike(mediaID any, commentID string) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaCommentUnlike(parsed, commentID)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/comment/%s/unlike/", pk, commentID),
		map[string]any{
			"_uid":  strconv.FormatInt(c.UserID, 10),
			"_uuid": c.UUID,
		},
		nil,
	)
	return err
}

// extractCommentFromMap extracts a Comment from raw API response data.
func extractCommentFromMap(data map[string]any) (*Comment, error) {
	comment := &Comment{}

	if pk, ok := data["pk"].(string); ok {
		comment.PK = pk
	} else if pkFloat, ok := data["pk"].(float64); ok {
		comment.PK = strconv.FormatInt(int64(pkFloat), 10)
	}
	if text, ok := data["text"].(string); ok {
		comment.Text = text
	}
	if userID, ok := data["user_id"].(string); ok {
		comment.UserID = userID
	}
	if createdAt, ok := data["created_at"].(float64); ok {
		comment.CreatedAt = int64(createdAt)
	}

	// Extract user info
	if userData, ok := data["user"].(map[string]any); ok {
		user, err := extractUserShortFromMap(userData)
		if err == nil && user != nil {
			comment.User = *user
		}
	}

	return comment, nil
}
