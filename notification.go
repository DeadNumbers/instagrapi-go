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

// NotificationLikeAndCommentOnPhotoUserTagged sets like and comment on photo user tagged notification setting.
func (c *Client) NotificationLikeAndCommentOnPhotoUserTagged(settingValue string) error {
	return c.setNotificationSetting("like_and_comment_on_photo_user_tagged", settingValue)
}

// NotificationUserTagged sets user tagged notification setting.
func (c *Client) NotificationUserTagged(settingValue string) error {
	return c.setNotificationSetting("user_tagged", settingValue)
}

// NotificationFirstPost sets first post notification setting.
func (c *Client) NotificationFirstPost(settingValue string) error {
	return c.setNotificationSetting("first_post", settingValue)
}

// NotificationConnection sets connection notification setting.
func (c *Client) NotificationConnection(settingValue string) error {
	return c.setNotificationSetting("connection_notification", settingValue)
}

// NotificationTaggedInBio sets tagged in bio notification setting.
func (c *Client) NotificationTaggedInBio(settingValue string) error {
	return c.setNotificationSetting("tagged_in_bio", settingValue)
}

// NotificationPendingDirectShare sets pending direct share notification setting.
func (c *Client) NotificationPendingDirectShare(settingValue string) error {
	return c.setNotificationSetting("pending_direct_share", settingValue)
}

// NotificationDirectShareActivity sets direct share activity notification setting.
func (c *Client) NotificationDirectShareActivity(settingValue string) error {
	return c.setNotificationSetting("direct_share_activity", settingValue)
}

// NotificationDirectGroupRequests sets direct group requests notification setting.
func (c *Client) NotificationDirectGroupRequests(settingValue string) error {
	return c.setNotificationSetting("direct_group_requests", settingValue)
}

// NotificationVideoCall sets video call notification setting.
func (c *Client) NotificationVideoCall(settingValue string) error {
	return c.setNotificationSetting("video_call", settingValue)
}

// NotificationRooms sets rooms notification setting.
func (c *Client) NotificationRooms(settingValue string) error {
	return c.setNotificationSetting("rooms", settingValue)
}

// NotificationLiveBroadcast sets live broadcast notification setting.
func (c *Client) NotificationLiveBroadcast(settingValue string) error {
	return c.setNotificationSetting("live_broadcast", settingValue)
}

// NotificationFelixUploadResult sets felix upload result notification setting.
func (c *Client) NotificationFelixUploadResult(settingValue string) error {
	return c.setNotificationSetting("felix_upload_result", settingValue)
}

// NotificationViewCount sets view count notification setting.
func (c *Client) NotificationViewCount(settingValue string) error {
	return c.setNotificationSetting("view_count", settingValue)
}

// NotificationFundraiserCreator sets fundraiser creator notification setting.
func (c *Client) NotificationFundraiserCreator(settingValue string) error {
	return c.setNotificationSetting("fundraiser_creator", settingValue)
}

// NotificationFundraiserSupporter sets fundraiser supporter notification setting.
func (c *Client) NotificationFundraiserSupporter(settingValue string) error {
	return c.setNotificationSetting("fundraiser_supporter", settingValue)
}

// NotificationReminders sets notification reminders setting.
func (c *Client) NotificationReminders(settingValue string) error {
	return c.setNotificationSetting("notification_reminders", settingValue)
}

// NotificationAnnouncements sets announcements notification setting.
func (c *Client) NotificationAnnouncements(settingValue string) error {
	return c.setNotificationSetting("announcements", settingValue)
}

// NotificationReportUpdated sets report updated notification setting.
func (c *Client) NotificationReportUpdated(settingValue string) error {
	return c.setNotificationSetting("report_updated", settingValue)
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
