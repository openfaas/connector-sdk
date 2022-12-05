// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	execute "github.com/alexellis/go-execute/pkg/v1"
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

	if len(password) == 0 {
		password = lookupPasswordViaKubectl()
	}

	creds := &auth.BasicAuthCredentials{
		User:     username,
		Password: password,
	}

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

	controller := types.NewController(creds, config)

	receiver := ResponseReceiver{}
	controller.Subscribe(&receiver)

	controller.BeginMapBuilder()

	additionalHeaders := http.Header{}
	additionalHeaders.Add("X-Connector", "cmd/timer")

	// Simulate events emitting from queue/pub-sub
	// by sleeping for 10 seconds between emitting the same message
	messageID := 0
	for {
		log.Printf("Emitting event on topic payment.received - %s\n", gateway)

		h := additionalHeaders.Clone()
		// Add a de-dupe header to the message
		h.Add("X-Message-Id", fmt.Sprintf("%d", messageID))

		eventData := []byte("test " + time.Now().String())
		controller.Invoke(topic, &eventData, h)

		messageID++
		time.Sleep(10 * time.Second)
	}
}

func lookupPasswordViaKubectl() string {

	cmd := execute.ExecTask{
		Command:      "kubectl",
		Args:         []string{"get", "secret", "-n", "openfaas", "basic-auth", "-o", "jsonpath='{.data.basic-auth-password}'"},
		StreamStdio:  false,
		PrintCommand: false,
	}

	res, err := cmd.Execute()
	if err != nil {
		panic(err)
	}

	if res.ExitCode != 0 {
		panic("Non-zero exit code: " + res.Stderr)
	}
	resOut := strings.Trim(res.Stdout, "\\'")

	decoded, err := b64.StdEncoding.DecodeString(resOut)
	if err != nil {
		panic(err)
	}
	password := strings.TrimSpace(string(decoded))

	return password
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
