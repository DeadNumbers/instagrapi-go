package instagrapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

// PhotoUpload uploads a photo to Instagram.
func (c *Client) PhotoUpload(filePath string, caption string, uploadID string) (*Media, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	_ = fileStat

	uploadID = uploadID + "_" + strconv.Itoa(int(time.Now().UnixNano()/1e6))

	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	// Add the image file
	part, err := writer.CreateFormFile("photo", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Add metadata
	writer.WriteField("upload_id", uploadID)
	writer.WriteField("caption", caption)
	writer.WriteField("_uuid", c.UUID)
	writer.WriteField("_uid", strconv.FormatInt(c.UserID, 10))
	writer.WriteField("device[manufacturer]", "Google")
	writer.WriteField("device[model]", "Pixel 8 Pro")
	writer.WriteField("device[android_version]", "34")
	writer.WriteField("device[android_release]", "14")

	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/upload/photo/", apiDomain), bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", c.authorizationHeader())
	req.Header.Set("X-Instagram-Rupload-Params", fmt.Sprintf(`{"upload_id":"%s"}`, uploadID))

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	c.lastResponse = resp

	if resp.StatusCode >= 400 {
		return nil, &ClientError{Message: fmt.Sprintf("Upload failed with status %d", resp.StatusCode), Response: string(bodyBytes)}
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.lastJSON = result

	status, _ := result["status"].(string)
	if status != "ok" {
		msg, _ := result["message"].(string)
		return nil, &ClientError{Message: msg, Response: result}
	}

	// Configure the upload
	return c.photoConfigure(uploadID, caption)
}

func (c *Client) photoConfigure(uploadID string, caption string) (*Media, error) {
	data := map[string]any{
		"upload_id":       uploadID,
		"caption":         caption,
		"_uid":            strconv.FormatInt(c.UserID, 10),
		"_uuid":           c.UUID,
		"source_type":     "4",
		"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
	}

	result, err := c.privateRequest(
		"media/configure/",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return extractMediaFromMap(result)
}

// VideoUpload uploads a video to Instagram.
func (c *Client) VideoUpload(filePath string, caption string, thumbnailPath string) (*Media, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	uploadID := strconv.Itoa(int(time.Now().UnixNano() / 1e6))

	// Upload thumbnail if provided
	if thumbnailPath != "" {
		c.videoUploadThumbnail(thumbnailPath, uploadID)
	}

	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	part, err := writer.CreateFormFile("video", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to write video: %w", err)
	}

	writer.WriteField("upload_id", uploadID)
	writer.WriteField("caption", caption)
	writer.WriteField("_uuid", c.UUID)
	writer.WriteField("_uid", strconv.FormatInt(c.UserID, 10))
	writer.WriteField("length", fmt.Sprintf("%.0f", float64(fileStat.Size())/1e6))
	writer.WriteField("source_type", "4")
	writer.WriteField("thumbnail_data", "")

	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/upload/video/", apiDomain), bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", c.authorizationHeader())
	req.Header.Set("X-Instagram-Rupload-Params", fmt.Sprintf(`{"upload_id":"%s","is_clips_video":"1"}`, uploadID))

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	c.lastResponse = resp

	if resp.StatusCode >= 400 {
		return nil, &ClientError{Message: fmt.Sprintf("Video upload failed with status %d", resp.StatusCode), Response: string(bodyBytes)}
	}

	var result map[string]any
	json.Unmarshal(bodyBytes, &result) // ignore errors
	c.lastJSON = result

	status, _ := result["status"].(string)
	if status != "ok" {
		msg, _ := result["message"].(string)
		return nil, &ClientError{Message: msg, Response: result}
	}

	return c.videoConfigure(uploadID, caption)
}

func (c *Client) videoUploadThumbnail(thumbnailPath string, uploadID string) error {
	file, err := os.Open(thumbnailPath)
	if err != nil {
		return err
	}
	defer file.Close()

	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	part, _ := writer.CreateFormFile("thumbnail", thumbnailPath)
	io.Copy(part, file)

	writer.WriteField("upload_id", uploadID)
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s/upload/thumbnail/", apiDomain), bodyBuf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", c.authorizationHeader())

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *Client) videoConfigure(uploadID string, caption string) (*Media, error) {
	data := map[string]any{
		"upload_id":       uploadID,
		"caption":         caption,
		"_uid":            strconv.FormatInt(c.UserID, 10),
		"_uuid":           c.UUID,
		"source_type":     "4",
		"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
		"clips":           []map[string]any{{"length": 30.0, "source_type": "4"}},
	}

	result, err := c.privateRequest(
		"media/configure_to_clips/",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return extractMediaFromMap(result)
}

// StoryUploadPhoto uploads a photo as a story.
func (c *Client) StoryUploadPhoto(filePath string, mentions []UserShort) (*Story, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	uploadID := strconv.Itoa(int(time.Now().UnixNano() / 1e6))

	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	part, _ := writer.CreateFormFile("photo", filePath)
	io.Copy(part, file)

	writer.WriteField("upload_id", uploadID)
	writer.WriteField("_uuid", c.UUID)
	writer.WriteField("_uid", strconv.FormatInt(c.UserID, 10))
	writer.WriteField("supported_capabilities_new", "[]")
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s/upload/photo/", apiDomain), bodyBuf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", c.authorizationHeader())

	resp, err := c.PrivateClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	return c.storyConfigurePhoto(uploadID)
}

func (c *Client) storyConfigurePhoto(uploadID string) (*Story, error) {
	data := map[string]any{
		"upload_id":       uploadID,
		"_uid":            strconv.FormatInt(c.UserID, 10),
		"_uuid":           c.UUID,
		"source_type":     "4",
		"timezone_offset": strconv.Itoa(c.Settings.TimezoneOffset),
	}

	result, err := c.privateRequest(
		"media/configure_to_story/",
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}

	item := navigateJSON(result, "item")
	if item == nil {
		return nil, &ClientError{Message: "No item in story response"}
	}

	return extractStoryFromMap(item.(map[string]any))
}

// DownloadPhoto downloads a photo from Instagram.
func (c *Client) DownloadPhoto(imageURL string, outputPath string) error {
	resp, err := c.PublicClient.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	io.Copy(outFile, resp.Body)
	return nil
}

// DownloadVideo downloads a video from Instagram.
func (c *Client) DownloadVideo(videoURL string, outputPath string) error {
	resp, err := c.PublicClient.Get(videoURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	io.Copy(outFile, resp.Body)
	return nil
}
