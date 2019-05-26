package lib

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/danlock/feedgen/lib/logger"
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
	ct   string
}

func NewResponse(ctx context.Context, code int) *Response {
	return &Response{Context: ctx, code: code}
}

func (r *Response) WithMsg(msg string) *Response {
	r.msg = msg
	return r
}

func (r *Response) WithContent(ct string) *Response {
	r.ct = ct
	return r
}

// WriteResponse to the client
func (r *Response) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if len(r.msg) == 0 {
		rw.Header().Del("Content-Type") //Remove Content-Type on empty responses
		rw.WriteHeader(r.code)
		return
	}

	if r.ct != "" {
		rw.Header().Set("Content-Type", r.ct)
	}
	rw.WriteHeader(r.code)
	if err := producer.Produce(rw, r.msg); err != nil {
		logger.Errf(r, "Failed to write response msg %s err %+v", r.msg, err)
	}
}
