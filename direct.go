package instagrapi

import (
	"fmt"
	"strconv"
	"time"
)

// DirectThread retrieves a direct message thread by thread ID.
func (c *Client) DirectThread(threadID string) (*DirectThread, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/", threadID),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	thread, err := extractDirectThreadFromMap(result)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

// DirectThreads retrieves direct message threads with pagination.
func (c *Client) DirectThreads(threadViewLimit uint, cursor string) ([]*DirectThread, string, error) {
	params := map[string]string{
		"thread_limit": strconv.Itoa(int(threadViewLimit)),
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	result, err := c.privateRequest(
		"direct_v2/inbox/",
		nil,
		params,
	)
	if err != nil {
		return nil, "", err
	}

	itemsArr := navigateJSON(result, "threads")
	if itemsArr == nil {
		return nil, "", nil
	}

	itemsList, ok := itemsArr.([]any)
	if !ok {
		return nil, "", nil
	}

	var threads []*DirectThread
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		thread, err := extractDirectThreadFromMap(itemMap)
		if err != nil {
			continue
		}
		threads = append(threads, thread)
	}

	nextCursor, _ := navigateJSON(result, "cursor").(string)
	return threads, nextCursor, nil
}

// DirectMessages retrieves messages from a thread with pagination.
func (c *Client) DirectMessages(threadID string, limit uint, cursor string) ([]*DirectMessage, string, error) {
	params := map[string]string{
		"limit": strconv.Itoa(int(limit)),
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	result, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/messages/", threadID),
		nil,
		params,
	)
	if err != nil {
		return nil, "", err
	}

	itemsArr := navigateJSON(result, "items")
	if itemsArr == nil {
		return nil, "", nil
	}

	itemsList, ok := itemsArr.([]any)
	if !ok {
		return nil, "", nil
	}

	var messages []*DirectMessage
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		msg, err := extractDirectMessageFromMap(itemMap)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	nextCursor, _ := navigateJSON(result, "cursor").(string)
	return messages, nextCursor, nil
}

// DirectSend sends a text message to a thread or creates one.
func (c *Client) DirectSend(threadID string, text string) (*DirectMessage, error) {
	data := map[string]any{
		"action":               "send_item",
		"item_type":            "text",
		"text":                 text,
		"_uid":                 strconv.FormatInt(c.UserID, 10),
		"_uuid":                c.UUID,
		"device_id":            c.generateUUID("", ""),
		"offline_threading_id": c.generateUUID("", ""),
	}

	result, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/broadcast/", threadID),
		data,
		map[string]string{
			"v":                  "1",
			"thread_ignition_il": "false",
			"send_silently":      "true",
		},
	)
	if err != nil {
		return nil, err
	}

	item := navigateJSON(result, "item")
	if item == nil {
		return nil, &DirectMessageNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "No item in response"}}}}
	}

	msg, err := extractDirectMessageFromMap(item.(map[string]any))
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// DirectSendToUser sends a message to users by creating or reusing a thread.
func (c *Client) DirectSendToUser(userIDs []int64, text string) (*DirectThread, error) {
	var userIDStrs []string
	for _, id := range userIDs {
		userIDStrs = append(userIDStrs, strconv.FormatInt(id, 10))
	}

	data := map[string]any{
		"action":               "send_item",
		"item_type":            "text",
		"text":                 text,
		"recipient_users":      jsonMarshalCompact(userIDStrs),
		"_uid":                 strconv.FormatInt(c.UserID, 10),
		"_uuid":                c.UUID,
		"device_id":            c.generateUUID("", ""),
		"offline_threading_id": c.generateUUID("", ""),
	}

	result, err := c.privateRequest(
		"direct_v2/broadcast/",
		data,
		map[string]string{
			"v":                  "1",
			"thread_ignition_il": "false",
			"send_silently":      "true",
		},
	)
	if err != nil {
		return nil, err
	}

	thread, err := extractDirectThreadFromMap(result)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

// DirectCreateThread creates a new direct message thread.
func (c *Client) DirectCreateThread(name string, userIDs []int64) (*DirectThread, error) {
	var userIDStrs []string
	for _, id := range userIDs {
		userIDStrs = append(userIDStrs, strconv.FormatInt(id, 10))
	}

	data := map[string]any{
		"action":          "create",
		"name":            name,
		"recipient_users": jsonMarshalCompact(userIDStrs),
		"_uid":            strconv.FormatInt(c.UserID, 10),
		"_uuid":           c.UUID,
		"device_id":       c.generateUUID("", ""),
	}

	result, err := c.privateRequest(
		"direct_v2/threads/",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	thread, err := extractDirectThreadFromMap(result)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

// DirectSearch searches for direct message threads by query.
func (c *Client) DirectSearch(query string) ([]*DirectShortThread, error) {
	result, err := c.privateRequest(
		"direct_v2/search/",
		nil,
		map[string]string{
			"query":          query,
			"search_surface": "user_search_page",
		},
	)
	if err != nil {
		return nil, err
	}

	usersArr := navigateJSON(result, "users")
	if usersArr == nil {
		return nil, nil
	}

	usersList, ok := usersArr.([]any)
	if !ok {
		return nil, nil
	}

	var threads []*DirectShortThread
	for _, u := range usersList {
		userMap, ok := u.(map[string]any)
		if !ok {
			continue
		}
		threadID, _ := navigateJSON(userMap, "thread").(string)
		if threadID == "" {
			continue
		}

		title, _ := navigateJSON(userMap, "thread", "title").(string)
		threadType, _ := navigateJSON(userMap, "thread", "thread_type").(string)

		usersArr2 := navigateJSON(userMap, "thread", "users")
		var users []*UserShort
		if usersArr2 != nil {
			if userList, ok := usersArr2.([]any); ok {
				for _, usr := range userList {
					if usrMap, ok := usr.(map[string]any); ok {
						user, err := extractUserShortFromMap(usrMap)
						if err == nil && user != nil {
							users = append(users, user)
						}
					}
				}
			}
		}

		userList := make([]UserShort, len(users))
		for i, u := range users {
			if u != nil {
				userList[i] = *u
			}
		}
		threads = append(threads, &DirectShortThread{
			ThreadID:   threadID,
			Title:      title,
			ThreadType: threadType,
			Users:      userList,
		})
	}

	return threads, nil
}

// DirectDelete deletes (unsends) a message from a thread.
func (c *Client) DirectDelete(threadID string, itemID string) error {
	data := map[string]any{
		"action": "undeliver",
		"_uid":   strconv.FormatInt(c.UserID, 10),
		"_uuid":  c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/items/%s/", threadID, itemID),
		data,
		nil,
	)
	return err
}

// DirectThreadByParticipants gets a thread by participant user IDs.
func (c *Client) DirectThreadByParticipants(userIDs []int64) (*DirectThread, error) {
	var userIDStrs []string
	for _, id := range userIDs {
		userIDStrs = append(userIDStrs, strconv.FormatInt(id, 10))
	}

	result, err := c.privateRequest(
		"direct_v2/threads/by_users/",
		nil,
		map[string]string{
			"user_ids": jsonMarshalCompact(userIDStrs),
		},
	)
	if err != nil {
		return nil, err
	}

	thread, err := extractDirectThreadFromMap(result)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

// DirectMute mutes a thread.
func (c *Client) DirectMute(threadID string) error {
	data := map[string]any{
		"muted": true,
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/mute/", threadID),
		data,
		nil,
	)
	return err
}

// DirectUnmute unmutes a thread.
func (c *Client) DirectUnmute(threadID string) error {
	data := map[string]any{
		"muted": false,
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/mute/", threadID),
		data,
		nil,
	)
	return err
}

// DirectMarkUnread marks a thread as unread.
func (c *Client) DirectMarkUnread(threadID string) error {
	data := map[string]any{
		"action": "mark_muted",
		"_uid":   strconv.FormatInt(c.UserID, 10),
		"_uuid":  c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/", threadID),
		data,
		nil,
	)
	return err
}

// DirectMediaShare shares a media item in a direct thread.
func (c *Client) DirectMediaShare(threadID string, mediaID any) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.DirectMediaShare(threadID, parsed)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	data := map[string]any{
		"action":               "send_item",
		"item_type":            "media_share",
		"media_id":             pk,
		"_uid":                 strconv.FormatInt(c.UserID, 10),
		"_uuid":                c.UUID,
		"device_id":            c.generateUUID("", ""),
		"offline_threading_id": c.generateUUID("", ""),
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/broadcast/", threadID),
		data,
		map[string]string{
			"v":             "1",
			"send_silently": "true",
		},
	)
	return err
}

// extractDirectThreadFromMap extracts a DirectThread from raw API response data.
func extractDirectThreadFromMap(data map[string]any) (*DirectThread, error) {
	thread := &DirectThread{}

	if pk, ok := data["pk"].(string); ok {
		thread.PK = pk
	}
	if threadID, ok := data["thread_id"].(string); ok {
		thread.ThreadID = threadID
	}
	if title, ok := data["title"].(string); ok {
		thread.Title = title
	}
	if threadType, ok := data["thread_type"].(string); ok {
		thread.ThreadType = threadType
	}
	if viewerID, ok := data["viewer_id"].(string); ok {
		thread.ViewerID = viewerID
	}

	// Extract members
	membersArr := navigateJSON(data, "members")
	if membersArr != nil {
		if membersList, ok := membersArr.([]any); ok {
			for _, m := range membersList {
				if memberMap, ok := m.(map[string]any); ok {
					user, err := extractUserShortFromMap(memberMap)
					if err == nil && user != nil {
						thread.Members = append(thread.Members, ThreadMember{User: *user})
					}
				}
			}
		}
	}

	// Extract items (messages)
	itemsArr := navigateJSON(data, "items")
	if itemsArr != nil {
		if itemList, ok := itemsArr.([]any); ok {
			for _, item := range itemList {
				if itemMap, ok := item.(map[string]any); ok {
					msg, err := extractDirectMessageFromMap(itemMap)
					if err == nil && msg != nil {
						thread.Items = append(thread.Items, *msg)
					}
				}
			}
		}
	}

	return thread, nil
}

// extractDirectMessageFromMap extracts a DirectMessage from raw API response data.
func extractDirectMessageFromMap(data map[string]any) (*DirectMessage, error) {
	msg := &DirectMessage{}

	if itemID, ok := data["item_id"].(string); ok {
		msg.ID = itemID
	} else if clientCtx, ok := data["client_context"].(string); ok {
		msg.ItemID = clientCtx
		msg.ID = clientCtx
	}

	if userID, ok := data["user_id"].(string); ok {
		msg.UserID = userID
	}
	if timestamp, ok := data["timestamp"].(float64); ok {
		msg.Timestamp = int64(timestamp)
	}
	if itemType, ok := data["item_type"].(string); ok {
		msg.ItemType = itemType
	}

	// Extract message content
	if directMsg, ok := data["text"].(string); ok {
		msg.DirectMessage.Text = directMsg
	} else if msgData, ok := data["message"].(map[string]any); ok {
		if text, ok := msgData["text"].(string); ok {
			msg.DirectMessage.Text = text
		}
	}

	return msg, nil
}

// DirectMedia retrieves media shared in direct messages.
func (c *Client) DirectMedia(threadID string, limit uint, cursor string) ([]*DirectMedia, string, error) {
	params := map[string]string{
		"limit": strconv.Itoa(int(limit)),
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	result, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/media/", threadID),
		nil,
		params,
	)
	if err != nil {
		return nil, "", err
	}

	itemsArr := navigateJSON(result, "items")
	if itemsArr == nil {
		return nil, "", nil
	}

	itemsList, ok := itemsArr.([]any)
	if !ok {
		return nil, "", nil
	}

	var medias []*DirectMedia
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		media := &DirectMedia{}
		if id, ok := itemMap["id"].(string); ok {
			media.ID = id
		}
		if videoURL, ok := itemMap["video_url"].(string); ok {
			media.VideoURL = videoURL
		}
		if thumbnailURL, ok := itemMap["thumbnail_url"].(string); ok {
			media.ThumbnailURL = thumbnailURL
		}
		if mediaType, ok := itemMap["media_type"].(float64); ok {
			media.MediaType = int(mediaType)
		}
		medias = append(medias, media)
	}

	nextCursor, _ := navigateJSON(result, "cursor").(string)
	return medias, nextCursor, nil
}

// DirectPendingInbox retrieves pending direct message requests.
func (c *Client) DirectPendingInbox(limit uint, cursor string) ([]*DirectThread, string, error) {
	params := map[string]string{
		"limit": strconv.Itoa(int(limit)),
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	result, err := c.privateRequest(
		"direct_v2/threads/request/",
		nil,
		params,
	)
	if err != nil {
		return nil, "", err
	}

	return c.extractThreadsFromResult(result)
}

// DirectSpamInbox retrieves spam direct messages.
func (c *Client) DirectSpamInbox(limit uint, cursor string) ([]*DirectThread, string, error) {
	params := map[string]string{
		"limit": strconv.Itoa(int(limit)),
	}
	if cursor != "" {
		params["cursor"] = cursor
	}

	result, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/spam/", ""),
		nil,
		params,
	)
	if err != nil {
		return nil, "", err
	}

	return c.extractThreadsFromResult(result)
}

func (c *Client) extractThreadsFromResult(result map[string]any) ([]*DirectThread, string, error) {
	itemsArr := navigateJSON(result, "threads")
	if itemsArr == nil {
		return nil, "", nil
	}

	itemsList, ok := itemsArr.([]any)
	if !ok {
		return nil, "", nil
	}

	var threads []*DirectThread
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		thread, err := extractDirectThreadFromMap(itemMap)
		if err != nil {
			continue
		}
		threads = append(threads, thread)
	}

	nextCursor, _ := navigateJSON(result, "cursor").(string)
	return threads, nextCursor, nil
}

// DirectApprove approves a pending direct message request.
func (c *Client) DirectApprove(threadID string) error {
	data := map[string]any{
		"action": "approve",
		"_uid":   strconv.FormatInt(c.UserID, 10),
		"_uuid":  c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/approve/", threadID),
		data,
		map[string]string{"v": "1"},
	)
	return err
}

// DirectHide hides a direct message thread.
func (c *Client) DirectHide(threadID string) error {
	data := map[string]any{
		"action": "hide",
		"_uid":   strconv.FormatInt(c.UserID, 10),
		"_uuid":  c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/", threadID),
		data,
		nil,
	)
	return err
}

// DirectUpdateThreadTitle updates a thread's title.
func (c *Client) DirectUpdateThreadTitle(threadID string, title string) error {
	data := map[string]any{
		"name":  title,
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/", threadID),
		data,
		nil,
	)
	return err
}

// DirectAddUsers adds users to a thread.
func (c *Client) DirectAddUsers(threadID string, userIDs []int64) error {
	var userIDStrs []string
	for _, id := range userIDs {
		userIDStrs = append(userIDStrs, strconv.FormatInt(id, 10))
	}

	data := map[string]any{
		"action":           "add_users",
		"new_participants": jsonMarshalCompact(userIDStrs),
		"_uid":             strconv.FormatInt(c.UserID, 10),
		"_uuid":            c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/", threadID),
		data,
		nil,
	)
	return err
}

// DirectSeen marks messages in a thread as seen.
func (c *Client) DirectSeen(threadID string, itemID string) error {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)

	data := map[string]any{
		"action":    "mark_seen",
		"timestamp": timestamp,
		"item_id":   itemID,
		"_uid":      strconv.FormatInt(c.UserID, 10),
		"_uuid":     c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("direct_v2/threads/%s/items/", threadID),
		data,
		map[string]string{"v": "1"},
	)
	return err
}
