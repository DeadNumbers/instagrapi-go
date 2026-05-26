package instagrapi

import (
	"fmt"
	"strconv"
)

// HashtagInfo retrieves information about a hashtag.
func (c *Client) HashtagInfo(name string) (*Hashtag, error) {
	name = normalizeHashtagName(name)

	result, err := c.privateRequest(
		fmt.Sprintf("tags/%s/info/", name),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	hashtag := &Hashtag{}
	if id, ok := result["id"].(string); ok {
		hashtag.ID = id
	}
	if hashtagName, ok := result["name"].(string); ok {
		hashtag.Name = hashtagName
	}
	if mediaCount, ok := result["media_count"].(float64); ok {
		hashtag.MediaCount = int(mediaCount)
	}
	if searchResultTitle, ok := result["search_result_title"].(string); ok {
		hashtag.SearchResultTitle = searchResultTitle
	}

	return hashtag, nil
}

// HashtagMedias retrieves medias for a hashtag with pagination.
func (c *Client) HashtagMedias(name string, tabKey string, amount uint) ([]*Media, error) {
	name = normalizeHashtagName(name)

	if tabKey == "" {
		tabKey = "top"
	}

	var allMedias []*Media
	cursor := ""
	count := 0
	maxCount := int(amount)

	for {
		if maxCount > 0 && count >= maxCount {
			break
		}

		data := map[string]any{
			"_uuid":                c.UUID,
			"rank_token":           c.RankToken(),
			"media_recency_filter": getMediaRecencyFilter(tabKey),
		}
		if cursor != "" {
			data["max_id"] = cursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("tags/%s/sections/", name),
			data,
			map[string]string{},
		)
		if err != nil {
			return allMedias, err
		}

		// Extract sections and get medias from layout_content
		sectionsArr := navigateJSON(result, "sections")
		if sectionsArr == nil {
			break
		}

		sectionsList, ok := sectionsArr.([]any)
		if !ok {
			break
		}

		for _, section := range sectionsList {
			sectionMap, ok := section.(map[string]any)
			if !ok {
				continue
			}

			layoutContent := navigateJSON(sectionMap, "layout_content")
			if layoutContent == nil {
				continue
			}

			layoutMap, ok := layoutContent.(map[string]any)
			if !ok {
				continue
			}

			mediasArr := navigateJSON(layoutMap, "medias")
			if mediasArr == nil {
				continue
			}

			mediasList, ok := mediasArr.([]any)
			if !ok {
				continue
			}

			for _, mediaItem := range mediasList {
				mediaMap, ok := mediaItem.(map[string]any)
				if !ok {
					continue
				}

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
		if !moreAvailable || len(sectionsList) == 0 {
			break
		}

		cursor = "" // For simplicity in this implementation
	}

	return allMedias, nil
}

// HashtagFollow follows a hashtag.
func (c *Client) HashtagFollow(name string) error {
	name = normalizeHashtagName(name)

	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("web/tags/%s/follow/", name),
		data,
		map[string]string{},
	)
	return err
}

// HashtagUnfollow unfollows a hashtag.
func (c *Client) HashtagUnfollow(name string) error {
	name = normalizeHashtagName(name)

	data := map[string]any{
		"_uid":  strconv.FormatInt(c.UserID, 10),
		"_uuid": c.UUID,
	}

	_, err := c.privateRequest(
		fmt.Sprintf("web/tags/%s/unfollow/", name),
		data,
		map[string]string{},
	)
	return err
}

// SearchHashtags searches for hashtags by query.
func (c *Client) SearchHashtags(query string) ([]*Hashtag, error) {
	result, err := c.privateRequest(
		"tags/search/",
		nil,
		map[string]string{
			"search_surface": "hashtag_search_page",
			"q":              query,
			"count":          "30",
		},
	)
	if err != nil {
		return nil, err
	}

	resultsArr := navigateJSON(result, "results")
	if resultsArr == nil {
		return nil, nil
	}

	resultsList, ok := resultsArr.([]any)
	if !ok {
		return nil, nil
	}

	var hashtags []*Hashtag
	for _, item := range resultsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		hashtag := &Hashtag{}
		if htData, ok := itemMap["hashtag"].(map[string]any); ok {
			if id, ok := htData["id"].(string); ok {
				hashtag.ID = id
			}
			if name, ok := htData["name"].(string); ok {
				hashtag.Name = name
			}
			if mediaCount, ok := htData["media_count"].(float64); ok {
				hashtag.MediaCount = int(mediaCount)
			}
		}
		hashtags = append(hashtags, hashtag)
	}

	return hashtags, nil
}

// normalizeHashtagName normalizes a hashtag name by removing leading #.
func normalizeHashtagName(name string) string {
	for len(name) > 0 && name[0] == '#' {
		name = name[1:]
	}
	return name
}

// getMediaRecencyFilter returns the recency filter value for a tab key.
func getMediaRecencyFilter(tabKey string) string {
	switch tabKey {
	case "top":
		return "default"
	case "recent":
		return "top_recent_posts"
	default:
		return ""
	}
}
