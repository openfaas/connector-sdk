package types

import (
	"log"
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

	// PrintResponseBody if true prints the function response body to stdout
	PrintResponseBody bool

	// RebuildInterval the interval at which the topic map is rebuilt
	RebuildInterval time.Duration
}

// Controller for the connector SDK
type Controller struct {
	// Config for the Controller
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
func NewController(credentials *auth.BasicAuthCredentials, config *ControllerConfig) *Controller {

	invoker := NewInvoker(config.GatewayURL,
		MakeClient(config.UpstreamTimeout),
		config.PrintResponse)

	subs := []ResponseSubscriber{}

	topicMap := NewTopicMap()

	controller := Controller{
		Config:      config,
		Invoker:     invoker,
		TopicMap:    &topicMap,
		Credentials: credentials,
		Subscribers: subs,
		Lock:        &sync.RWMutex{},
	}

	if config.PrintResponse {
		// printer := &{}
		controller.Subscribe(&ResponsePrinter{config.PrintResponseBody})
	}

	go func(ch *chan InvokerResponse, controller *Controller) {
		for {
			res := <-*ch

			controller.Lock.RLock()
			for _, sub := range controller.Subscribers {
				sub.Response(res)
			}
			controller.Lock.RUnlock()
		}
	}(&invoker.Responses, &controller)

	return &controller
}

// Subscribe adds a ResponseSubscriber to the list of subscribers
// which receive messages upon function invocation or error
// Note: it is not possible to Unsubscribe at this point using
// the API of the Controller
func (c *Controller) Subscribe(subscriber ResponseSubscriber) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Subscribers = append(c.Subscribers, subscriber)
}

// Invoke attempts to invoke any functions which match the
// topic the incoming message was published on.
func (c *Controller) Invoke(topic string, message *[]byte) {
	c.Invoker.Invoke(c.TopicMap, topic, message)
}

// BeginMapBuilder begins to build a map of function->topic by
// querying the API gateway.
func (c *Controller) BeginMapBuilder() {

	lookupBuilder := FunctionLookupBuilder{
		GatewayURL:  c.Config.GatewayURL,
		Client:      MakeClient(c.Config.UpstreamTimeout),
		Credentials: c.Credentials,
	}

	ticker := time.NewTicker(c.Config.RebuildInterval)
	go synchronizeLookups(ticker, &lookupBuilder, c.TopicMap)
}

func synchronizeLookups(ticker *time.Ticker,
	lookupBuilder *FunctionLookupBuilder,
	topicMap *TopicMap) {

	for {
		<-ticker.C
		lookups, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Syncing topic map")
		topicMap.Sync(&lookups)
	}
}

// Topics gets the list of topics that functions have indicated should
// be used as triggers.
func (c *Controller) Topics() []string {
	return c.TopicMap.Topics()
}
