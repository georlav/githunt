package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/georlav/githunt/internal/client"
	"github.com/georlav/githunt/internal/utils"
	"github.com/georlav/githunt/internal/worker"
)

var version string

func main() {
	var (
		fmtError = color.New(color.FgRed, color.Bold)
		fmtInfo  = color.New(color.FgGreen, color.Bold)
	)

	// CLI params
	target := flag.String("url", "", "check single url")
	targets := flag.String("urls", "", "file containing multiple urls (one per line)")
	urlPath := flag.String("path", "/.git/config", "sets the path to .git config file (default: /.git/config)")
	workers := flag.Int("workers", 50, "sets the desirable number of http workers")
	cpus := flag.Int("cpus", runtime.NumCPU()-1, "sets the maximum number of CPUs that can be utilized")
	timeout := flag.Duration("timeout", time.Second*15,
		`sets a time limit for requests, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".`,
	)
	output := flag.String("output", "", "save vulnerable targets in a file")
	flag.Usage = utils.Usage(runtime.NumCPU()-1, version)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go terminate(cancel)

	// set the maximum number of CPUs that can be utilized
	runtime.GOMAXPROCS(*cpus)

	// target is required
	if *targets == "" && *target == "" {
		flag.Usage()
		fmtError.Fprint(os.Stderr, "You need to specify a target\n")
		os.Exit(0)
	}

	// Initialize http c
	c := client.NewClient(
		client.SetTimeout(*timeout),
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
	targetsCH, err := utils.LoadTargetURLs(ctx, *targets, *target, *urlPath)
	if err != nil {
		fmtError.Printf("Failed to load targets. Error: %s\n", err)
		os.Exit(1)
	}

	// save vulnerable targets in a file
	vulnerableCH := make(chan string)
	if err = utils.SaveResults(ctx, vulnerableCH, *output); err != nil {
		fmtError.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	resultCH := worker.Work(ctx, targetsCH, c, *workers)

	// handle results
	for result := range resultCH {
		// handle request errors
		if result.Error != nil {
			fmtError.Fprintf(os.Stderr, "Request Error: %s\n", result.Error)

			if strings.Contains(result.Error.Error(), "too many open files") {
				fmtError.Fprintf(os.Stderr, "%s, You need to increase ulimit for open files or decrease number of workers\n", err)
				os.Exit(1)
			}
		}

		if result.Vulnerable {
			tVulnerable++
			fmtInfo.Printf("Target: %s is vulnerable.\n", result.URL.String())
			if *output != "" {
				vulnerableCH <- result.URL.String()
			}
		}

		atomic.AddUint64(&tScanned, 1)
		fmtInfo.Printf("Scanned: %d target(s) in %s found: %d vulnerable\r",
			atomic.LoadUint64(&tScanned),
			time.Since(started).String(),
			tVulnerable,
		)
	}
}

// terminate on SIGINT or SIGTERM.
func terminate(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
}
