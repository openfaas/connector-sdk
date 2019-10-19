// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"log"
	"time"

	"github.com/openfaas-incubator/connector-sdk/types"
	"github.com/openfaas/faas-provider/auth"
)

func main() {

	var username, password, gateway string

	flag.StringVar(&username, "username", "admin", "username")
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&gateway, "gateway", "http://127.0.0.1:8080", "gateway")

	flag.Parse()

	creds := &auth.BasicAuthCredentials{
		User:     username,
		Password: password,
	}

	config := &types.ControllerConfig{
		RebuildInterval:   time.Millisecond * 1000,
		GatewayURL:        gateway,
		PrintResponse:     true,
		PrintResponseBody: true,
	}

	controller := types.NewController(creds, config)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()

	// Simulate events emitting from queue/pub-sub
	for {
		log.Printf("Invoking on topic vm.powered.on - %s\n", gateway)
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
