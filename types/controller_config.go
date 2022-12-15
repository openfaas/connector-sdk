// ControllerConfig configures a connector SDK controller
package types

import "time"

type ControllerConfig struct {
	// UpstreamTimeout controls maximum timeout for a function invocation, which is done via the gateway
	UpstreamTimeout time.Duration

	// GatewayURL is the remote OpenFaaS gateway
	GatewayURL string

	// PrintResponse if true prints the function responses
	PrintResponse bool

	// PrintResponseBody prints the function's response body to stdout
	PrintResponseBody bool

	// PrintRequestBody prints the request's body to stdout.
	PrintRequestBody bool

	// RebuildInterval the interval at which the topic map is rebuilt
	RebuildInterval time.Duration

	// TopicAnnotationDelimiter defines the character upon which to split the Topic annotation value
	TopicAnnotationDelimiter string

	// AsyncFunctionInvocation if true points to the asynchronous function route
	AsyncFunctionInvocation bool

	// PrintSync indicates whether the sync should be logged.
	PrintSync bool

	// ContentType defines which content type will be set in the header to inkoke the function. i.e "application/json".
	// Optional, if not set the Content-Type header will not be set.
	ContentType string

	// BasicAuth whether basic auth is enabled or disabled
	BasicAuth bool

	// UserAgent defines the user agent to be used in the request to invoke the function, it should be of the format:
	// company/NAME-connector
	UserAgent string
}
