package fcm

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const defaultEndpoint = "https://fcm.googleapis.com/v1"

// Client for the Firebase Cloud Messaging (FCM) service.
type Client struct {
	httpClient httpClient
	endpoint   string
	project    string
	version    string
}

type Config struct {
	Client      httpClient
	Credentials []byte
	ProjectID   string
	Endpoint    string
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewClient creates a new instance of the Firebase Cloud Messaging Client.
func NewClient(cfg Config) (*Client, error) {
	switch {
	case len(cfg.Credentials) == 0:
		return nil, errors.New("credentials not provided")
	case cfg.ProjectID == "":
		return nil, errors.New("project ID is required to access Firebase Cloud Messaging client")
	}

	if cfg.Client == nil {
		trans, err := newHTTPClient(cfg.Credentials)
		if err != nil {
			return nil, fmt.Errorf("cannot create HTTP client: %w", err)
		}
		cfg.Client = trans
	}

	sendEndpoint := cmp.Or(cfg.Endpoint, defaultEndpoint)

	return &Client{
		httpClient: cfg.Client,
		endpoint:   fmt.Sprintf("%s/projects/%s/messages:send", sendEndpoint, cfg.ProjectID),
		version:    "github.com/cristalhq/fcm",
	}, nil
}

// Send a [Message] to Firebase Cloud Messaging (FCM).
//
// The Message must specify exactly one of Token, Topic and Condition fields.
// FCM will customize the message for each target platform based on the arguments specified in the [Message].
func (c *Client) Send(ctx context.Context, message *Message) (string, error) {
	if err := validateMessage(message); err != nil {
		return "", err
	}
	return c.send(ctx, message)
}

func (c *Client) send(ctx context.Context, message *Message) (string, error) {
	msg := struct {
		Message *Message `json:"message"`
	}{
		Message: message,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("c.httpClient.Do: %w", err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("io.ReadAll: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("code: %d, body: '%s", resp.StatusCode, string(b))
	}

	var result fcmResponse
	if err := json.Unmarshal(b, &result); err != nil {
		var errResp fcmErrorResponse
		if err := json.Unmarshal(b, &errResp); err != nil {
			return "", fmt.Errorf("json.Unmarshal(b, &errResp): %w", err)
		}
		return "", fmt.Errorf("json.Unmarshal(b, &resp): %w", err)
	}

	return result.Name, nil
}

type fcmResponse struct {
	Name string `json:"name"`
}

type fcmErrorResponse struct {
	Error struct {
		Details []struct {
			Type      string `json:"@type"`
			ErrorCode string `json:"errorCode"`
		}
	} `json:"error"`
}
