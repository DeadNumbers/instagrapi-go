package instagrapi

import (
	"fmt"
)

// LocationSearch searches for locations by latitude and longitude.
func (c *Client) LocationSearch(lat, lng float64) ([]*Location, error) {
	result, err := c.privateRequest(
		"location_search/",
		nil,
		map[string]string{
			"latitude":  fmt.Sprintf("%f", lat),
			"longitude": fmt.Sprintf("%f", lng),
		},
	)
	if err != nil {
		return nil, err
	}

	venuesArr := navigateJSON(result, "venues")
	if venuesArr == nil {
		return nil, nil
	}

	venuesList, ok := venuesArr.([]any)
	if !ok {
		return nil, nil
	}

	var locations []*Location
	for _, venue := range venuesList {
		venueMap, ok := venue.(map[string]any)
		if !ok {
			continue
		}
		location := &Location{}
		if pkFloat, ok := venueMap["pk"].(float64); ok {
			location.PK = int(pkFloat)
		}
		if name, ok := venueMap["name"].(string); ok {
			location.Name = name
		}
		if address, ok := venueMap["address"].(string); ok {
			location.Address = address
		}
		if city, ok := venueMap["city"].(string); ok {
			location.City = city
		}
		if latVal, ok := venueMap["lat"].(float64); ok {
			location.Lat = latVal
		} else {
			location.Lat = lat
		}
		if lngVal, ok := venueMap["lng"].(float64); ok {
			location.Lng = lngVal
		} else {
			location.Lng = lng
		}
		if externalID, ok := venueMap["external_id"].(string); ok {
			location.ExternalID = externalID
		}
		if externalSource, ok := venueMap["external_id_source"].(string); ok {
			location.ExternalIDSource = externalSource
		}
		locations = append(locations, location)
	}

	return locations, nil
}

// LocationInfo retrieves information about a location by PK.
func (c *Client) LocationInfo(locationPK int) (*Location, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("locations/%d/location_info/", locationPK),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	location := &Location{}
	if name, ok := result["name"].(string); !ok || name == "" {
		return nil, &LocationNotFound{NotFoundError: NotFoundError{PrivateError: PrivateError{ClientError: ClientError{Reason: "Not found"}}}}
	} else {
		location.Name = name
	}
	if address, ok := result["address"].(string); ok {
		location.Address = address
	}
	if city, ok := result["city"].(string); ok {
		location.City = city
	}
	if phoneNumber, ok := result["phone_number"].(string); ok {
		location.PhoneNumber = phoneNumber
	}
	if latVal, ok := result["lat"].(float64); ok {
		location.Lat = latVal
	}
	if lngVal, ok := result["lng"].(float64); ok {
		location.Lng = lngVal
	}
	if externalID, ok := result["external_id"].(string); ok {
		location.ExternalID = externalID
	}
	if externalSource, ok := result["external_id_source"].(string); ok {
		location.ExternalIDSource = externalSource
	}

	return location, nil
}

// LocationMediasTop retrieves top medias for a location.
func (c *Client) LocationMediasTop(locationPK int, amount uint) ([]*Media, error) {
	return c.locationMedias(locationPK, "ranked", amount)
}

// LocationMediasRecent retrieves recent medias for a location.
func (c *Client) LocationMediasRecent(locationPK int, amount uint) ([]*Media, error) {
	return c.locationMedias(locationPK, "recent", amount)
}

func (c *Client) locationMedias(locationPK int, tabKey string, amount uint) ([]*Media, error) {
	var allMedias []*Media
	cursor := ""
	count := 0
	maxCount := int(amount)

	for {
		if maxCount > 0 && count >= maxCount {
			break
		}

		data := map[string]any{
			"_uuid":      c.UUID,
			"session_id": c.generateUUID("", ""),
			"tab":        tabKey,
		}
		if cursor != "" {
			data["max_id"] = cursor
		}

		result, err := c.privateRequest(
			fmt.Sprintf("locations/%d/sections/", locationPK),
			data,
			map[string]string{},
		)
		if err != nil {
			return allMedias, err
		}

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

		cursor = "" // Simplified cursor handling
	}

	return allMedias, nil
}

// LocationGuides retrieves guides for a location.
func (c *Client) LocationGuides(locationPK int) ([]*Guide, error) {
	result, err := c.privateRequest(
		fmt.Sprintf("guides/location/%d/", locationPK),
		nil,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	guidesArr := navigateJSON(result, "guides")
	if guidesArr == nil {
		return nil, nil
	}

	guidesList, ok := guidesArr.([]any)
	if !ok {
		return nil, nil
	}

	var guides []*Guide
	for _, item := range guidesList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		guide := &Guide{}
		if id, ok := itemMap["id"].(string); ok {
			guide.ID = id
		}
		if title, ok := itemMap["title"].(string); ok {
			guide.Title = title
		}
		if type_, ok := itemMap["type"].(string); ok {
			guide.Type = type_
		}
		guides = append(guides, guide)
	}

	return guides, nil
}
