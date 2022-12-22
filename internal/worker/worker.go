package worker

import (
	"context"
	"net/url"
	"sync"

	"github.com/georlav/githunt/internal/client"
)

type Target struct {
	URL   *url.URL
	Error error
}

type Result struct {
	URL        *url.URL
	Vulnerable bool
	Error      error
}

func Work(
	ctx context.Context,
	targets <-chan Target,
	c *client.Client,
	workers int,
) <-chan Result {
	resultCH := make(chan Result)

	wg := sync.WaitGroup{}
	wg.Add(workers)

	for i := 1; i <= workers; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-targets:
					if !ok {
						return
					}

					// handle invalid targets
					if t.Error != nil {
						resultCH <- Result{Error: t.Error}
						continue
					}

					isVulnerable, err := c.CheckGit(ctx, t.URL)
					resultCH <- Result{
						URL:        t.URL,
						Vulnerable: isVulnerable,
						Error:      err,
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCH)
	}()

	return resultCH
}
