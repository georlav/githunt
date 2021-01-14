package client

import (
	"context"
	"net/url"
	"sync"
)

type Target struct {
	URL     url.URL
	Error   error
	Retries int
}

type TargetResult struct {
	Target Target
	Result *CheckGitResult
	Error  error
}

func RunGitCheckWorkers(
	ctx context.Context,
	targets <-chan Target,
	client *Client,
	workers int,
) <-chan TargetResult {
	resultCH := make(chan TargetResult)

	wg := sync.WaitGroup{}
	wg.Add(workers)

	go func() {
		for i := 1; i <= workers; i++ {
			go func() {
				for {
					select {
					case <-ctx.Done():
						wg.Done()
						return
					case t, ok := <-targets:
						if !ok {
							wg.Done()
							return
						}

						if t.Error != nil {
							resultCH <- TargetResult{Error: t.Error}
							continue
						}

						result, err := client.CheckGit(ctx, t.URL)
						resultCH <- TargetResult{Result: result, Target: t, Error: err}
					}
				}
			}()
		}
	}()

	go func() {
		wg.Wait()
		close(resultCH)
	}()

	return resultCH
}
