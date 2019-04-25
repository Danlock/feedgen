package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"mime"
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
		enc = encoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	mux := goahttp.NewMuxer()

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	feedgenServer := feedgensvr.New(feedgen.NewEndpoints(api.NewFeedSrvc(u.String(), mangaStore)), mux, dec, enc, errorHandler())
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
	go func() {
		logger.Errf(ctx, "Server has returned with: %+v", srv.ListenAndServe())
	}()

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
		w.Write([]byte("[" + id + "] Something went wrong"))
		logger.Errf(ctx, "%s", err.Error())
	}
}

type stringEncoder struct{ w io.Writer }

func (e *stringEncoder) Encode(v interface{}) error {
	var err error
	switch c := v.(type) {
	case string:
		_, err = e.w.Write([]byte(c))
	case *string: // v may be a string pointer when the Response Body is set to the field of a custom response type.
		if c != nil {
			_, err = e.w.Write([]byte(*c))
		}
	case []byte:
		_, err = e.w.Write(c)
	default:
		err = fmt.Errorf("stringEncoder can't encode %T", c)
	}
	return err
}

// encoder is a copy of goa's default encoder, except it does not use go's XML Encoder.
// gorilla/feeds already provides correct xml, so this avoids decoding xml twice

func encoder(ctx context.Context, w http.ResponseWriter) goahttp.Encoder {
	negotiate := func(a string) (goahttp.Encoder, string) {
		switch a {
		case "", "application/json":
			// default to JSON
			return json.NewEncoder(w), "application/json"
		case "application/xml":
			return &stringEncoder{w}, "application/xml"
		case "application/gob":
			return gob.NewEncoder(w), "application/gob"
		case "text/html":
			return &stringEncoder{w}, "text/html"
		default:
			return nil, ""
		}
	}
	var accept string

	if a := ctx.Value(goahttp.AcceptTypeKey); a != nil {
		accept = a.(string)
	}

	var ct string

	if a := ctx.Value(goahttp.ContentTypeKey); a != nil {
		ct = a.(string)
	}

	var (
		enc goahttp.Encoder
		mt  string
		err error
	)

	if ct != "" {
		// If content type explicitly set in the DSL, infer the response encoder
		// from the content type context key.
		if mt, _, err = mime.ParseMediaType(ct); err == nil {
			enc, _ = negotiate(ct)
		}
	} else {
		// If Accept header exists in the request, infer the response encoder
		// from the header value.
		if enc, mt = negotiate(accept); enc == nil {
			// attempt to normalize
			if mt, _, err = mime.ParseMediaType(accept); err == nil {
				enc, mt = negotiate(mt)
			}
		}

	}
	if enc == nil {
		enc, mt = negotiate("")
	}

	goahttp.SetContentType(w, mt)
	return enc
}
