package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	netpprof "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"

	"github.com/rs/cors"

	"github.com/danlock/feedgen/api"

	openruntime "github.com/go-openapi/runtime"

	"github.com/danlock/feedgen/db"
	"github.com/danlock/feedgen/gen/restapi"
	"github.com/danlock/feedgen/gen/restapi/operations"
	"github.com/danlock/feedgen/scrape"
	loads "github.com/go-openapi/loads"

	"github.com/danlock/feedgen/lib"
	"github.com/jmoiron/sqlx"

	"github.com/danlock/feedgen/lib/logger"

	"github.com/joho/godotenv"
)

var (
	buildInfo = "NO INFO"
	buildTag  = "NO TAG"
)

func helpAndQuit() {
	fmt.Fprintf(flag.CommandLine.Output(), `
Usage of %s:
	poll :	Polls sites for updates with the given frequency in go time.Duration format
	populate-db:	Scrapes the given range of ids from MangaUpdates
	api:	serves an API on the URL provided (defaulting to http://localhost:8080) with RSS, Atom or JSON Feed endpoints.
`, os.Args[0])
	flag.PrintDefaults()
	os.Exit(0)
}

func main() {
	logger.SetupLogger(buildTag + " ")
	ctx := context.Background()
	logger.Infof(ctx, "%s Built With: %s", buildInfo, runtime.Version())
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		dotenvLocation string
		help           bool
	)
	flag.Usage = helpAndQuit
	flag.StringVar(&dotenvLocation, "e", "./ops/.env", "Location of .env file with environment variables in KEY=VALUE format")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.Parse()

	if help {
		helpAndQuit()
	}
	if err := godotenv.Overload(dotenvLocation); err != nil {
		logger.Warnf(ctx, "No .env file found")
	}
	var crdb *sqlx.DB
	var err error
	for {
		crdb, err = sqlx.Connect("postgres", lib.GetEnvOrWarn("CRDB_URI"))
		if err != nil {
			logger.Errf(ctx, "Unable to connect to db, retrying... err:%+v", err)
			time.Sleep(250 * time.Millisecond)
		} else {
			break
		}
	}
	mangaStore := db.NewMangaStore(crdb)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		logger.Infof(ctx, "Received signal %s, shutting down...", <-c)
		if logger.IsDebug() {
			f, err := os.Create(fmt.Sprintf("%s/%s_goroutine_trace.txt", os.TempDir(), path.Base(os.Args[0])))
			if err == nil {
				logger.Dbgf(ctx, "Writing pprof profiles to %s", f.Name())
				pprof.Lookup("goroutine").WriteTo(f, 2)
			} else {
				logger.Dbgf(ctx, "Failed to write pprof profiles err: %+v", err)
			}
		}
		cancel()
	}()

	// Spin off http debug pprof server for realtime
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", netpprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", netpprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", netpprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", netpprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", netpprof.Trace)
		debugURL := "localhost:18080"
		logger.Infof(ctx, "Serving runtime profiling server on %s...", debugURL)
		http.ListenAndServe(debugURL, mux)
	}()

	switch flag.Arg(0) {
	case "poll":
		freq, err := time.ParseDuration(flag.Arg(1))
		if err != nil {
			logger.Errf(ctx, "poll takes in 1 arg, the duration between each polling attempt. 6h is recommended.")
			os.Exit(1)
		}
		if handlePoll(ctx, mangaStore, freq) != nil {
			os.Exit(1)
		}
	case "populate-db":
		start, serr := strconv.Atoi(flag.Arg(1))
		end, eerr := strconv.Atoi(flag.Arg(2))
		if serr != nil || eerr != nil || start >= end {
			logger.Errf(ctx, "populate-db takes in two args, the start and end range of the MangaUpdate manga ids to scrape from.")
			os.Exit(1)
		}
		if handleDBPopulation(ctx, mangaStore, start, end) != nil {
			os.Exit(1)
		}
	case "api":
		u, err := url.Parse(flag.Arg(1))
		if flag.Arg(1) == "" || err != nil {
			defaultURL := "http://localhost:8080"
			logger.Errf(ctx, "invalid URL %s: %s, using default %s", flag.Arg(1), err, defaultURL)
			if u, err = url.Parse(defaultURL); err != nil {
				panic("defaultURL is invalid URL")
			}
		}
		handleHTTPServer(ctx, u, apiModels{mangaStore: mangaStore})
	default:
		logger.Infof(ctx, "Available commands are poll,api,populate-db")
		helpAndQuit()
	}
}

