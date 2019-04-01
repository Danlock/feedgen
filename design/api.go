package design

import "goa.design/goa/dsl"

// API describes the global properties of the API server.
var _ = dsl.API("rss-gen", func() {
	dsl.Title("RSS Generator")
	dsl.Description("Generates RSS feed for different types of media")
})

// Service describes a service
var _ = dsl.Service("feedgen", func() {
	dsl.HTTP(func() {
		dsl.Path("/feed")
	})
	dsl.Method("manga", func() {
		dsl.Payload(func() {
			dsl.Attribute("titles", dsl.ArrayOf(dsl.String), "List of manga titles to subscribe to", func() {
				dsl.MinLength(1)
			})
			dsl.Required("titles")
		})
		dsl.Result(dsl.String)
		dsl.HTTP(func() {
			dsl.POST("/manga")
			dsl.Response(dsl.StatusOK)
			dsl.Response(dsl.StatusNotFound)
		})
	})
})
