// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"time"

	"github.com/openfaas-incubator/connector-sdk/types"
)

func main() {
	creds := types.GetCredentials()
	config := &types.ControllerConfig{
		RebuildInterval: time.Millisecond * 1000,
		GatewayURL:      "http://127.0.0.1:8080",
		PrintResponse:   true,
	}

	controller := types.NewController(creds, config)
	fmt.Println(controller)
	controller.BeginMapBuilder()

	// Simulate events emitting from queue/pub-sub
	for {
		time.Sleep(2 * time.Second)
		data := []byte("test " + time.Now().String())
		controller.Invoke("test", &data)
	}
}