func handleDBPopulation(ctx context.Context, mangaStore db.MangaStorer, start, end int) error {
	before := time.Now()
	manga, err := scrape.QueryMUSeriesRange(ctx, start, end)
	if err != nil {
		logger.Errf(ctx, "Failed to query Mangaupdates err: %+v", err)
		return err
	}
	logger.Infof(ctx, "Took %s to get %d manga between %d and %d", time.Since(before).String(), len(manga), start, end)
	select {
	case <-ctx.Done():
		logger.Infof(ctx, "Context closed, shutting down populate-db...")
		return ctx.Err()
	default:
	}
	if len(manga) == 0 {
		return nil
	}
	if err := mangaStore.UpsertManga(ctx, manga); err != nil {
		logger.Errf(ctx, "Failed to upserting manga: err %+v", err)
		return err
	}
	return nil
}

func handlePoll(ctx context.Context, mangaStore db.MangaStorer, freq time.Duration) error {
	// Scrape new releases out of MU
	releaseChan := scrape.PollMUForReleases(ctx, freq)
	for {
		select {
		case releases, running := <-releaseChan:
			if !running {
				return nil
			}
			// Each new batch of releases may include new manga not in the db, filter them out
			newReleases, err := mangaStore.FilterOutReleasesWithoutMangaInDB(ctx, releases)
			if err != nil {
				logger.Errf(ctx, "Failed to filter out new manga releases!")
			}
			// Scrape the info page for each new manga found and place them in db
			newRelLen := len(newReleases)
			if newRelLen > 0 {
				logger.Infof(ctx, "Found %d new titles, scraping their pages...", newRelLen)
				muids := make([]int, 0, newRelLen)
				for _, r := range newReleases {
					muids = append(muids, r.MUID)
				}
				if manga, err := scrape.QueryMUSeriesIds(ctx, muids); err != nil {
					logger.Errf(ctx, "Failed to query new manga release pages err:%+v", err)
				} else if err := mangaStore.UpsertManga(ctx, manga); err != nil {
					logger.Errf(ctx, "Failed to upsert manga for new releases err: %+v", err)
				}
			}
			// Finally upsert releases in DB in case there are duplicates
			if err := mangaStore.UpsertRelease(ctx, releases); err != nil {
				logger.Errf(ctx, "Failed to upsert manga releases err:%+v", err)
			}
		case <-ctx.Done():
			logger.Infof(ctx, "exiting (%v)", ctx.Err())
			return ctx.Err()
		}
	}
}

type apiModels struct {
	mangaStore db.MangaStorer
}

func handleHTTPServer(ctx context.Context, u *url.URL, models apiModels) {
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		logger.Errf(ctx, "%+v", err)
		return
	}
	operationsAPI := operations.NewFeedgenAPI(swaggerSpec)
	if err := operationsAPI.Validate(); err != nil {
		logger.Errf(ctx, "%+v", err)
		return
	}
	operationsAPI.JSONConsumer = openruntime.JSONConsumer()
	operationsAPI.XMLConsumer = openruntime.XMLConsumer()
	operationsAPI.Logger = func(msg string, vals ...interface{}) { logger.InfofWithCallDepth(ctx, 5, msg, vals) }

	// lazyEncoder just check if the result is a string before encoding. If it is, it assumes it has already been encoded and sends it directly.
	type encoder interface{ Encode(interface{}) error }
	lazyEncoder := func(enc encoder, writer io.Writer, data interface{}) error {
		if str, isString := data.(string); isString {
			_, err := writer.Write([]byte(str))
			return err
		}
		return enc.Encode(data)
	}
	operationsAPI.XMLProducer = openruntime.ProducerFunc(func(writer io.Writer, data interface{}) error {
		return lazyEncoder(xml.NewEncoder(writer), writer, data)
	})
	operationsAPI.JSONProducer = openruntime.ProducerFunc(func(writer io.Writer, data interface{}) error {
		enc := json.NewEncoder(writer)
		enc.SetEscapeHTML(false)
		return lazyEncoder(enc, writer, data)
	})
	fs := api.NewFeedSrvc(u.String(), models.mangaStore)
	operationsAPI.FeedgenMangaHandler = operations.FeedgenMangaHandlerFunc(fs.Manga)
	operationsAPI.FeedgenViewMangaHandler = operations.FeedgenViewMangaHandlerFunc(fs.ViewManga)
	operationsAPI.Init()

	server := restapi.NewServer(operationsAPI)
	defer server.Shutdown()
	// Manually setup middleware chain for more control over specific middleware,
	// Such as not serving the latest, currently broken version of ReDoc.
	server.SetHandler(
		logger.LoggerMiddleware(
			cors.AllowAll().Handler(
				middleware.Spec("", swaggerSpec.Raw(),
					middleware.Redoc(
						middleware.RedocOpts{
							Path:     "api/docs",
							RedocURL: "https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js",
						},
						operationsAPI.Context().RoutesHandler(nil),
					)))))
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		port = 80
	}
	server.Port = port
	if err := server.Serve(); err != nil {
		logger.Errf(ctx, "%+v", err)
	}
}
