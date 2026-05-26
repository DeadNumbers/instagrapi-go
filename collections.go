package instagrapi

import (
	"fmt"
	"strconv"
)

// Collections retrieves all saved collections.
func (c *Client) Collections() ([]*Collection, error) {
	result, err := c.privateRequest(
		"collections/list/",
		nil,
		map[string]string{
			"collection_types": `["ALL_MEDIA_AUTO_COLLECTION","PRODUCT_AUTO_COLLECTION","MEDIA"]`,
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

	var collections []*Collection
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		collection := &Collection{}
		if idFloat, ok := itemMap["id"].(float64); ok {
			collection.ID = int64(idFloat)
		}
		if name, ok := itemMap["name"].(string); ok {
			collection.Name = name
		}
		if slug, ok := itemMap["slug"].(string); ok {
			collection.Slug = slug
		}
		if type_, ok := itemMap["type"].(string); ok {
			collection.Type = type_
		}
		if def, ok := itemMap["default"].(bool); ok {
			collection.Default = def
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// CollectionMedias retrieves medias in a collection.
func (c *Client) CollectionMedias(collectionPK any, amount uint) ([]*Media, error) {
	var pk string
	switch v := collectionPK.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.CollectionMedias(parsed, amount)
		}
		pk = v
	default:
		return nil, &ClientError{Message: "invalid collection PK type"}
	}

	var endpoint string
	if pk == "liked" {
		endpoint = "feed/liked/"
	} else if _, err := strconv.ParseInt(pk, 10, 64); err == nil {
		endpoint = fmt.Sprintf("feed/collection/%s/", pk)
	} else {
		endpoint = "feed/saved/posts/"
	}

	var allMedias []*Media
	cursor := ""
	count := 0
	maxCount := int(amount)

	for {
		if maxCount > 0 && count >= maxCount {
			break
		}

		params := map[string]string{
			"include_igtv_preview": "false",
		}
		if cursor != "" {
			params["max_id"] = cursor
		}

		result, err := c.privateRequest(
			endpoint,
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
			mediaData := navigateJSON(itemMap, "media")
			if mediaData == nil {
				mediaData = itemMap
			}
			if mediaMap, ok := mediaData.(map[string]any); ok {
				media, err := extractMediaFromMap(mediaMap)
				if err != nil {
					continue
				}
				allMedias = append(allMedias, media)
				count++

				if maxCount > 0 && count >= maxCount {
					break
				}
			}
		}

		moreAvailable, _ := navigateJSON(result, "more_available").(bool)
		if !moreAvailable || len(itemsList) == 0 {
			break
		}

		cursor, _ = navigateJSON(result, "next_max_id", "value").(string)
		if cursor == "" {
			cursor, _ = navigateJSON(result, "next_max_id").(string)
		}
	}

	return allMedias, nil
}

// LikedMedias retrieves medias the user has liked.
func (c *Client) LikedMedias(amount uint) ([]*Media, error) {
	return c.CollectionMedias("liked", amount)
}

// MediaSave saves a media to a collection.
func (c *Client) MediaSave(mediaID any, collectionPK int64) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaSave(parsed, collectionPK)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	data := map[string]any{
		"module":     "feed_timeline",
		"radio_type": "wifi-none",
		"_uid":       strconv.FormatInt(c.UserID, 10),
		"_uuid":      c.UUID,
	}
	if collectionPK > 0 {
		data["added_collection_ids"] = fmt.Sprintf("[%d]", collectionPK)
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/save/", pk),
		data,
		nil,
	)
	return err
}

// MediaUnsave unsaves a media from a collection.
func (c *Client) MediaUnsave(mediaID any, collectionPK int64) error {
	var pk string
	switch v := mediaID.(type) {
	case int64:
		pk = strconv.FormatInt(v, 10)
	case int:
		pk = strconv.Itoa(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return c.MediaUnsave(parsed, collectionPK)
		}
		pk = v
	default:
		return &ClientError{Message: "invalid media ID type"}
	}

	data := map[string]any{
		"module":     "feed_timeline",
		"radio_type": "wifi-none",
		"_uid":       strconv.FormatInt(c.UserID, 10),
		"_uuid":      c.UUID,
	}
	if collectionPK > 0 {
		data["removed_collection_ids"] = fmt.Sprintf("[%d]", collectionPK)
	}

	_, err := c.privateRequest(
		fmt.Sprintf("media/%s/unsave/", pk),
		data,
		nil,
	)
	return err
}
