// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// FeedgenViewMangaHandlerFunc turns a function with the right signature into a feedgen view manga handler
type FeedgenViewMangaHandlerFunc func(FeedgenViewMangaParams) middleware.Responder

// Handle executing the request and returning a response
func (fn FeedgenViewMangaHandlerFunc) Handle(params FeedgenViewMangaParams) middleware.Responder {
	return fn(params)
}

// FeedgenViewMangaHandler interface for that can handle valid feedgen view manga params
type FeedgenViewMangaHandler interface {
	Handle(FeedgenViewMangaParams) middleware.Responder
}

// NewFeedgenViewManga creates a new http.Handler for the feedgen view manga operation
func NewFeedgenViewManga(ctx *middleware.Context, handler FeedgenViewMangaHandler) *FeedgenViewManga {
	return &FeedgenViewManga{Context: ctx, Handler: handler}
}

/*FeedgenViewManga swagger:route GET /api/feed/manga/{hash} feedgenViewManga

Get feed of manga updates

Returns an RSS/Atom/JSON Feed of the manga titles.

*/
type FeedgenViewManga struct {
	Context *middleware.Context
	Handler FeedgenViewMangaHandler
}

func (o *FeedgenViewManga) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewFeedgenViewMangaParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
