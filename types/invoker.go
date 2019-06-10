package types

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

type Invoker struct {
	PrintResponse bool
	Client        *http.Client
	GatewayURL    string
	Responses     chan InvokerResponse
}

type InvokerResponse struct {
	Body     *[]byte
	Header   *http.Header
	Status   int
	Error    error
	Topic    string
	Function string
}

func NewInvoker(gatewayURL string, client *http.Client, printResponse bool) *Invoker {
	return &Invoker{
		PrintResponse: printResponse,
		Client:        client,
		GatewayURL:    gatewayURL,
		Responses:     make(chan InvokerResponse),
	}
}

// Invoke triggers a function by accessing the API Gateway
func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message *[]byte) {
	if len(*message) == 0 {
		i.Responses <- InvokerResponse{
			Error: fmt.Errorf("no message to send"),
		}
	}

	matchedFunctions := topicMap.Match(topic)
	for _, matchedFunction := range matchedFunctions {
		log.Printf("Invoke function: %s", matchedFunction)

		gwURL := fmt.Sprintf("%s/%s", i.GatewayURL, matchedFunction)
		reader := bytes.NewReader(*message)

		body, statusCode, header, doErr := invokefunction(i.Client, gwURL, reader)

		if doErr != nil {
			i.Responses <- InvokerResponse{
				Error: errors.Wrap(doErr, fmt.Sprintf("unable to invoke %s", matchedFunction)),
			}
			continue
		}

		i.Responses <- InvokerResponse{
			Body:     body,
			Status:   statusCode,
			Header:   header,
			Function: matchedFunction,
			Topic:    topic,
		}
	}
}

func invokefunction(c *http.Client, gwURL string, reader io.Reader) (*[]byte, int, *http.Header, error) {

	httpReq, _ := http.NewRequest(http.MethodPost, gwURL, reader)

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
