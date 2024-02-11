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
		out, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("opening file %s. Error: %w", output, err)
		}

		if err := out.Truncate(0); err != nil {
			return fmt.Errorf("truncating file %s. Error: %w", output, err)
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
func LoadTargetURLs(ctx context.Context, filename, target, urlPath string) (<-chan worker.Target, error) {
	targets := make(chan worker.Target)

	// single target
	if target != "" {
		tURL, err := url.Parse(target)
		if err != nil {
			return nil, fmt.Errorf("parsing url %s. Error: %w", target, err)
		}
		if tURL.Scheme == "" {
			tURL.Scheme = "https"
		}
		tURL.Path += urlPath

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
			return nil, fmt.Errorf("opening targets file %s. Error: %w", filename, err)
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
							Error: fmt.Errorf("parsing %s. Error: %w", scanner.Text(), err),
						}
						continue
					}

					if u.Scheme == "" {
						u.Scheme = "https"
					}

					u.Path += urlPath

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

// help menu.
func Usage(cpus int, version string) func() {
	return func() {
		usage := `
  _   o  _|_  |_        ._   _|_ 
 (_|  |   |_  | |  |_|  | |   |_  %s 
  _|
Usage: githunt [options...] 

Usage Examples:
  githunt -url example.com
  githunt -urls urls.txt -workers 100 -timeout 30s -output out.txt

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
