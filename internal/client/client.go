package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	handle *http.Client
}

func NewClient(options ...Option) *Client {
	client := Client{
		handle: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:    &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				DisableKeepAlives:  true,
				DisableCompression: true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 15,
		},
	}

	for i := range options {
		options[i](&client)
	}

	return &client
}

// Checks check and verify if a target is vulnerable.
func (c *Client) CheckGit(ctx context.Context, u *url.URL) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return false, fmt.Errorf("creating request. Error: %w", err)
	}

	resp, err := c.handle.Do(req)
	if err != nil {
		return false, fmt.Errorf("sending request. Error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err == nil && bytes.Contains(b, []byte("[core]")) {
			return true, nil
		}
	}

	return false, nil
}
