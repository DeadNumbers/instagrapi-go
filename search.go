package instagrapi

import (
	"fmt"
	"strconv"
)

// TopSearch performs a top search across Instagram.
func (c *Client) TopSearch(query string) (map[string]any, error) {
	result, err := c.privateRequest(
		"web/search/topsearch/",
		nil,
		map[string]string{
			"search_surface": "web_top_search",
			"context":        "blended",
			"include_reel":   "true",
			"query":          query,
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FbSearchTopSearchFlat performs a flat top search.
func (c *Client) FbSearchTopSearchFlat(query string) ([]any, error) {
	result, err := c.privateRequest(
		"fbsearch/topsearch_flat/",
		nil,
		map[string]string{
			"search_surface": "top_search_page",
			"context":        "blended",
			"count":          "30",
			"query":          query,
		},
	)
	if err != nil {
		return nil, err
	}

	list := navigateJSON(result, "list")
	if list == nil {
		return nil, nil
	}

	listArr, ok := list.([]any)
	if !ok {
		return nil, nil
	}

	return listArr, nil
}

// FbSearchPlaces searches for places by query and location.
func (c *Client) FbSearchPlaces(query string, lat float64, lng float64) ([]*Location, error) {
	result, err := c.privateRequest(
		"fbsearch/places/",
		nil,
		map[string]string{
			"search_surface":  "places_search_page",
			"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
			"lat":             fmt.Sprintf("%f", lat),
			"lng":             fmt.Sprintf("%f", lng),
			"count":           "30",
			"query":           query,
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

	var locations []*Location
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		locationData := navigateJSON(itemMap, "location")
		if locationData == nil {
			locationData = itemMap
		}

		loc := &Location{}
		if locMap, ok := locationData.(map[string]any); ok {
			if pkFloat, ok := locMap["pk"].(float64); ok {
				loc.PK = int(pkFloat)
			}
			if name, ok := locMap["name"].(string); ok {
				loc.Name = name
			}
			if address, ok := locMap["address"].(string); ok {
				loc.Address = address
			}
			if latVal, ok := locMap["lat"].(float64); ok {
				loc.Lat = latVal
			}
			if lngVal, ok := locMap["lng"].(float64); ok {
				loc.Lng = lngVal
			}
			if externalID, ok := locMap["external_id"].(string); ok {
				loc.ExternalID = externalID
			}
			if externalSource, ok := locMap["external_id_source"].(string); ok {
				loc.ExternalIDSource = externalSource
			}
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// FbSearchAccountsV2 searches accounts via v2 SERP endpoint.
func (c *Client) FbSearchAccountsV2(query string, pageToken string) (map[string]any, error) {
	params := map[string]string{
		"search_surface":  "account_serp",
		"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
		"query":           query,
	}
	if pageToken != "" {
		params["page_token"] = pageToken
	}

	result, err := c.privateRequest(
		"fbsearch/account_serp/",
		nil,
		params,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FbSearchReelsV2 searches reels via v2 SERP endpoint.
func (c *Client) FbSearchReelsV2(query string, reelsMaxID string, rankToken string) (map[string]any, error) {
	params := map[string]string{
		"search_surface":  "clips_search_page",
		"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
		"query":           query,
	}
	if reelsMaxID != "" {
		params["reels_max_id"] = reelsMaxID
	}
	if rankToken != "" {
		params["rank_token"] = rankToken
	}

	result, err := c.privateRequest(
		"fbsearch/reels_serp/",
		nil,
		params,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FbSearchTopSearchV2 performs blended search via v2 SERP endpoint.
func (c *Client) FbSearchTopSearchV2(query string, nextMaxID string, reelsMaxID string, rankToken string) (map[string]any, error) {
	params := map[string]string{
		"search_surface":  "top_serp",
		"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
		"query":           query,
		"rank_token":      c.RankToken(),
	}
	if nextMaxID != "" {
		params["next_max_id"] = nextMaxID
	}
	if rankToken != "" {
		params["rank_token"] = rankToken
	}
	if reelsMaxID != "" {
		params["reels_max_id"] = reelsMaxID
	}

	result, err := c.privateRequest(
		"fbsearch/top_serp/",
		nil,
		params,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FbSearchRecent retrieves recently searched results.
func (c *Client) FbSearchRecent() ([]any, error) {
	result, err := c.privateRequest(
		"fbsearch/recent_searches/",
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	recentArr := navigateJSON(result, "recent")
	if recentArr == nil {
		return nil, nil
	}

	recentList, ok := recentArr.([]any)
	if !ok {
		return nil, nil
	}

	var results []any
	for _, item := range recentList {
		_, ok := item.(map[string]any)
		if !ok {
			continue
		}
		results = append(results, item)
	}

	return results, nil
}

// ExplorePage retrieves the explore/discover page.
func (c *Client) ExplorePage() (map[string]any, error) {
	result, err := c.privateRequest(
		"discover/topical_explore/",
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ReportExploreMedia reports a media item on the explore page.
func (c *Client) ReportExploreMedia(mediaPK int64) error {
	result, err := c.privateRequest(
		"discover/explore_report/",
		nil,
		map[string]string{
			"m_pk": strconv.FormatInt(mediaPK, 10),
		},
	)
	if err != nil {
		return err
	}

	status, _ := result["explore_report_status"].(string)
	if status != "OK" {
		return &ClientError{Message: "Failed to report explore media"}
	}

	return nil
}

// Reels retrieves connected reels.
func (c *Client) Reels(amount uint) ([]*Media, error) {
	return c.reelsTimelineMedia("reels", amount)
}

// ExploreReels retrieves discover reels.
func (c *Client) ExploreReels(amount uint) ([]*Media, error) {
	return c.reelsTimelineMedia("explore_reels", amount)
}

// FriendsReels retrieves friends tab reels.
func (c *Client) FriendsReels(amount uint) ([]*Media, error) {
	return c.reelsTimelineMedia("friends_reels", amount)
}

func (c *Client) reelsTimelineMedia(collectionPK string, amount uint) ([]*Media, error) {
	endpointMap := map[string]string{
		"reels":         "clips/connected/",
		"explore_reels": "clips/discover/",
		"friends_reels": "clips/discover/social/",
	}

	endpoint, ok := endpointMap[collectionPK]
	if !ok {
		return nil, &ClientError{Message: fmt.Sprintf("Unsupported reels timeline collection: %s", collectionPK)}
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
			"max_id": cursor,
		}

		result, err := c.privateRequest(
			endpoint,
			map[string]any{},
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

		pagingInfo := navigateJSON(result, "paging_info")
		if pagingInfo == nil {
			break
		}

		moreAvailable, _ := navigateJSON(pagingInfo, "more_available").(bool)
		if !moreAvailable || len(itemsList) == 0 {
			break
		}

		cursor, _ = navigateJSON(pagingInfo, "max_id").(string)
		if cursor == "" {
			break
		}
	}

	return allMedias, nil
}
