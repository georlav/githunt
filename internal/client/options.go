package client

import "time"

type Option func(*Client)

// SetTimeout change request timeout
func SetTimeout(seconds int64) Option {
	return func(args *Client) {
		args.c.Timeout = time.Second * time.Duration(seconds)
	}
}

// SetQPS set queries per second limit
func SetQPS(limit int) Option {
	return func(args *Client) {
		args.qps = limit
	}
}
