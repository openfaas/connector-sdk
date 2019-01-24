package types

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Invoker struct {
	PrintResponse bool
	Client        *http.Client
	GatewayURL    string
}

type InvocationResult struct {
	StatusCode int
	Body       *[]byte
	Error      error
}

func NewInvocationResult(statusCode int, body *[]byte, error error) *InvocationResult {
	return &InvocationResult{StatusCode: statusCode, Body: body, Error: error}
}

type InvocationResponse struct {
	StatusCode int
	Body       *[]byte
	Headers http.Header
}

func NewInvocationResponse(statusCode int, body *[]byte, headers http.Header) *InvocationResponse {
	return &InvocationResponse{StatusCode: statusCode, Body: body, Headers: headers}
}

func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message *[]byte) (result map[string]*InvocationResult) {
	result = make(map[string]*InvocationResult)

	if message != nil && len(*message) > 0 {

		matchedFunctions := topicMap.Match(topic)
		for _, matchedFunction := range matchedFunctions {

			log.Printf("Invoke function: %s", matchedFunction)
			functionURL := fmt.Sprintf("%s/function/%s", i.GatewayURL, matchedFunction)
			reader := bytes.NewReader(*message)

			response, err := i.performInvocation(functionURL, reader)

			if err != nil {
				if response != nil {
					result[matchedFunction] = NewInvocationResult(response.StatusCode, nil, err)
				} else {
					result[matchedFunction] = NewInvocationResult(-1, nil, err)
				}
			} else {
				result[matchedFunction] = NewInvocationResult(response.StatusCode, response.Body, err)
			}

			if response != nil && response.Body != nil && i.PrintResponse {
				stringOutput := string(*response.Body)
				log.Printf("Headers: %s", response.Headers)
				log.Printf("Response: [%d] from %s %s", response.StatusCode, matchedFunction, stringOutput)
			}
		}
	}
	return result
}

func (i *Invoker) performInvocation(functionURL string, bodyReader io.Reader) (*InvocationResponse, error) {

	httpReq, requestErr := http.NewRequest(http.MethodPost, functionURL, bodyReader)

	if requestErr != nil {
		return nil, requestErr
	}

	if httpReq.Body != nil {
		defer httpReq.Body.Close()
	}

	var body *[]byte

	res, doErr := i.Client.Do(httpReq)
	if doErr != nil {
		return nil, doErr
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Printf("Error reading body")
			return NewInvocationResponse(res.StatusCode, nil, res.Header), readErr

		}
		body = &bytesOut
	}

	return NewInvocationResponse(res.StatusCode, body, res.Header), nil
}
