package logger

import (
	"context"
	"fmt"
	"log"
	"os"

	"goa.design/goa/middleware"
)

type logCtxKeyType struct{}

var logCtxKey logCtxKeyType

func AddLogCtx(ctx context.Context, s string) context.Context {
	oldLogCtxRaw := ctx.Value(logCtxKey)
	if oldLogCtxRaw == nil {
		return context.WithValue(ctx, logCtxKey, s)
	}
	oldLogCtx, ok := oldLogCtxRaw.(string)
	if !ok {
		panic("Got context with malformed log context value " + fmt.Sprintf("%+v", oldLogCtx))
	}
	return context.WithValue(ctx, logCtxKey, oldLogCtx+" "+s)
}

func AddLogCtxf(ctx context.Context, msg string, vals ...interface{}) context.Context {
	return AddLogCtx(ctx, fmt.Sprintf(msg, vals...))
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
	reqIDStr := ""
	reqID := ctx.Value(middleware.RequestIDKey)
	if reqID != nil {
		id, ok := reqID.(string)
		if ok {
			reqIDStr = "req_id:" + id + " "
		}
	}
	return fmt.Sprintf("%s%s", reqIDStr, logCtx)
}

func SetupLogger(prefix string) *log.Logger {
	lgr := log.New(os.Stderr, prefix, log.Ltime|log.LUTC|log.Lshortfile)
	log.SetFlags(lgr.Flags())
	log.SetPrefix(lgr.Prefix())
	return lgr
}

const calldepth = 3

func logf(ctx context.Context, prefix, msg string, vals ...interface{}) {
	log.Output(calldepth, fmt.Sprintf("%s %s %s", prefix, fmt.Sprintf(msg, vals...), getLogCtx(ctx)))
}

func Warnf(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, "[WRN]", msg, vals...)
}
func Infof(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, "[INF]", msg, vals...)
}
func Errf(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, "[ERR]", msg, vals...)
}

func Debugf(ctx context.Context, msg string, vals ...interface{}) {
	if IsDebug() {
		logf(ctx, "[DBG]", msg, vals...)
	}
}

// isDebug is set during build time using build flags, which can only be strings for some reason
var isDebug = "t"

func IsDebug() bool {
	return isDebug == "t"
}

type GoaLogAdapter struct{}

func (*GoaLogAdapter) Log(keyvals ...interface{}) error {
	Infof(context.Background(), "%+v", keyvals)
	return nil
}
