// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/openfaas/faas-provider/auth"
)

// ControllerConfig configures a connector SDK controller
type ControllerConfig struct {
	// UpstreamTimeout controls maximum timeout invoking a function via the gateway
	UpstreamTimeout time.Duration

	//  GatewayURL is the remote OpenFaaS gateway
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
}

// Controller is used to invoke functions on a per-topic basis and to subscribe to responses returned by said functions.
type Controller interface {
	Subscribe(subscriber ResponseSubscriber)
	Invoke(topic string, message *[]byte, headers http.Header)
	InvokeWithContext(ctx context.Context, topic string, message *[]byte, headers http.Header)
	BeginMapBuilder()
	Topics() []string
}

// controller is the default implementation of the Controller interface.
type controller struct {
	// Config for the controller
	Config *ControllerConfig

	// Invoker to invoke functions via HTTP(s)
	Invoker *Invoker

	// Map of which functions subscribe to which topics
	TopicMap *TopicMap

	// Credentials to access gateway
	Credentials *auth.BasicAuthCredentials

	// Subscribers which can receive messages from invocations.
	// See note on ResponseSubscriber interface about blocking/long-running
	// operations
	Subscribers []ResponseSubscriber

	// Lock used for synchronizing subscribers
	Lock *sync.RWMutex
}

// NewController create a new connector SDK controller
func NewController(credentials *auth.BasicAuthCredentials, config *ControllerConfig) Controller {

	gatewayFunctionPath := gatewayRoute(config)

	invoker := NewInvoker(gatewayFunctionPath,
		MakeClient(config.UpstreamTimeout),
		config.ContentType,
		config.PrintResponse,
		config.PrintRequestBody)

	subs := []ResponseSubscriber{}

	topicMap := NewTopicMap()

	c := controller{
		Config:      config,
		Invoker:     invoker,
		TopicMap:    &topicMap,
		Credentials: credentials,
		Subscribers: subs,
		Lock:        &sync.RWMutex{},
	}

	if config.PrintResponse {
		// printer := &{}
		c.Subscribe(&ResponsePrinter{config.PrintResponseBody})
	}

	go func(ch *chan InvokerResponse, controller *controller) {
		for {
			res := <-*ch

			controller.Lock.RLock()
			for _, sub := range controller.Subscribers {
				sub.Response(res)
			}
			controller.Lock.RUnlock()
		}
	}(&invoker.Responses, &c)

	return &c
}

// Subscribe adds a ResponseSubscriber to the list of subscribers
// which receive messages upon function invocation or error
// Note: it is not possible to Unsubscribe at this point using
// the API of the controller
func (c *controller) Subscribe(subscriber ResponseSubscriber) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Subscribers = append(c.Subscribers, subscriber)
}

// Invoke attempts to invoke any functions which match the
// topic the incoming message was published on.
func (c *controller) Invoke(topic string, message *[]byte, headers http.Header) {
	c.InvokeWithContext(context.Background(), topic, message, headers)
}

// InvokeWithContext attempts to invoke any functions which match the topic
// the incoming message was published on while propagating context.
func (c *controller) InvokeWithContext(ctx context.Context, topic string, message *[]byte, headers http.Header) {
	c.Invoker.InvokeWithContext(ctx, c.TopicMap, topic, message, headers)
}

// BeginMapBuilder begins to build a map of function->topic by
// querying the API gateway.
func (c *controller) BeginMapBuilder() {

	lookupBuilder := FunctionLookupBuilder{
		GatewayURL:     c.Config.GatewayURL,
		Client:         MakeClient(c.Config.UpstreamTimeout),
		Credentials:    c.Credentials,
		TopicDelimiter: c.Config.TopicAnnotationDelimiter,
	}

	ticker := time.NewTicker(c.Config.RebuildInterval)
	go c.synchronizeLookups(ticker, &lookupBuilder, c.TopicMap)
}

func (c *controller) synchronizeLookups(ticker *time.Ticker,
	lookupBuilder *FunctionLookupBuilder,
	topicMap *TopicMap) {

	fn := func() {
		lookups, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		if c.Config.PrintSync {
			log.Println("Syncing topic map")
		}

		topicMap.Sync(&lookups)
	}

	fn()
	for {
		<-ticker.C
		fn()
	}
}

// Topics gets the list of topics that functions have indicated should
// be used as triggers.
func (c *controller) Topics() []string {
	return c.TopicMap.Topics()
}

func gatewayRoute(config *ControllerConfig) string {
	if config.AsyncFunctionInvocation {
		return fmt.Sprintf("%s/%s", config.GatewayURL, "async-function")
	}
	return fmt.Sprintf("%s/%s", config.GatewayURL, "function")
}
