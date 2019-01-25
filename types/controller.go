package types

import (
	"log"
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

	// RebuildInterval the interval at which the topic map is rebuilt
	RebuildInterval time.Duration
}

// Controller for the connector SDK
type Controller struct {
	Config      *ControllerConfig
	Invoker     *Invoker
	TopicMap    *TopicMap
	Credentials *auth.BasicAuthCredentials
	Subscribers []ResponseSubscriber
}

// Subscribe adds a ResponseSubscriber to the list of subscribers
// which receive messages upon function invocation or error
func (c *Controller) Subscribe(subscriber ResponseSubscriber) {
	c.Subscribers = append(c.Subscribers, subscriber)
}

// ResponseSubscriber enables connector or another client in connector
// to receive results from the function invocation
type ResponseSubscriber interface {
	// Response is triggered by the controller when a message is
	// received from the function invocation
	Response(InvokerResponse)
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
	}

	if config.PrintResponse {
		// printer := &{}
		controller.Subscribe(&ResponsePrinter{})
	}

	go func(ch *chan InvokerResponse, controller *Controller) {
		for {
			res := <-*ch
			for _, sub := range controller.Subscribers {
				sub.Response(res)
			}
		}
	}(&invoker.Responses, &controller)

	return &controller
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
