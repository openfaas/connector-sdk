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
			Context: ctx,
			Error:   fmt.Errorf("no message to send"),
		}
	}

	matchedFunctions := topicMap.Match(topic)
	for _, matchedFunction := range matchedFunctions {
		log.Printf("Invoke function: %s", matchedFunction)

		gwURL := fmt.Sprintf("%s/%s", i.GatewayURL, matchedFunction)
		reader := bytes.NewReader(*message)

		if i.PrintRequest {
			fmt.Printf("[invoke] %s => %s\n\t%s\n", topic, matchedFunction, string(*message))
		}

		body, statusCode, header, err := invokefunction(ctx, i.Client, gwURL, i.ContentType, topic, reader, headers)
		if err != nil {
			i.Responses <- InvokerResponse{
				Context: ctx,
				Error:   fmt.Errorf("unable to invoke %s, error: %w", matchedFunction, err),
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
		}
	}
}

func invokefunction(ctx context.Context, c *http.Client, gwURL, contentType, topic string, reader io.Reader, headers http.Header) (*[]byte, int, *http.Header, error) {

	httpReq, err := http.NewRequest(http.MethodPost, gwURL, reader)
	if err != nil {
		return nil, http.StatusServiceUnavailable, nil, err
	}

	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	httpReq.Header.Add("X-Topic", topic)

	for k, values := range headers {
		for _, value := range values {
			httpReq.Header.Add(k, value)
		}
	}

	httpReq = httpReq.WithContext(ctx)
	if httpReq.Body != nil {
		defer httpReq.Body.Close()
	}

	var body *[]byte

	res, doErr := c.Do(httpReq)
	if doErr != nil {
		return nil, http.StatusServiceUnavailable, nil, doErr
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Printf("Error reading body")
			return nil, http.StatusServiceUnavailable, nil, doErr

		}
		body = &bytesOut
	}

	return body, res.StatusCode, &res.Header, doErr
}
