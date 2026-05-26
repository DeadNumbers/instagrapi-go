package instagrapi

import (
	"time"
)

// GetNotes retrieves direct notes.
func (c *Client) GetNotes() ([]*Note, error) {
	result, err := c.privateRequest(
		"notes/get_notes/",
		nil,
		map[string]string{},
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

	var notes []*Note
	for _, item := range itemsList {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		note := &Note{}
		if id, ok := itemMap["id"].(string); ok {
			note.ID = id
		}
		if text, ok := itemMap["text"].(string); ok {
			note.Text = text
		}
		if userID, ok := itemMap["user_id"].(string); ok {
			note.UserID = userID
		}
		if audience, ok := itemMap["audience"].(float64); ok {
			note.Audience = int(audience)
		}
		if createdAt, ok := itemMap["created_at"].(float64); ok {
			note.CreatedAt = time.Unix(int64(createdAt), 0).UTC()
		}
		if expiresAt, ok := itemMap["expires_at"].(float64); ok {
			note.ExpiresAt = time.Unix(int64(expiresAt), 0).UTC()
		}

		// Extract user info
		if userData, ok := itemMap["user"].(map[string]any); ok {
			user, err := extractUserShortFromMap(userData)
			if err == nil && user != nil {
				note.User = *user
			}
		}

		notes = append(notes, note)
	}

	return notes, nil
}

// GetNoteByUser finds a note for a specific username from a notes list.
func (c *Client) GetNoteByUser(notes []*Note, username string) *Note {
	usernameLower := username
	for _, note := range notes {
		if note.User.Username != "" {
			if note.User.Username == usernameLower {
				return note
			}
		}
	}
	return nil
}

// GetNoteTextByUser gets the text of a note for a specific user.
func (c *Client) GetNoteTextByUser(notes []*Note, username string) string {
	note := c.GetNoteByUser(notes, username)
	if note != nil {
		return note.Text
	}
	return ""
}

// DeleteNote deletes a personal note.
func (c *Client) DeleteNote(noteID int64) error {
	data := map[string]any{
		"id":    noteID,
		"_uuid": c.UUID,
	}

	result, err := c.privateRequest(
		"notes/delete_note/",
		data,
		nil,
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Failed to delete note"}
	}

	return nil
}

// CreateNote creates a personal note.
func (c *Client) CreateNote(text string, audience int) (*Note, error) {
	data := map[string]any{
		"note_style": 0,
		"text":       text,
		"_uuid":      c.UUID,
		"audience":   audience,
	}

	result, err := c.privateRequest(
		"notes/create_note",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		msg, _ := result["message"].(string)
		return nil, &ClientError{Message: msg}
	}

	note := &Note{}
	if id, ok := result["id"].(string); ok {
		note.ID = id
	}
	if text, ok := result["text"].(string); ok {
		note.Text = text
	}
	if userID, ok := result["user_id"].(string); ok {
		note.UserID = userID
	}

	return note, nil
}

// UpdateNotesLastSeen updates the last seen timestamp for notes.
func (c *Client) UpdateNotesLastSeen() error {
	data := map[string]any{
		"_uuid": c.UUID,
	}

	result, err := c.privateRequest(
		"notes/update_notes_last_seen_timestamp/",
		data,
		nil,
	)
	if err != nil {
		return err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		return &ClientError{Message: "Failed to update notes last seen"}
	}

	return nil
}

// NotesMusicBrowser retrieves music candidates for Instagram Notes.
func (c *Client) NotesMusicBrowser() (map[string]any, error) {
	data := map[string]any{
		"product": "music_notes",
		"_uuid":   c.UUID,
	}

	result, err := c.privateRequest(
		"music/notes_audio_browser/",
		data,
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	status, _ := result["status"].(string)
	if status != "ok" {
		msg, _ := result["message"].(string)
		return nil, &ClientError{Message: msg}
	}

	return result, nil
}
