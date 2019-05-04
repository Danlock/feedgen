package lib

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/danlock/go-rss-gen/lib/logger"

	"github.com/go-openapi/runtime"
)

func GetEnvOrWarn(k string) (v string) {
	if v = os.Getenv(k); v == "" {
		log.Printf("Missing env %s", k)
	}
	return v
}

// SentinelError is a simple error based around a string so it can be made const.
type SentinelError string

func (s SentinelError) Error() string { return string(s) }

type Response struct {
	context.Context
	msg  string
	code int
}

func NewResponse(ctx context.Context, code int) *Response {
	return NewResponseWithMsg(ctx, code, "")
}
func NewResponseWithMsg(ctx context.Context, code int, msg string) *Response {
	return &Response{Context: ctx, msg: msg, code: code}
}

// WriteResponse to the client
func (r *Response) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if len(r.msg) == 0 {
		rw.Header().Del("Content-Type") //Remove Content-Type on empty responses
		rw.WriteHeader(r.code)
		return
	}
	rw.WriteHeader(r.code)
	if err := producer.Produce(rw, r.msg); err != nil {
		logger.Errf(r, "Failed to write response msg %s err %+v", r.msg, err)
	}
}
