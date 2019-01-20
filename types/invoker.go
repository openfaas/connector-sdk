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

func (i *Invoker) Invoke(topicMap *TopicMap, topic string, message *[]byte) (result map[string]*InvocationResult) {
	result = make(map[string]*InvocationResult)

	if message != nil && len(*message) > 0 {

		matchedFunctions := topicMap.Match(topic)
		for _, matchedFunction := range matchedFunctions {

			log.Printf("Invoke function: %s", matchedFunction)

			gwURL := fmt.Sprintf("%s/function/%s", i.GatewayURL, matchedFunction)
			reader := bytes.NewReader(*message)

			err, statusCode, headers, body := performInvocation(i.Client, gwURL, reader)

			if err != nil {
				result[matchedFunction] = &InvocationResult{
					StatusCode: statusCode,
					Body:       nil,
					Error:      err,
				}
				return
			}

			if body != nil && i.PrintResponse {
				stringOutput := string(*body)
				log.Printf("Headers: %s", headers)
				log.Printf("Response: [%d] from %s %s", statusCode, matchedFunction, stringOutput)
			}

			result[matchedFunction] = &InvocationResult{
				StatusCode: statusCode,
				Body:       body,
				Error:      nil,
			}
		}
	}
	return result
}

func performInvocation(c *http.Client, gwURL string, reader io.Reader) (err error, statusCode int, responseHeaders http.Header, responseBody *[]byte) {

	httpReq, _ := http.NewRequest(http.MethodPost, gwURL, reader)

	if httpReq.Body != nil {
		defer httpReq.Body.Close()
	}

	var body *[]byte

	res, doErr := c.Do(httpReq)
	if doErr != nil {
		return doErr, http.StatusServiceUnavailable, nil, nil
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			log.Printf("Error reading body")
			return doErr, http.StatusServiceUnavailable, nil, nil

		}
		body = &bytesOut
	}

	return nil, res.StatusCode, res.Header, body
}
