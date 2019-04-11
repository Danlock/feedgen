package design

import "goa.design/goa/dsl"

// API describes the global properties of the API server.
var _ = dsl.API("feed-gen", func() {
	dsl.Title("Feed Generator")
	dsl.Description("Generates feed for different types of media")
})

// Service describes a service
var _ = dsl.Service("feedgen", func() {
	dsl.HTTP(func() {
		dsl.Path("/feed")
	})
	dsl.Method("manga", func() {
		dsl.Payload(func() {
			dsl.Attribute("feedType", dsl.String, "RSS, Atom, or JSON Feed", func() {
				dsl.Enum("rss", "atom", "json")
				dsl.Default("json")
			})
			dsl.Attribute("titles", dsl.ArrayOf(dsl.String), "List of manga titles to subscribe to", func() {
				dsl.MinLength(1)
				dsl.MaxLength(65535)
			})
			dsl.Required("titles")
		})
		dsl.Result(dsl.String, "path for desired feed")
		dsl.HTTP(func() {
			dsl.POST("/manga")
			dsl.Param("feedType")
			dsl.Response(dsl.StatusOK)
			dsl.Response(dsl.StatusNotFound)
		})
	})
	dsl.Method("viewManga", func() {
		dsl.Payload(func() {
			dsl.Attribute("hash", dsl.String, "Identifier of previously created manga feed")
			dsl.Required("hash")
		})
		dsl.Result(dsl.String, "Feed")
		dsl.HTTP(func() {
			dsl.GET("/manga/{hash}")
			dsl.Param("hash")
			dsl.Response(dsl.StatusOK)
			dsl.Response(dsl.StatusNotFound)
		})
	})
})
