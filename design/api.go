package design

import (
	"goa.design/goa/dsl"
)

// API describes the global properties of the API server.
var _ = dsl.API("feed-gen", func() {
	dsl.Title("Feed Generator")
	dsl.Description("Generates feed for different types of media")
})

var ResponseMap = map[int]func(){
	dsl.StatusNotFound:            func() { dsl.Response("NotFound", dsl.StatusNotFound) },
	dsl.StatusBadGateway:          func() { dsl.Response("BadGateway", dsl.StatusBadGateway) },
	dsl.StatusInternalServerError: func() { dsl.Response("InternalServerError", dsl.StatusInternalServerError) },
}

var ErrorMap = map[int]func(){
	dsl.StatusNotFound:            func() { dsl.Error("NotFound") },
	dsl.StatusBadGateway:          func() { dsl.Error("BadGateway") },
	dsl.StatusInternalServerError: func() { dsl.Error("InternalServerError") },
}

func SetErrors(funcMap map[int]func(), errs ...int) {
	for _, e := range errs {
		if funcMap[e] != nil {
			funcMap[e]()
		}
	}
}

// Service describes a service
var _ = dsl.Service("feedgen", func() {
	dsl.HTTP(func() {
		dsl.Path("/feed")
	})
	SetErrors(ErrorMap, dsl.StatusNotFound, dsl.StatusBadGateway, dsl.StatusInternalServerError)

	dsl.Method("manga", func() {
		dsl.Payload(func() {
			dsl.Attribute("feedType", dsl.String, "RSS, Atom, or JSON Feed", func() {
				dsl.Enum("rss", "atom", "json")
				dsl.Default("json")
			})
			dsl.Attribute("titles", dsl.ArrayOf(dsl.String), "List of manga titles to subscribe to", func() {
				dsl.MinLength(1)
				dsl.MaxLength(2048)
			})
			dsl.Required("titles")
		})
		dsl.Result(dsl.String, "path for desired feed")
		dsl.HTTP(func() {
			dsl.POST("/manga")
			dsl.Param("feedType")
			dsl.Response(dsl.StatusOK)
			SetErrors(ResponseMap, dsl.StatusNotFound, dsl.StatusBadGateway, dsl.StatusInternalServerError)
		})
	})
	dsl.Method("viewManga", func() {
		dsl.Payload(func() {
			dsl.Attribute("hash", dsl.String, "Identifier of previously created manga feed")
			dsl.Required("hash")
		})
		dsl.Result(func() {
			dsl.Attribute("feed", dsl.Bytes)
			dsl.Attribute("contentType", dsl.String)
			dsl.Required("feed", "contentType")
		})
		dsl.HTTP(func() {
			dsl.GET("/manga/{hash}")
			dsl.Param("hash")
			dsl.Response(dsl.StatusOK, func() {
				dsl.Header("contentType:Content-Type")
				dsl.Body("feed")
			})
			SetErrors(ResponseMap, dsl.StatusNotFound, dsl.StatusBadGateway, dsl.StatusInternalServerError)
		})
	})
})
