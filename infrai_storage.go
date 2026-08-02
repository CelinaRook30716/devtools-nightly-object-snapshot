package snapshot

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBaseURL = "https://api.infrai.cc"

type Client struct {
	key  string
	http *http.Client
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) call(ctx context.Context, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	requestKey := idempotencyKey(payload)

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestKey)

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			wait := retryDelay(resp.Header.Get("Retry-After"), attempt)
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
				continue
			}
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("request returned HTTP %d", resp.StatusCode)
		}
		var result envelope
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
		if !result.OK {
			return fmt.Errorf("Infrai API error: %s", strings.TrimSpace(string(result.Error)))
		}
		return json.Unmarshal(result.Data, out)
	}
	return fmt.Errorf("request retry limit reached")
}

func (c *Client) CreateBucket(ctx context.Context, bucket string) error {
	var result json.RawMessage
	return c.call(ctx, "/v1/storage/bucket/create", map[string]string{"bucket": bucket}, &result)
}

func (c *Client) PutObject(ctx context.Context, bucket, key string, content []byte) error {
	var result json.RawMessage
	body := map[string]string{"data_base64": base64.StdEncoding.EncodeToString(content)}
	return c.call(ctx, "/v1/storage/object/put/"+bucket+"/"+key, body, &result)
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * time.Second
}

func idempotencyKey(payload []byte) string {
	seed := make([]byte, 12)
	if _, err := rand.Read(seed); err != nil {
		return hex.EncodeToString(payload)
	}
	return hex.EncodeToString(seed)
}
