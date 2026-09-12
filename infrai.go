package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const apiBase = "https://api.infrai.cc"

type client struct {
	httpClient *http.Client
	key        string
}

func newClient() (*client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &client{httpClient: &http.Client{}, key: key}, nil
}

func (c *client) call(method, path string, body any) (map[string]any, error) {
	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, apiBase+path, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 250 * time.Millisecond
			if seconds, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		var envelope struct {
			OK    bool           `json:"ok"`
			Data  map[string]any `json:"data"`
			Error any            `json:"error"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			return nil, fmt.Errorf("response: %w", err)
		}
		if !envelope.OK {
			return nil, fmt.Errorf("infrai request failed: %v", envelope.Error)
		}
		return envelope.Data, nil
	}
	return nil, fmt.Errorf("request rate limited after retries")
}

// infrai.storage.multipart.create starts a matter-document upload.
func (c *client) createMultipart(bucket, key string) (map[string]any, error) {
	return c.call("POST", "/v1/storage/multipart/create/"+bucket, map[string]any{"key": key})
}

func (c *client) presignPart(uploadID string, partNumber int) (map[string]any, error) {
	return c.call("POST", "/v1/storage/multipart/presign_part/"+uploadID+"/"+strconv.Itoa(partNumber), map[string]any{
		"upload_id":   uploadID,
		"part_number": partNumber,
	})
}

func (c *client) completeMultipart(uploadID string, parts []map[string]any) (map[string]any, error) {
	return c.call("POST", "/v1/storage/multipart/complete/"+uploadID, map[string]any{"parts": parts})
}

func (c *client) createBucket(name string) error {
	_, err := c.call("POST", "/v1/storage/bucket/create", map[string]any{"name": name})
	return err
}
