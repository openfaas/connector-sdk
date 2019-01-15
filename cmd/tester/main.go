// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"log"
	"time"

	"github.com/openfaas-incubator/connector-sdk/types"
)

func main() {
	creds := types.GetCredentials()
	config := &types.ControllerConfig{
		RebuildInterval:   time.Millisecond * 1000,
		GatewayURL:        "http://127.0.0.1:8080",
		PrintResponse:     true,
		PrintResponseBody: true,
	}

	controller := types.NewController(creds, config)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()

	// Simulate events emitting from queue/pub-sub
	for {
		time.Sleep(2 * time.Second)
		data := []byte("test " + time.Now().String())
		controller.Invoke("vm.powered.on", &data)
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
		log.Printf("tester got result: [%d] %s => %s (%d) bytes", res.Status, res.Topic, res.Function, len(*res.Body))
	}
}