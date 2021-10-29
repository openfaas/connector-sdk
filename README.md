## connector-sdk

The connector-sdk is a library written in Go that you can use to create event-connectors for OpenFaaS functions.

How it works:

* Create a new CLI application with Go
* Add the code to subscribe to events or messages from your source - i.e. a database, webhooks, message queue
* Add the `github.com/openfaas/connector-sdk/types` package to your code
* Setup a `types.ControllerConfig` with the gateway location
* Run `types.NewController`
* Whenever you receive a message from your source, run `controller.Invoke(topic, data)`

Then whichever functions in your cluster have matching annotation of "topic: topic" will be invoked.

### Conceptual design:

![Conceptual design](https://pbs.twimg.com/media/DrlGTNtWkAEGbnQ.jpg)

> Each function expresses which topics it can be triggered by, the broker then invokes them using the SDK.

See also: [Triggers in OpenFaaS](https://docs.openfaas.com/reference/triggers/)

## Creating your own connector

See the examples in the [OpenFaaS Docs](https://docs.openfaas.com/reference/triggers/) for inspiration.

You can copy one of them and adapt it, or see the "tester" example in this repo.

The tester example doesn't have an event subscription, but a for loop and sleep combination which simulates receiving an event. You would replace the timer with the callback function from your source such as a HTTP webhook endpoint, a pub-sub SDK or likewise.

Within the event subscriber code, you should call "Invoke()", passing in the topic and message. The functions advertise their "topic".

```go

	additionalHeaders := http.Header{}
	additionalHeaders.Add("X-Served-By", "cmd/tester")

	// Simulate events emitting from queue/pub-sub
	for {
		log.Printf("Invoking on topic payment - %s\n", gateway)
		time.Sleep(2 * time.Second)
		data := []byte("test " + time.Now().String())
		controller.Invoke("payment", &data, additionalHeaders)
	}
```

The results can then be printed using a result receiver.

```go
// ResponseReceiver enables connector to receive results from the
// function invocation
type ResponseReceiver struct {
}

// Response is triggered by the controller when a message is
// received from the function invocation
func (ResponseReceiver) Response(res types.InvokerResponse) {
	if res.Error != nil {
		log.Printf("tester got error: %s", res.Error.Error())
	} else {
		log.Printf("tester got result: [%d] %s => %s (%d) bytes", res.Status, res.Topic, res.Function, len(*res.Body))
	}
}
```

There are no retry mechanisms at present, but you could use the receiver to requeue failed invocations, or to send on to a dead-letter queue (DLQ).

If you expect many requests in a short period of time, you may want to defer the executions using OpenFaaS' built-in asynchronous queue.

Set the following in `ControllerConfig`:

```go
	config := &types.ControllerConfig{
        ...
		AsyncFunctionInvocation: true,
	}
```

If you need to use the `Content-Type` header to validate or to check when it invoke the function you can set the `Content-Type`
in `ControllerConfig`:

```go
	config := &types.ControllerConfig{
        ...
		ContentType: "application/json",
	}
```

View the code: [cmd/tester/main.go](cmd/tester/main.go)

## License

MIT
