// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"encoding/json"
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
		interval time.Duration
	)

	flag.StringVar(&username, "username", "admin", "username")
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&gateway, "gateway", "http://127.0.0.1:8080", "gateway")
	flag.DurationVar(&interval, "interval", time.Second*10, "Interval between emitting a sample message")
	flag.StringVar(&topic, "topic", "payment.received", "Sample topic name to emit from timer")

	flag.Parse()

	if len(password) == 0 {
		password = lookupPasswordViaKubectl()
	}

	creds := &auth.BasicAuthCredentials{
		User:     username,
		Password: password,
	}

	// Set Print* variables to false for production use

	config := &types.ControllerConfig{
		RebuildInterval:         time.Second * 30,
		GatewayURL:              gateway,
		PrintResponse:           true,
		PrintRequestBody:        true,
		PrintResponseBody:       true,
		AsyncFunctionInvocation: false,
		ContentType:             "text/plain",
		UserAgent:               "openfaasltd/timer-connector",
		UpstreamTimeout:         time.Second * 120,
	}

	fmt.Printf("Tester connector. Topic: %s, Interval: %s\n", topic, interval)

	controller := types.NewController(creds, config)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()

	additionalHeaders := http.Header{}
	additionalHeaders.Add("X-Connector", "cmd/timer")

	// Simulate events emitting from queue/pub-sub
	// by sleeping for 10 seconds between emitting the same message
	messageID := 0

	t := time.NewTicker(interval)
	for {
		<-t.C

		log.Printf("[tester] Emitting event on topic payment.received - %s\n", gateway)

		h := additionalHeaders.Clone()
		// Add a de-dupe header to the message
		h.Add("X-Message-Id", fmt.Sprintf("%d", messageID))

		payload, _ := json.Marshal(samplePayload{
			CreatedAt: time.Now(),
			MessageID: messageID,
		})

		controller.Invoke(topic, &payload, h)

		messageID++

		t.Reset(interval)
	}
}

type samplePayload struct {
	CreatedAt time.Time `json:"createdAt"`
	MessageID int       `json:"messageId"`
}

// ResponseReceiver enables connector to receive results from the
// function invocation
type ResponseReceiver struct {
}

// Response is triggered by the controller when a message is
// received from the function invocation
func (ResponseReceiver) Response(res types.InvokerResponse) {
	if res.Error != nil {
		log.Printf("[tester] error: %s", res.Error.Error())
	} else {
		log.Printf("[tester] result: [%d] %s => %s (%d) bytes (%fs)", res.Status, res.Topic, res.Function, len(*res.Body), res.Duration.Seconds())
	}
}
