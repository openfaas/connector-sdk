// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
				if err != nil{
					fmt.Printf("Recieved %s",err)
				}else{
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
		"All":        []string{"success", "headers", "wrong_payload"},
		"Contains_Fail":   []string{"success", "server_error", "aborts", "headers"},
		"NOP": []string{},
	}

	topicMap.Sync(&sampleFunc)

	t.Run("Should invoke no function when body is nil", func(t *testing.T) {
		target := &Invoker{
			PrintResponse:false,
			Client:client,
			GatewayURL: srv.URL,
		}

		results := target.Invoke(&topicMap, "NOP", nil)

		if len(results) != 0 {
			t.Errorf("When body is nil it should perform a request")
		}
	})

	t.Run("Should invoke no function when body is empty", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("")
		results := target.Invoke(&topicMap, "NOP", &body)

		if len(results) != 0 {
			t.Errorf("When body is empty it should perform a request")
		}
	})

	t.Run("Should invoke all functions", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("Some Input")
		results := target.Invoke(&topicMap, "All", &body)

		const ExpectedResults = 3
		if len(results) != ExpectedResults {
			t.Errorf("Expected %d results recieved %d", ExpectedResults, len(results))
		}

		for name, result := range results {
			if result.Error != nil {
				t.Errorf("Received unexpected error %s for %s", result.Error, name)
			}

			if result.StatusCode != 200 && result.StatusCode != 400 {
				t.Errorf("Received unexpected status code %d for %s", result.StatusCode, name)
			}
		}

	})

	t.Run("Should invoke all functions even if one request fails", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("Hello World")
		results := target.Invoke(&topicMap, "Contains_Fail", &body)

		const ExpectedResults = 4
		if len(results) != ExpectedResults {
			t.Errorf("Expected %d results recieved %d", ExpectedResults, len(results))
		}

		for name, result := range results {
			if name == "aborts" {
				if result.Error == nil {
					t.Errorf("Expected call for %s to fail", name)
				}
			} else {
				if result.Error != nil {
					t.Errorf("Received unexpected error %s for %s", result.Error, name)
				}
			}
		}
	})
}
