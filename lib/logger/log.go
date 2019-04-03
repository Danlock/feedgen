package logger

import (
	"context"
	"fmt"
	"log"

	"github.com/danlock/go-rss-gen/lib"
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

func LogCtx(ctx context.Context) string {
	oldLogCtxRaw := ctx.Value(logCtxKey)
	if oldLogCtxRaw == nil {
		return ""
	}
	oldLogCtx, ok := oldLogCtxRaw.(string)
	if !ok {
		panic("Got context with malformed log context value " + fmt.Sprintf("%+v", oldLogCtx))
	}
	return oldLogCtx
}

func SetupLogger() {
	log.SetFlags(log.Ltime | log.LUTC | log.Lshortfile)
}

const calldepth = 4

func logf(ctx context.Context, prefix, msg string, vals ...interface{}) {
	log.Output(calldepth, fmt.Sprintf("%s %s %s", prefix, fmt.Sprintf(msg, vals...), LogCtx(ctx)))
}

func Warnf(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, "[WARN]", msg, vals...)
}
func Infof(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, "[INFO]", msg, vals...)
}
func Errf(ctx context.Context, msg string, vals ...interface{}) {
	logf(ctx, "[ERROR]", msg, vals...)
}

var (
	isDebug      = false
	checkedDebug = false
)

func Debugf(ctx context.Context, msg string, vals ...interface{}) {
	if !checkedDebug {
		checkedDebug = true
		isDebug = lib.GetEnvOrWarn("RSS_GEN_DEBUG") == "true"
	}
	if isDebug {
		logf(ctx, "[DEBUG]", msg, vals...)
	}
}
