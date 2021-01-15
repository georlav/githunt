package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	c       http.Client
	limiter *rate.Limiter
	qps     int
}

func NewClient(options ...Option) *Client {
	client := Client{
		c: http.Client{
			Transport: &http.Transport{
				TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
				DisableKeepAlives:  true,
				DisableCompression: true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 30,
		},
	}

	for i := range options {
		options[i](&client)
	}

	client.limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(client.qps)), client.qps)

	return &client
}

// Checks check and verify if a target is vulnerable
func (c *Client) CheckGit(ctx context.Context, u url.URL) (*CheckGitResult, error) {
	_ = c.limiter.Wait(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	t, d := trace()
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), t))

	result := &CheckGitResult{Debug: d, URL: u}
	resp, err := c.c.Do(req)
	d.Request.End = time.Now()
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil && bytes.Contains(b, []byte("[core]")) {
			result.Vulnerable = true
			return result, nil
		}
	}

	return result, nil
}
