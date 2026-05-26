package instagrapi

import (
	"strconv"
)

// NotificationMuteAll mutes all notifications for a duration.
func (c *Client) NotificationMuteAll(settingValue string) error {
	data := map[string]any{
		"content_type":  "mute_all",
		"setting_value": settingValue,
		"_uid":          strconv.FormatInt(c.UserID, 10),
		"_uuid":         c.UUID,
	}

	result, err := c.privateRequest(
		"notifications/change_notification_settings/",
		data,
		nil,
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Failed to update notification setting"}
	}
	return nil
}

// NotificationLikes sets likes notification setting.
func (c *Client) NotificationLikes(settingValue string) error {
	return c.setNotificationSetting("likes", settingValue)
}

// NotificationComments sets comments notification setting.
func (c *Client) NotificationComments(settingValue string) error {
	return c.setNotificationSetting("comments", settingValue)
}

// NotificationCommentLikes sets comment likes notification setting.
func (c *Client) NotificationCommentLikes(settingValue string) error {
	return c.setNotificationSetting("comment_likes", settingValue)
}

// NotificationNewFollower sets new follower notification setting.
func (c *Client) NotificationNewFollower(settingValue string) error {
	return c.setNotificationSetting("new_follower", settingValue)
}

// NotificationFollowRequestAccepted sets follow request accepted notification setting.
func (c *Client) NotificationFollowRequestAccepted(settingValue string) error {
	return c.setNotificationSetting("follow_request_accepted", settingValue)
}

// NotificationLogin sets login notification setting.
func (c *Client) NotificationLogin(settingValue string) error {
	return c.setNotificationSetting("login_notification", settingValue)
}

// NotificationDisable disables all notifications.
func (c *Client) NotificationDisable() error {
	settings := []string{
		"likes", "comment_likes", "comments", "new_follower",
		"follow_request_accepted", "login_notification",
		"first_post", "connection_notification", "tagged_in_bio",
		"pending_direct_share", "direct_share_activity",
		"direct_group_requests", "video_call", "rooms",
		"live_broadcast", "felix_upload_result", "view_count",
		"fundraiser_creator", "fundraiser_supporter",
		"notification_reminders", "announcements",
		"report_updated", "like_and_comment_on_photo_user_tagged",
		"user_tagged",
	}

	for _, setting := range settings {
		if err := c.setNotificationSetting(setting, "off"); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) setNotificationSetting(contentType string, settingValue string) error {
	data := map[string]any{
		"content_type":  contentType,
		"setting_value": settingValue,
		"_uid":          strconv.FormatInt(c.UserID, 10),
		"_uuid":         c.UUID,
	}

	result, err := c.privateRequest(
		"notifications/change_notification_settings/",
		data,
		nil,
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Failed to update notification setting"}
	}

	return nil
}
