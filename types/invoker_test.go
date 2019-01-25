package types

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestInvoker_Invoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		function := strings.Split(r.URL.Path, "/")[2]

		switch function {
		case "success":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			w.Write([]byte("Hello World"))
			break
		case "headers":
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Charset", "utf-8")
			w.WriteHeader(200)
			w.Write([]byte("<h1>Hello World</h1>"))
			break
		case "wrong_payload":
			w.WriteHeader(400)
			w.Write([]byte(""))
			break
		case "aborts":
			if wr, ok := w.(http.Hijacker); ok {
				conn, _, err := wr.Hijack()
				if err != nil {
					fmt.Printf("Recieved %s", err)
				} else {
					conn.Close()
				}

			}
			break
		case "server_error":
			w.WriteHeader(500)
			w.Write([]byte(""))
			break
		}
	}))

	client := srv.Client()
	topicMap := NewTopicMap()

	sampleFunc := map[string][]string{
		"All":           []string{"success", "headers", "wrong_payload"},
		"Contains_Fail": []string{"success", "server_error", "aborts", "headers"},
		"NOP":           []string{},
	}

	topicMap.Sync(&sampleFunc)

	t.Run("Should invoke no function when body is empty", func(t *testing.T) {
		responseChannel := make(chan InvokerResponse, 1)
		defer close(responseChannel)

		var wg sync.WaitGroup

		target := &Invoker{
			PrintResponse: false,
			Client:        client,
			GatewayURL:    srv.URL,
			Responses:     responseChannel,
		}
		body := []byte("")
		target.Invoke(&topicMap, "NOP", &body)

		const ExpectedAmountOfResponses = 1
		if len(responseChannel) != ExpectedAmountOfResponses {
			t.Errorf("Response Channel does not contain the expected amount of %d responses instead it has %d responses", ExpectedAmountOfResponses, len(responseChannel))
		}

		wg.Add(1)

		go func() {
			defer wg.Done()
			response := <-responseChannel
			const ExpectedErrorMessage = "no message to send"
			if response.Error.Error() != ExpectedErrorMessage {
				t.Errorf("Expected %s received %s", ExpectedErrorMessage, response.Error)
			}
		}()

		wg.Wait()
	})

	t.Run("Should invoke all functions", func(t *testing.T) {
		responseChannel := make(chan InvokerResponse, 3)
		defer close(responseChannel)

		var wg sync.WaitGroup

		target := &Invoker{
			PrintResponse: false,
			Client:        client,
			GatewayURL:    srv.URL,
			Responses:     responseChannel,
		}

		body := []byte("Some Input")
		target.Invoke(&topicMap, "All", &body)

		const ExpectedAmountOfResponses = 3
		if len(responseChannel) != ExpectedAmountOfResponses {
			t.Errorf("Response Channel does not contain the expected amount of %d responses instead it has %d responses", ExpectedAmountOfResponses, len(responseChannel))
		}

		wg.Add(3)

		for i := 0; i < 3; i++ {
			go func() {
				defer wg.Done()
				response := <-responseChannel
				if response.Status != 200 && response.Status != 400 {
					t.Errorf("Received unexpected status code %d for %s", response.Status, response.Function)
				}
				if response.Error != nil {
					t.Errorf("Received unexpected error %s for %s", response.Error, response.Function)
				}
			}()
		}

		wg.Wait()

	})

	t.Run("Should invoke all functions even if one request fails", func(t *testing.T) {
		responseChannel := make(chan InvokerResponse, 4)
		defer close(responseChannel)

		var wg sync.WaitGroup

		target := &Invoker{
			PrintResponse: false,
			Client:        client,
			GatewayURL:    srv.URL,
			Responses:     responseChannel,
		}

		body := []byte("Hello World")
		target.Invoke(&topicMap, "Contains_Fail", &body)

		const ExpectedAmountOfResponses = 4
		if len(responseChannel) != ExpectedAmountOfResponses {
			t.Errorf("Response Channel does not contain the expected amount of %d responses instead it has %d responses", ExpectedAmountOfResponses, len(responseChannel))
		}

		wg.Add(4)

		for i := 0; i < 4; i++ {
			go func() {
				defer wg.Done()
				response := <-responseChannel

				if response.Status == 0 {
					if response.Error == nil {
						t.Errorf("Expected call for %s to fail", response.Function)
					}
				} else {
					if response.Error != nil {
						t.Errorf("Received unexpected error %s for %s", response.Error, response.Function)
					}
				}
			}()
		}

		wg.Wait()
	})
}
