// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// feedgen HTTP server encoders and decoders
//
// Command:
// $ goa gen github.com/danlock/go-rss-gen/design

package server

import (
	"context"
	"io"
	"net/http"

	goa "goa.design/goa"
	goahttp "goa.design/goa/http"
)

// EncodeMangaResponse returns an encoder for responses returned by the feedgen
// manga endpoint.
func EncodeMangaResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(string)
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeMangaRequest returns a decoder for requests sent to the feedgen manga
// endpoint.
func DecodeMangaRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			body MangaRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = ValidateMangaRequestBody(&body)
		if err != nil {
			return nil, err
		}
		payload := NewMangaPayload(&body)

		return payload, nil
	}
}

// EncodeMangaError returns an encoder for errors returned by the manga feedgen
// endpoint.
func EncodeMangaError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "NotFound":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewMangaNotFoundResponseBody(res)
			w.Header().Set("goa-error", "NotFound")
			w.WriteHeader(http.StatusNotFound)
			return enc.Encode(body)
		case "BadGateway":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewMangaBadGatewayResponseBody(res)
			w.Header().Set("goa-error", "BadGateway")
			w.WriteHeader(http.StatusBadGateway)
			return enc.Encode(body)
		case "InternalServerError":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewMangaInternalServerErrorResponseBody(res)
			w.Header().Set("goa-error", "InternalServerError")
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}

// EncodeViewMangaResponse returns an encoder for responses returned by the
// feedgen viewManga endpoint.
func EncodeViewMangaResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, interface{}) error {
	return func(ctx context.Context, w http.ResponseWriter, v interface{}) error {
		res := v.(string)
		ctx = context.WithValue(ctx, goahttp.ContentTypeKey, "application/xml")
		enc := encoder(ctx, w)
		body := res
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// DecodeViewMangaRequest returns a decoder for requests sent to the feedgen
// viewManga endpoint.
func DecodeViewMangaRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (interface{}, error) {
		var (
			hash     string
			feedType string
			err      error

			params = mux.Vars(r)
		)
		hash = params["hash"]
		feedTypeRaw := r.URL.Query().Get("feedType")
		if feedTypeRaw != "" {
			feedType = feedTypeRaw
		} else {
			feedType = "rss"
		}
		if !(feedType == "rss" || feedType == "atom") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("feedType", feedType, []interface{}{"rss", "atom"}))
		}
		if err != nil {
			return nil, err
		}
		payload := NewViewMangaPayload(hash, feedType)

		return payload, nil
	}
}

// EncodeViewMangaError returns an encoder for errors returned by the viewManga
// feedgen endpoint.
func EncodeViewMangaError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		en, ok := v.(ErrorNamer)
		if !ok {
			return encodeError(ctx, w, v)
		}
		switch en.ErrorName() {
		case "NotFound":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewViewMangaNotFoundResponseBody(res)
			w.Header().Set("goa-error", "NotFound")
			w.WriteHeader(http.StatusNotFound)
			return enc.Encode(body)
		case "BadGateway":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewViewMangaBadGatewayResponseBody(res)
			w.Header().Set("goa-error", "BadGateway")
			w.WriteHeader(http.StatusBadGateway)
			return enc.Encode(body)
		case "InternalServerError":
			res := v.(*goa.ServiceError)
			enc := encoder(ctx, w)
			body := NewViewMangaInternalServerErrorResponseBody(res)
			w.Header().Set("goa-error", "InternalServerError")
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
