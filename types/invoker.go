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

	"github.com/pkg/errors"
)

// Invoker is used to send requests to functions. Responses are
// returned via the Responses channel.
type Invoker struct {
	PrintResponse bool
	Client        *http.Client
	GatewayURL    string
	CallbackURL   string
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
func NewInvoker(gatewayURL, callbackURL string, client *http.Client, printResponse bool) *Invoker {
	return &Invoker{
		PrintResponse: printResponse,
		Client:        client,
		GatewayURL:    gatewayURL,
		CallbackURL:   callbackURL,
		Responses:     make(chan InvokerResponse),
	}
}

// Invoke triggers a function by accessing the API Gateway
func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message *[]byte) {
	i.InvokeWithContext(context.Background(), topicMap, topic, message)
}

//InvokeWithContext triggers a function by accessing the API Gateway while propagating context
func (i *Invoker) InvokeWithContext(ctx context.Context, topicMap *TopicMap, topic string, message *[]byte) {
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

		body, statusCode, header, doErr := invokefunction(ctx, i.Client, gwURL, i.CallbackURL, reader)

		if doErr != nil {
			i.Responses <- InvokerResponse{
				Context: ctx,
				Error:   errors.Wrap(doErr, fmt.Sprintf("unable to invoke %s", matchedFunction)),
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

func invokefunction(ctx context.Context, c *http.Client, gwURL, callbackURL string, reader io.Reader) (*[]byte, int, *http.Header, error) {

	httpReq, err := http.NewRequest(http.MethodPost, gwURL, reader)
	if err != nil {
		return nil, http.StatusServiceUnavailable, nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	if httpReq.Body != nil {
		defer httpReq.Body.Close()
	}

	if callbackURL != "" {
		httpReq.Header.Add("X-Callback-Url", callbackURL)
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
