package client

import "time"

type Option func(*Client)

// SetTimeout change request timeout
func SetTimeout(duration time.Duration) Option {
	return func(args *Client) {
		args.c.Timeout = duration
	}
}

// SetQPS set queries per second limit
func SetQPS(limit int) Option {
	return func(args *Client) {
		args.qps = limit
	}
}
