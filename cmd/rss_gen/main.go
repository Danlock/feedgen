package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/danlock/go-rss-gen/db"

	"github.com/jmoiron/sqlx"

	"github.com/danlock/go-rss-gen/lib"

	"github.com/danlock/go-rss-gen/lib/logger"

	"github.com/danlock/go-rss-gen/feed"
	"github.com/danlock/go-rss-gen/gen/feedgen"
	"github.com/joho/godotenv"
)

var (
	buildInfo = "NO INFO"
	buildTag  = "NO TAG"
)

func helpAndQuit() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	logger.SetupLogger(buildTag + " ")
	ctx := context.Background()
	logger.Infof(ctx, "\n%s\nBuild Version: %s\n", buildInfo, runtime.Version())
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

	crdb, err := sqlx.Connect("postgres", lib.GetEnvOrWarn("CRDB_URI"))
	if err != nil {
		logger.Errf(ctx, "Unable to connect to db err:%+v", err)
		os.Exit(1)
	}
	mangaStore := db.NewMangaStore(crdb)
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

	ctx, cancel := context.WithCancel(ctx)
	switch flag.Arg(0) {
	case "poll":
		freq, err := time.ParseDuration(lib.GetEnvOrWarn("RSS_GEN_POLL_FREQUENCY"))
		if err != nil {
			freq = 6 * time.Hour
			logger.Warnf(ctx, "Using default poll duration %s", freq.String())
		}
		releaseChan := feed.PollMUForReleases(ctx, freq)
		for {
			select {
			case releases := <-releaseChan:
				logger.Debugf(ctx, "Got these releases \n%+v", releases)
				if err := mangaStore.UpsertRelease(ctx, releases); err != nil {
					logger.Errf(ctx, "Failed to upsert manga releases err:%+v", err)
				}
			case err := <-errc:
				logger.Infof(ctx, "exiting (%v)", err)
				cancel()
			}
		}
	case "populate-db":
		start, serr := strconv.Atoi(flag.Arg(1))
		end, eerr := strconv.Atoi(flag.Arg(2))
		if serr != nil || eerr != nil || start >= end {
			logger.Errf(ctx, "populate-db takes in two args, the start and end range of the MangaUpdate manga ids to scrape from.")
			os.Exit(1)
		}
		before := time.Now()
		manga, err := feed.QueryMUSeriesRange(ctx, start, end)
		if err != nil {
			logger.Errf(ctx, "Failed to query Mangaupdates err: %+v", err)
			os.Exit(1)
		}
		logger.Infof(ctx, "Took %s to get manga between %d and %d", time.Since(before).String(), start, end)
		logger.Debugf(ctx, "Got these manga\n%+v", manga)
		if err := mangaStore.UpsertManga(ctx, manga); err != nil {
			logger.Errf(ctx, "Failed to upserting manga: err %+v", err)
			os.Exit(1)
		}
		cancel()
	case "api":
		u, err := url.Parse(host)
		if err != nil {
			logger.Errf(ctx, "invalid URL %#v: %s", u.String(), err)
			os.Exit(1)
		}
		var wg sync.WaitGroup
		handleHTTPServer(ctx, u, feedgen.NewEndpoints(feed.New()), &wg, errc, logger.IsDebug())
		// Wait for signal.
		logger.Infof(ctx, "exiting (%v)", <-errc)
		// Send cancellation signal to the goroutines.
		cancel()
		wg.Wait()
	default:
		logger.Infof(ctx, "Available commands are poll,api,populate-db")
		helpAndQuit()
	}
}
