package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/danlock/go-rss-gen/db"

	"github.com/danlock/go-rss-gen/api"
	"github.com/danlock/go-rss-gen/gen/feedgen"
	feedgensvr "github.com/danlock/go-rss-gen/gen/http/feedgen/server"
	"github.com/danlock/go-rss-gen/lib/logger"
	goahttp "goa.design/goa/http"
	httpmdlwr "goa.design/goa/http/middleware"
	"goa.design/goa/middleware"
)

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(ctx context.Context, u *url.URL, wg *sync.WaitGroup, mangaStore db.MangaStorer) {

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	mux := goahttp.NewMuxer()

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	feedgenServer := feedgensvr.New(feedgen.NewEndpoints(api.NewFeedSrvc(mangaStore)), mux, dec, enc, errorHandler())
	// Configure the mux.
	feedgensvr.Mount(mux, feedgenServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if logger.IsDebug() {
			handler = httpmdlwr.Debug(mux, os.Stdout)(handler)
		}
		handler = httpmdlwr.PopulateRequestContext()(handler)
		handler = httpmdlwr.RequestID(middleware.UseRequestIDOption(true))(handler)
		handler = httpmdlwr.Log(&logger.GoaLogAdapter{})(handler)
	}
	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler}
	for _, m := range feedgenServer.Mounts {
		logger.Infof(ctx, "HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}

	logger.Infof(ctx, "HTTP server listening on %q", u.Host)
	wg.Add(1)
	// Spin off goroutine for http server
	go logger.Errf(ctx, "Server has returned with: %+v", srv.ListenAndServe())
	// Spin off goroutine for server graceful shutdown
	go func() {
		defer wg.Done()
		<-ctx.Done()
		logger.Infof(ctx, "gracefully shutting down HTTP server at %q", u.Host)
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler() func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Errf(ctx, "req_id: %s ERROR: %s", id, err.Error())
	}
}
