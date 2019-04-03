package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/danlock/go-rss-gen/lib"

	"github.com/danlock/go-rss-gen/lib/logger"

	"github.com/danlock/go-rss-gen/feed"
	"github.com/danlock/go-rss-gen/gen/feedgen"
	"github.com/joho/godotenv"
)

func helpAndQuit() {
	flag.Usage()
	os.Exit(0)
}

func main() {
	logger.SetupLogger()
	ctx := context.Background()
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		dotenvLocation string
		help           bool
		host           string
	)
	flag.StringVar(&host, "host", "http://localhost:80", "Server host ")
	flag.StringVar(&dotenvLocation, "e", "./ops/.env", "Location of .env file with environment variables in KEY=VALUE format")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.Parse()

	if help {
		helpAndQuit()
	}
	if err := godotenv.Overload(dotenvLocation); err != nil {
		logger.Warnf(ctx, "No .env file found")
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		errc <- fmt.Errorf("%s", <-c)
	}()

	ctx, cancel : = context.WithCancel(ctx)
	switch flag.Arg(0) {
	case "poll":
		freq, err := time.ParseDuration(lib.GetEnvOrWarn("RSS_GEN_POLL_FREQUENCY"))
		if err != nil {
			freq = 6 * time.Hour
			logger.Warnf(ctx, "Using default poll duration %s", freq.String())
		}
		releaseChan := feed.PollMUForReleases(ctx, freq)
		select {
		case releases := <-releaseChan:
			logger.Infof(ctx,"Got these releases \n%+v",releases)
		case err := <-errc:
			logger.Printf("exiting (%v)", err)
			cancel()
		}
	case "api":
		u, err := url.Parse(host)
		if err != nil {
			logger.Errf(ctx,"invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		var wg sync.WaitGroup
		handleHTTPServer(ctx, u, feedgen.NewEndpoints(feed.New()), &wg, errc, logger, lib.GetEnvOrWarn("RSS_GEN_DEBUG") == "true")
		// Wait for signal.
		logger.Printf("exiting (%v)", <-errc)
		// Send cancellation signal to the goroutines.
		cancel()
		wg.Wait()
		logger.Println("exited")
	default:
		logger.Infof(ctx,"Available commands are poll,api")
		helpAndQuit()
	}
}
