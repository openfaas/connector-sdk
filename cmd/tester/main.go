// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/openfaas/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
)

func main() {

	var (
		username,
		password,
		gateway,
		topic string
	)

	flag.StringVar(&username, "username", "admin", "username")
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&gateway, "gateway", "http://127.0.0.1:8080", "gateway")
	flag.StringVar(&topic, "topic", "payment.received", "Sample topic name to emit from timer")

	flag.Parse()

	creds := &auth.BasicAuthCredentials{
		User:     username,
		Password: password,
	}

	config := &types.ControllerConfig{
		RebuildInterval:         time.Second * 10,
		GatewayURL:              gateway,
		PrintResponse:           true,
		PrintRequestBody:        true,
		PrintResponseBody:       true,
		AsyncFunctionInvocation: false,
		ContentType:             "text/plain",
	}

	controller := types.NewController(creds, config)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()

	additionalHeaders := http.Header{}
	additionalHeaders.Add("X-Connector", "cmd/tester")

	// Simulate events emitting from queue/pub-sub
	messageID := 0
	for {
		log.Printf("Emitting event on topic payment.received - %s\n", gateway)
		time.Sleep(5 * time.Second)

		// Add a de-dupe header to the message
		additionalHeaders.Add("X-Message-Id", fmt.Sprintf("%d", messageID))

		eventData := []byte("test " + time.Now().String())
		controller.Invoke(topic, &eventData, additionalHeaders)

		messageID++
	}
}

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
		log.Printf("tester got result: [%d] %s => %s (%d) bytes (%fs)", res.Status, res.Topic, res.Function, len(*res.Body), res.Duration.Seconds())
	}
}
