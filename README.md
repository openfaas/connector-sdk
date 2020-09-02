## connector-sdk

The connector-sdk is a library written in Go that you can use to create event-connectors for OpenFaaS functions.

How it works:

* Create a new CLI application with Go
* Add the code to subscribe to events or messages from your source - i.e. a database, webhooks, message queue
* Add the `github.com/openfaas-incubator/connector-sdk/types` package to your code
* Setup a `types.ControllerConfig` with the gateway location
* Run `types.NewController`
* Whenever you receive a message from your source, run `controller.Invoke(topic, data)`

Then whichever functions in your cluster have matching annotation of "topic: topic" will be invoked.

You can see a full example which is triggered by a for loop and a timer here: [cmd/tester/main.go](cmd/tester/main.go)

### Conceptual design:

![Conceptual design](https://pbs.twimg.com/media/DrlGTNtWkAEGbnQ.jpg)

> Each function expresses which topics it can be triggered by, the broker then invokes them using the SDK.

See also: [Triggers in OpenFaaS](https://docs.openfaas.com/reference/triggers/)

### Development

To test your changes, please run `make`

```sh
make
```

