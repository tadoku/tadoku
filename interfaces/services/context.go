//go:generate gex mockgen -source=context.go -package services -destination=context_mock.go

package services

// based on https://github.com/labstack/echo/blob/a2d4cb9c7a629e2ee21861501690741d2374de10/context.go

// Context is a subset of the echo framework context, so we are not directly depending on it
type Context interface {
	// Get retrieves data from the context.
	Get(key string) interface{}

	// Set saves data in the context.
	Set(key string, val interface{})

	// Bind binds the request body into provided type `i`. The default binder
	// does it based on Content-Type header.
	Bind(i interface{}) error

	// String sends a string response with status code.
	String(code int, s string) error

	// NoContent sends a response with no body and a status code.
	NoContent(code int) error

	// JSON sends a JSON response with status code.
	JSON(code int, i interface{}) error
}
