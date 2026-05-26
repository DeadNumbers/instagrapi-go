package instagrapi

import (
	"fmt"
	"strconv"
)

// UserInfoByUsername retrieves user information by username.
func (c *Client) UserInfoByUsername(username string) (*User, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("users/%s/web_profile_info/", username),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	userData := navigateJSON(result, "data", "user")
	if userData == nil {
		return nil, &UserNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	}

	user, err := extractUserFromMap(userData.(map[string]any))
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserInfoByPK retrieves user information by user ID (PK).
func (c *Client) UserInfoByPK(userID int64) (*User, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("users/%d/web_profile_info/", userID),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	userData := navigateJSON(result, "data", "user")
	if userData == nil {
		return nil, &UserNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	}

	user, err := extractUserFromMap(userData.(map[string]any))
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserShortInfoByPK retrieves a minimal user profile by ID.
func (c *Client) UserShortInfoByPK(userID int64) (*UserShort, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("users/%d/info/", userID),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	usersArr := navigateJSON(result, "users")
	if usersArr == nil {
		return nil, &UserNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	}

	usersList, ok := usersArr.([]any)
	if !ok || len(usersList) == 0 {
		return nil, &UserNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	}

	userMap, ok := usersList[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid user data format")
	}

	return extractUserShortFromMap(userMap)
}

// SearchUsers searches for users by query string.
func (c *Client) SearchUsers(query string) ([]*UserShort, error) {
	result, err := c.privateRequest(
		"users/search/",
		nil,
		map[string]string{
			"search_surface": "user_search_page",
			"q":              query,
			"count":          "30",
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

	resultUsers := make([]*UserShort, 0, len(usersList))
	for _, u := range usersList {
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

// SearchFollowers searches followers of a user by query.
func (c *Client) SearchFollowers(userID int64, query string) ([]*UserShort, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("users/%d/subscribers/", userID),
		nil,
		map[string]string{
			"search_surface": "follow_search_page",
			"q":              query,
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

	resultUsers := make([]*UserShort, 0, len(usersList))
	for _, u := range usersList {
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

// SearchFollowing searches following of a user by query.
func (c *Client) SearchFollowing(userID int64, query string) ([]*UserShort, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("users/%d/following/", userID),
		nil,
		map[string]string{
			"search_surface": "follow_search_page",
			"q":              query,
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

	resultUsers := make([]*UserShort, 0, len(usersList))
	for _, u := range usersList {
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

// GetUserFollowers retrieves followers of a user with pagination.
func (c *Client) GetUserFollowers(userID int64, amount int) ([]*UserShort, error) {
	var results []*UserShort
	cursor := ""

	for {
		params := map[string]string{
			"count": strconv.Itoa(amount),
		}
		if cursor != "" {
			params["max_id"] = cursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("users/%d/subscribers/", userID),
			nil,
			params,
		)
		if err != nil {
			return results, err
		}

		usersArr := navigateJSON(result, "users")
		if usersArr == nil {
			break
		}

		usersList, ok := usersArr.([]any)
		if !ok {
			break
		}

		for _, u := range usersList {
			userMap, ok := u.(map[string]any)
			if !ok {
				continue
			}
			user, err := extractUserShortFromMap(userMap)
			if err != nil {
				continue
			}
			results = append(results, user)
		}

		moreAvailable, _ := navigateJSON(result, "has_more").(bool)
		if !moreAvailable || len(usersList) == 0 {
			break
		}

		cursor, _ = navigateJSON(result, "next_max_id", "value").(string)
		if cursor == "" {
			cursor, _ = navigateJSON(result, "next_max_id").(string)
		}
	}

	return results, nil
}

// GetUserFollowing retrieves following of a user with pagination.
func (c *Client) GetUserFollowing(userID int64, amount int) ([]*UserShort, error) {
	var results []*UserShort
	cursor := ""

	for {
		params := map[string]string{
			"count": strconv.Itoa(amount),
		}
		if cursor != "" {
			params["max_id"] = cursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("users/%d/following/", userID),
			nil,
			params,
		)
		if err != nil {
			return results, err
		}

		usersArr := navigateJSON(result, "users")
		if usersArr == nil {
			break
		}

		usersList, ok := usersArr.([]any)
		if !ok {
			break
		}

		for _, u := range usersList {
			userMap, ok := u.(map[string]any)
			if !ok {
				continue
			}
			user, err := extractUserShortFromMap(userMap)
			if err != nil {
				continue
			}
			results = append(results, user)
		}

		moreAvailable, _ := navigateJSON(result, "has_more").(bool)
		if !moreAvailable || len(usersList) == 0 {
			break
		}

		cursor, _ = navigateJSON(result, "next_max_id", "value").(string)
		if cursor == "" {
			cursor, _ = navigateJSON(result, "next_max_id").(string)
		}
	}

	return results, nil
}

// Follow follows a user.
func (c *Client) Follow(userID int64) error {
	_, err := c.privateRequest(
		fmt.Sprintf("users/%d/follow/", userID),
		map[string]any{
			"container": "feed",
			"source":    "all",
			"_uid":      strconv.FormatInt(c.UserID, 10),
			"_uuid":     c.UUID,
		},
		nil,
	)
	return err
}

// Unfollow unfollows a user.
func (c *Client) Unfollow(userID int64) error {
	_, err := c.privateRequest(
		fmt.Sprintf("users/%d/unfollow/", userID),
		map[string]any{
			"container": "feed",
			"source":    "all",
			"_uid":      strconv.FormatInt(c.UserID, 10),
			"_uuid":     c.UUID,
		},
		nil,
	)
	return err
}

// Block blocks a user.
func (c *Client) Block(userID int64) error {
	_, err := c.privateRequest(
		fmt.Sprintf("users/%d/block/", userID),
		map[string]any{
			"source": "organic",
			"_uid":   strconv.FormatInt(c.UserID, 10),
			"_uuid":  c.UUID,
		},
		nil,
	)
	return err
}

// Unblock unblocks a user.
func (c *Client) Unblock(userID int64) error {
	_, err := c.privateRequest(
		fmt.Sprintf("users/%d/unblock/", userID),
		map[string]any{
			"source": "profile",
			"_uid":   strconv.FormatInt(c.UserID, 10),
			"_uuid":  c.UUID,
		},
		nil,
	)
	return err
}

// MutePostsFromFollow mutes posts from a follower.
func (c *Client) MutePostsFromFollow(userID int64) error {
	return c.muteUser(userID, "posts")
}

// UnmutePostsFromFollow unmutes posts from a follower.
func (c *Client) UnmutePostsFromFollow(userID int64) error {
	return c.unmuteUser(userID, "posts")
}

// MuteStoriesFromFollow mutes stories from a follower.
func (c *Client) MuteStoriesFromFollow(userID int64) error {
	return c.muteUser(userID, "stories")
}

// UnmuteStoriesFromFollow unmutes stories from a follower.
func (c *Client) UnmuteStoriesFromFollow(userID int64) error {
	return c.unmuteUser(userID, "stories")
}

func (c *Client) muteUser(userID int64, type_ string) error {
	_, err := c.privateRequest(
		"users/mute/",
		map[string]any{
			"_uid":    strconv.FormatInt(c.UserID, 10),
			"_uuid":   c.UUID,
			"type":    type_,
			"user_id": strconv.FormatInt(userID, 10),
		},
		nil,
	)
	return err
}

func (c *Client) unmuteUser(userID int64, type_ string) error {
	_, err := c.privateRequest(
		"users/unmute/",
		map[string]any{
			"_uid":    strconv.FormatInt(c.UserID, 10),
			"_uuid":   c.UUID,
			"type":    type_,
			"user_id": strconv.FormatInt(userID, 10),
		},
		nil,
	)
	return err
}

// AddCloseFriend adds a user to close friends list.
func (c *Client) AddCloseFriend(userID int64) error {
	_, err := c.privateRequest(
		"friends/add/",
		map[string]any{
			"_uid":    strconv.FormatInt(c.UserID, 10),
			"_uuid":   c.UUID,
			"user_id": strconv.FormatInt(userID, 10),
		},
		nil,
	)
	return err
}

// RemoveCloseFriend removes a user from close friends list.
func (c *Client) RemoveCloseFriend(userID int64) error {
	_, err := c.privateRequest(
		"friends/remove/",
		map[string]any{
			"_uid":    strconv.FormatInt(c.UserID, 10),
			"_uuid":   c.UUID,
			"user_id": strconv.FormatInt(userID, 10),
		},
		nil,
	)
	return err
}

// extractUserFromMap extracts a User from a raw API response map.
func extractUserFromMap(data map[string]any) (*User, error) {
	user := &User{}

	if pk, ok := data["pk"].(string); ok {
		user.PK = pk
	} else if pkFloat, ok := data["pk"].(float64); ok {
		user.PK = strconv.FormatInt(int64(pkFloat), 10)
	}
	if username, ok := data["username"].(string); ok {
		user.Username = username
	}
	if fullName, ok := data["full_name"].(string); ok {
		user.FullName = fullName
	}
	if profilePicURL, ok := data["profile_pic_url"].(string); ok {
		user.ProfilePicURL = profilePicURL
	}
	if isPrivate, ok := data["is_private"].(bool); ok {
		user.IsPrivate = isPrivate
	}
	if isVerified, ok := data["is_verified"].(bool); ok {
		user.IsVerified = isVerified
	}
	if biography, ok := data["biography"].(string); ok {
		user.Biography = biography
	}
	if externalURL, ok := data["external_url"].(string); ok {
		user.ExternalURL = externalURL
	}
	if followerCount, ok := data["follower_count"].(float64); ok {
		user.FollowerCount = int(followerCount)
	}
	if followingCount, ok := data["following_count"].(float64); ok {
		user.FollowingCount = int(followingCount)
	}
	if mediaCount, ok := data["media_count"].(float64); ok {
		user.MediaCount = int(mediaCount)
	}

	return user, nil
}

// extractUserShortFromMap extracts a UserShort from a raw API response map.
func extractUserShortFromMap(data map[string]any) (*UserShort, error) {
	user := &UserShort{}

	if pk, ok := data["pk"].(string); ok {
		user.PK = pk
	} else if pkFloat, ok := data["pk"].(float64); ok {
		user.PK = strconv.FormatInt(int64(pkFloat), 10)
	}
	if username, ok := data["username"].(string); ok {
		user.Username = username
	}
	if fullName, ok := data["full_name"].(string); ok {
		user.FullName = fullName
	}
	if profilePicURL, ok := data["profile_pic_url"].(string); ok {
		user.ProfilePicURL = profilePicURL
	}
	if isPrivate, ok := data["is_private"].(bool); ok {
		user.IsPrivate = isPrivate
	}

	return user, nil
}
