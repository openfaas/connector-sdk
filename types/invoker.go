// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Invoker is used to send requests to functions. Responses are
// returned via the Responses channel.
type Invoker struct {
	PrintResponse bool
	PrintRequest  bool
	Client        *http.Client
	GatewayURL    string
	ContentType   string
	Responses     chan InvokerResponse
}

// InvokerResponse is a wrapper to contain the response or error the Invoker
// receives from the function. Networking errors wil be found in the Error field.
type InvokerResponse struct {
	Context  context.Context
	Body     *[]byte
	Header   *http.Header
	Status   int
	Error    error
	Topic    string
	Function string
	Duration time.Duration
}

// NewInvoker constructs an Invoker instance
func NewInvoker(gatewayURL string, client *http.Client, contentType string, printResponse, printRequest bool) *Invoker {
	return &Invoker{
		PrintResponse: printResponse,
		PrintRequest:  printRequest,
		Client:        client,
		GatewayURL:    gatewayURL,
		ContentType:   contentType,
		Responses:     make(chan InvokerResponse),
	}
}

// Invoke triggers a function by accessing the API Gateway
func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message *[]byte, headers http.Header) {
	i.InvokeWithContext(context.Background(), topicMap, topic, message, headers)
}

//InvokeWithContext triggers a function by accessing the API Gateway while propagating context
func (i *Invoker) InvokeWithContext(ctx context.Context, topicMap *TopicMap, topic string, message *[]byte, headers http.Header) {
	if len(*message) == 0 {
		i.Responses <- InvokerResponse{
			Context:  ctx,
			Error:    fmt.Errorf("no message to send"),
			Duration: time.Millisecond * 0,
		}
	}

	matchedFunctions := topicMap.Match(topic)
	for _, matchedFunction := range matchedFunctions {
		log.Printf("Invoking: %s", matchedFunction)

		gwURL := fmt.Sprintf("%s/%s", i.GatewayURL, matchedFunction)
		reader := bytes.NewReader(*message)

		if i.PrintRequest {
			fmt.Printf("[invoke] %s => %s\n\t%s\n", topic, matchedFunction, string(*message))
		}

		start := time.Now()
		body, statusCode, header, err := invoke(ctx, i.Client, gwURL, i.ContentType, topic, reader, headers)
		if err != nil {
			i.Responses <- InvokerResponse{
				Context:  ctx,
				Error:    fmt.Errorf("unable to invoke %s, error: %w", matchedFunction, err),
				Duration: time.Since(start),
			}
			continue
		}

		i.Responses <- InvokerResponse{
			Context:  ctx,
			Body:     body,
			Status:   statusCode,
			Header:   header,
			Function: matchedFunction,
			Topic:    topic,
			Duration: time.Since(start),
		}
	}
}

func invoke(ctx context.Context, c *http.Client, gwURL, contentType, topic string, reader io.Reader, headers http.Header) (*[]byte, int, *http.Header, error) {
	req, err := http.NewRequest(http.MethodPost, gwURL, reader)
	if err != nil {
		return nil, http.StatusServiceUnavailable, nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if v := req.Header.Get("X-Connector"); v == "" {
		req.Header.Set("X-Connector", "connector-sdk")
	}

	if v := req.Header.Get("X-Topic"); v == "" {
		req.Header.Set("X-Topic", topic)
	}

	for k, values := range headers {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}

	req = req.WithContext(ctx)
	if req.Body != nil {
		defer req.Body.Close()
	}

	var body *[]byte
	res, err := c.Do(req)
	if err != nil {
		return nil, http.StatusServiceUnavailable, nil,
			fmt.Errorf("unable to reach endpoint %s, error: %w", gwURL, err)
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, http.StatusServiceUnavailable,
				nil,
				fmt.Errorf("unable to read body from response %w", err)
		}
		body = &bytesOut
	}

	return body, res.StatusCode, &res.Header, err
}
