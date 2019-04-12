// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// feedgen HTTP server types
//
// Command:
// $ goa gen github.com/danlock/go-rss-gen/design

package server

import (
	feedgen "github.com/danlock/go-rss-gen/gen/feedgen"
	goa "goa.design/goa"
)

// MangaRequestBody is the type of the "feedgen" service "manga" endpoint HTTP
// request body.
type MangaRequestBody struct {
	// List of manga titles to subscribe to
	Titles []string `form:"titles,omitempty" json:"titles,omitempty" xml:"titles,omitempty"`
}

// MangaNotFoundResponseBody is the type of the "feedgen" service "manga"
// endpoint HTTP response body for the "NotFound" error.
type MangaNotFoundResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// MangaBadGatewayResponseBody is the type of the "feedgen" service "manga"
// endpoint HTTP response body for the "BadGateway" error.
type MangaBadGatewayResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// MangaInternalServerErrorResponseBody is the type of the "feedgen" service
// "manga" endpoint HTTP response body for the "InternalServerError" error.
type MangaInternalServerErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ViewMangaNotFoundResponseBody is the type of the "feedgen" service
// "viewManga" endpoint HTTP response body for the "NotFound" error.
type ViewMangaNotFoundResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ViewMangaBadGatewayResponseBody is the type of the "feedgen" service
// "viewManga" endpoint HTTP response body for the "BadGateway" error.
type ViewMangaBadGatewayResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ViewMangaInternalServerErrorResponseBody is the type of the "feedgen"
// service "viewManga" endpoint HTTP response body for the
// "InternalServerError" error.
type ViewMangaInternalServerErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// NewMangaNotFoundResponseBody builds the HTTP response body from the result
// of the "manga" endpoint of the "feedgen" service.
func NewMangaNotFoundResponseBody(res *goa.ServiceError) *MangaNotFoundResponseBody {
	body := &MangaNotFoundResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMangaBadGatewayResponseBody builds the HTTP response body from the result
// of the "manga" endpoint of the "feedgen" service.
func NewMangaBadGatewayResponseBody(res *goa.ServiceError) *MangaBadGatewayResponseBody {
	body := &MangaBadGatewayResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMangaInternalServerErrorResponseBody builds the HTTP response body from
// the result of the "manga" endpoint of the "feedgen" service.
func NewMangaInternalServerErrorResponseBody(res *goa.ServiceError) *MangaInternalServerErrorResponseBody {
	body := &MangaInternalServerErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewViewMangaNotFoundResponseBody builds the HTTP response body from the
// result of the "viewManga" endpoint of the "feedgen" service.
func NewViewMangaNotFoundResponseBody(res *goa.ServiceError) *ViewMangaNotFoundResponseBody {
	body := &ViewMangaNotFoundResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewViewMangaBadGatewayResponseBody builds the HTTP response body from the
// result of the "viewManga" endpoint of the "feedgen" service.
func NewViewMangaBadGatewayResponseBody(res *goa.ServiceError) *ViewMangaBadGatewayResponseBody {
	body := &ViewMangaBadGatewayResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewViewMangaInternalServerErrorResponseBody builds the HTTP response body
// from the result of the "viewManga" endpoint of the "feedgen" service.
func NewViewMangaInternalServerErrorResponseBody(res *goa.ServiceError) *ViewMangaInternalServerErrorResponseBody {
	body := &ViewMangaInternalServerErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMangaPayload builds a feedgen service manga endpoint payload.
func NewMangaPayload(body *MangaRequestBody, feedType string) *feedgen.MangaPayload {
	v := &feedgen.MangaPayload{}
	v.Titles = make([]string, len(body.Titles))
	for i, val := range body.Titles {
		v.Titles[i] = val
	}
	v.FeedType = feedType
	return v
}

// NewViewMangaPayload builds a feedgen service viewManga endpoint payload.
func NewViewMangaPayload(hash string) *feedgen.ViewMangaPayload {
	return &feedgen.ViewMangaPayload{
		Hash: hash,
	}
}

// ValidateMangaRequestBody runs the validations defined on MangaRequestBody
func ValidateMangaRequestBody(body *MangaRequestBody) (err error) {
	if body.Titles == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("titles", "body"))
	}
	if len(body.Titles) < 1 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("body.titles", body.Titles, len(body.Titles), 1, true))
	}
	if len(body.Titles) > 2048 {
		err = goa.MergeErrors(err, goa.InvalidLengthError("body.titles", body.Titles, len(body.Titles), 2048, false))
	}
	return
}
