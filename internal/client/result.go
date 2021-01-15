package client

import "net/url"

// CheckGitResult result of check function
type CheckGitResult struct {
	Vulnerable bool
	URL        *url.URL
	Debug      *Debug
}
