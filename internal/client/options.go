package client

import "time"

type Option func(*Client)

// SetTimeout change request timeout.
func SetTimeout(duration time.Duration) Option {
	return func(args *Client) {
		args.handle.Timeout = duration
	}
}
