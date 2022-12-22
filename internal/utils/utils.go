package utils

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/fatih/color"

	"github.com/georlav/githunt/internal/worker"
)

func SaveResults(ctx context.Context, results <-chan string, output string) error {
	if output != "" {
		out, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		if err := out.Truncate(0); err != nil {
			return err
		}

		go func() {
			defer out.Close()

			for {
				select {
				case <-ctx.Done():
					return
				case r, ok := <-results:
					if !ok {
						return
					}

					if _, err := out.WriteString(r + "\n"); err != nil {
						panic(fmt.Sprintf("Failed to save result %s. Error: %s\n", r, err))
					}
				}
			}
		}()
	}

	return nil
}

//nolint:gocognit
func LoadTargetURLs(ctx context.Context, filename, target string) (<-chan worker.Target, error) {
	targets := make(chan worker.Target)

	// single target
	if target != "" {
		tURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		if tURL.Scheme == "" {
			tURL.Scheme = "https"
		}
		tURL.Path += "/.git/config"

		go func() {
			targets <- worker.Target{URL: tURL}
			if filename == "" {
				close(targets)
			}
		}()
	}

	// targets from file
	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		go func() {
			defer file.Close()
			defer close(targets)

			select {
			case <-ctx.Done():
				return
			default:
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					u, err := url.Parse(scanner.Text())
					if err != nil {
						targets <- worker.Target{
							Error: fmt.Errorf("unable to parse %s, will skip. Error: %w", scanner.Text(), err),
						}
						continue
					}
					if u.Scheme == "" {
						u.Scheme = "https"
					}
					u.Path += "/.git/config"

					targets <- worker.Target{URL: u}
				}

				if err := scanner.Err(); err != nil {
					targets <- worker.Target{
						Error: err,
					}
				}
			}
		}()
	}

	return targets, nil
}

// help menu
func Usage(cpus int, version string) func() {
	return func() {
		var usage = `
  _   o  _|_  |_        ._   _|_ 
 (_|  |   |_  | |  |_|  | |   |_  %s 
  _|
Usage: githunt [options...] 

Usage Examples:
  githunt -target example.com
  githunt -targets urls.txt -workers 100 -timeout 30s -output out.txt

Options:
  Target:
    -url         check single url
    -urls        file containing multiple urls (one per line)

  Request:
    -workers     sets the desirable number of http workers (default: 50)
    -cpus        sets the maximum number of CPUs that can be utilized (default: %d)
    -timeout     sets a time limit for requests, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default: 15s)
  
  General:
    -output      save vulnerable targets to a file

`
		color.New(color.FgGreen, color.Bold).Printf(usage, version, cpus)
	}
}
