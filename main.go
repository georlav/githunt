package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/georlav/githunt/internal/client"
)

var version string

func main() {
	var (
		fmtError = color.New(color.FgRed, color.Bold)
		fmtInfo  = color.New(color.FgGreen, color.Bold)
	)

	// CLI params
	target := flag.String("url", "", "check single url")
	targets := flag.String("urls", "", "file with urls, should have one url per line")
	workers := flag.Int("workers", 50, "sets the number of http workers")
	cpus := flag.Int("cpus", runtime.NumCPU()-1, "sets the maximum number of CPUs that can be utilized")
	rateLimit := flag.Int("rate-limit", 500, "limit requests per second (default: 500)")
	timeout := flag.Int64("timeout", 15, "set a time limit for requests in seconds (default: 15)")
	output := flag.String("output", "", "save vulnerable targets in a file")
	debug := flag.Bool("debug", false, "enable debug")
	flag.Usage = usage(runtime.NumCPU()-1, version)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "debug", *debug)
	go terminate(cancel)

	// set the maximum number of CPUs that can be utilized
	runtime.GOMAXPROCS(*cpus)

	// target is required
	if *targets == "" && *target == "" {
		fmtError.Fprint(os.Stderr, "You need to specify a url\n")
		os.Exit(1)
	}

	// Initialize http client
	c := client.NewClient(
		client.SetTimeout(*timeout),
		client.SetQPS(*rateLimit),
	)

	var (
		tScanned    uint64
		tVulnerable uint64
		started     = time.Now()
	)

	defer func() {
		fmtInfo.Printf("Scanned: %d target(s) in %s found: %d vulnerable\n\n",
			atomic.LoadUint64(&tScanned),
			time.Since(started).String(),
			tVulnerable,
		)
	}()

	// load targets
	targetsCH, err := loadTargets(ctx, *targets, *target)
	if err != nil {
		fmtError.Printf("Failed to load targets. Error: %s\n", err)
		os.Exit(1)
	}

	// save vulnerable targets to output file
	vulnerableCH := make(chan string)
	if err := save(ctx, vulnerableCH, *output); err != nil {
		fmtError.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	resultCh := client.RunGitCheckWorkers(ctx, targetsCH, c, *workers)

	// handle results
	for r := range resultCh {
		// handle target parsing errors
		if r.Target.Error != nil {
			fmtError.Fprintf(os.Stderr, "%s | Error: %s\n", r.Result.Debug.String(), r.Target.Error)
			os.Exit(1)
		}

		// handle request errors
		if r.Error != nil {
			if *debug && r.Result != nil {
				fmtError.Fprintf(os.Stderr, "%s | Error: %s\n", r.Result.Debug.String(), r.Error)
			}

			if strings.Contains(r.Error.Error(), "too many open files") {
				fmtError.Fprint(os.Stderr, "You need to increase ulimit for open files or decrease number of workers\n")
				os.Exit(1)
			}
		}

		if *output != "" && r.Result.Vulnerable {
			tVulnerable++
			fmtInfo.Printf("  Target: %s is vulnerable.\n", r.Target.URL.String())
			vulnerableCH <- r.Result.URL.String()
		}

		if *debug && r.Result != nil {
			fmt.Println(r.Result.Debug)
		}

		atomic.AddUint64(&tScanned, 1)
		fmtInfo.Printf("Scanned: %d target(s) in %s found: %d vulnerable\r",
			atomic.LoadUint64(&tScanned),
			time.Since(started).String(),
			tVulnerable,
		)
	}
}

// save results to file
func save(ctx context.Context, results <-chan string, output string) error {
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

func loadTargets(ctx context.Context, filename string, target string) (<-chan client.Target, error) {
	targets := make(chan client.Target)

	// single target
	if target != "" {
		tURL, err := url.Parse(target)
		if err != nil {
			return nil, err
		}
		if tURL.Scheme == "" {
			tURL.Scheme = "https"
		}

		go func() {
			targets <- client.Target{URL: *tURL}
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
						targets <- client.Target{
							Error: fmt.Errorf("unable to parse %s, will skip. Error: %w", scanner.Text(), err),
						}
						continue
					}
					if u.Scheme == "" {
						u.Scheme = "https"
					}

					u.Path = u.Path + "/.git/config"
					targets <- client.Target{URL: *u}
				}

				if err := scanner.Err(); err != nil {
					targets <- client.Target{
						Error: err,
					}
				}
			}
		}()
	}

	return targets, nil
}

// terminate on termination signal send cancellation signal
func terminate(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
}

// help menu
func usage(cpus int, version string) func() {
	return func() {
		var usage = `
  _   o  _|_  |_        ._   _|_ 
 (_|  |   |_  | |  |_|  | |   |_  %s 
  _|
Usage: githunt [options...] 

Usage Examples:
  githunt -target example.com
  githunt -targets urls.txt -workers 100 -timeout 5 -output out.txt

Options:
  -url         url of target
  -urls        file containing multiple urls (one per line)
  -rate-limit  limit requests per second (default: 500)
  -workers     sets the desirable number of http workers (default: 50)
  -cpus        sets the maximum number of CPUs that can be utilized (default: %d)
  -timeout     set a time limit for requests in seconds (default: 15)
  -output      save vulnerable targets in a file
  -debug       enable debug messages (default: disabled)

`
		color.New(color.FgGreen, color.Bold).Printf(usage, version, cpus)
	}
}
