package logger

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/felixge/httpsnoop"
)

type ctxKeyType uint

const (
	logCtxKey ctxKeyType = iota
)

// WithLogCtx appends a string to the context to be logged. Useful for capturing per request information.
func WithLogCtx(ctx context.Context, s string) context.Context {
	oldLogCtxRaw := ctx.Value(logCtxKey)
	if oldLogCtxRaw == nil {
		return context.WithValue(ctx, logCtxKey, "{ "+s)
	}
	oldLogCtx, ok := oldLogCtxRaw.(string)
	if !ok {
		panic("Got context with malformed log context value " + fmt.Sprintf("%+v", oldLogCtx))
	}
	return context.WithValue(ctx, logCtxKey, fmt.Sprintf("%s %s", oldLogCtx, s))
}

// WithLogCtxf appends a string to the context to be logged in fmt.sprintf format.
func WithLogCtxf(ctx context.Context, msg string, vals ...interface{}) context.Context {
	return WithLogCtx(ctx, fmt.Sprintf(msg, vals...))
}

func getLogCtx(ctx context.Context) string {
	logCtxRaw := ctx.Value(logCtxKey)
	if logCtxRaw == nil {
		return ""
	}
	logCtx, ok := logCtxRaw.(string)
	if !ok {
		panic("Got context with malformed log context value " + fmt.Sprintf("%+v", logCtx))
	}
	return logCtx + " }"
}

func SetupLogger(prefix string) *log.Logger {
	lgr := log.New(os.Stderr, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC|log.Lshortfile)
	log.SetFlags(lgr.Flags())
	log.SetPrefix(lgr.Prefix())
	return lgr
}

// calldepth is set so that the line number of the callee hitting *F functions will be reported accurately.
const calldepth = 3

func logf(ctx context.Context, cd int, prefix, msg string, vals ...interface{}) {
	log.Output(cd, fmt.Sprintf("%s %s %s", prefix, fmt.Sprintf(msg, vals...), getLogCtx(ctx)))
}
func Warnf(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, calldepth, "[WARN]", msg, vals...)
}
func Infof(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, calldepth, "[INFO]", msg, vals...)
}
func InfofWithCallDepth(ctx context.Context, cd int, msg string, vals ...interface{}) {
	logf(ctx, cd, "[INFO]", msg, vals...)
}
func Errf(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, calldepth, "[ERROR]", msg, vals...)
}

func Dbgf(ctx context.Context, msg string, vals ...interface{}) {
	if IsDebug() {
		logf(ctx, calldepth, "[DEBUG]", msg, vals...)
	}
}

var isDebug *bool

func IsDebug() bool {
	if isDebug == nil {
		if dbg, err := strconv.ParseBool(os.Getenv("FG_DEBUG")); err != nil {
			defaultDbg := true
			isDebug = &defaultDbg
		} else {
			isDebug = &dbg
		}
	}
	return *isDebug
}

var reqCounter uint64
var prefix string

func newReqID() string {
	if prefix == "" {
		prefixB := make([]byte, 4)
		rand.Read(prefixB)
		prefix = base64.RawURLEncoding.EncodeToString(prefixB)
	}
	id := fmt.Sprintf("%s-%d", prefix, reqCounter)
	atomic.AddUint64(&reqCounter, 1)
	return id
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqID := newReqID()
		if reqIDHdr := req.Header.Get("X-Request-Id"); reqIDHdr != "" {
			reqID = reqIDHdr
		}
		ctx := WithLogCtxf(req.Context(), "reqID: %s", reqID)
		req.WithContext(ctx)
		// Pass in modified req to DumpRequest to avoid printing all headers
		savedHeaders := req.Header
		req.Header = http.Header{}
		req.Header.Set("User-Agent", savedHeaders.Get("User-Agent"))
		req.Header.Set("Origin", savedHeaders.Get("Origin"))
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			Errf(ctx, "Failed to dump request info, aborting req as it is in an undefined state err:%+v", err)
			return
		}
		req.Header = savedHeaders
		Infof(ctx, "starting %s", reqDump)
		rw.Header().Set("X-Request-Id", reqID)
		metrics := httpsnoop.CaptureMetrics(next, rw, req)
		Infof(ctx, "completed %s %s in %s with status %d and wrote %d bytes", req.Method, req.RequestURI, metrics.Duration.String(), metrics.Code, metrics.Written)
	})
}
