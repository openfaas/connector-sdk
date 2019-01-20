// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInvoker_Invoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		function := strings.Split(r.URL.Path, "/")[2]

		switch function {
		case "200":
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(200)
			w.Write([]byte("Hello World"))
		case "401":
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(401)
			w.Write([]byte(""))
		case "500":
		}
		w.WriteHeader(500)
	}))
	client := srv.Client()
	topicMap := NewTopicMap()
	sampleFunc := map[string][]string{
		"Success":        []string{"200"},
		"UnAuthorized":   []string{"401"},
		"Internal Error": []string{"500"},
	}
	topicMap.Sync(&sampleFunc)

	t.Run("Empty Body", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("")
		results := target.Invoke(&topicMap, "Success", &body)

		if len(results) != 0 {
			t.Errorf("Expected 0 results recieved %d", len(results))
		}
	})

	t.Run("Successful Response", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("Body")
		results := target.Invoke(&topicMap, "Success", &body)

		if len(results) != 1 {
			t.Errorf("Expected 1 results recieved %d", len(results))
		}

		result := results["200"]

		if result.Error != nil {
			t.Errorf("Recieved unexpected error %s", result.Error)
		}

		if result.StatusCode != 200 {
			t.Errorf("Expected Statuscode 200 recieved %d", result.StatusCode)
		}

		expected := []byte("Hello World")
		if !bytes.Equal(expected, *result.Body) {
			t.Errorf("Expected %s as body recieved %s", expected, *result.Body)
		}
	})

	t.Run("Unauthorized Response", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("Body")
		results := target.Invoke(&topicMap, "UnAuthorized", &body)

		if len(results) != 1 {
			t.Errorf("Expected 1 results recieved %d", len(results))
		}

		result := results["401"]

		if result.Error != nil {
			t.Errorf("Recieved unexpected error %s", result.Error)
		}

		if result.StatusCode != 401 {
			t.Errorf("Expected Statuscode 200 recieved %d", result.StatusCode)
		}

		expected := []byte("")
		if !bytes.Equal(expected, *result.Body) {
			t.Errorf("Expected %s as body recieved %s", expected, *result.Body)
		}
	})

	t.Run("Server Error Response", func(t *testing.T) {
		target := &Invoker{
			true,
			client,
			srv.URL,
		}

		body := []byte("Body")
		results := target.Invoke(&topicMap, "Internal Error", &body)

		if len(results) != 1 {
			t.Errorf("Expected 1 results recieved %d", len(results))
		}

		result := results["500"]

		if result.Error != nil {
			t.Errorf("Recieved unexpected error %s", result.Error)
		}

		if result.StatusCode != 500 {
			t.Errorf("Expected Statuscode 200 recieved %d", result.StatusCode)
		}

		expected := []byte("")
		if !bytes.Equal(expected, *result.Body) {
			t.Errorf("Expected %s as body recieved %s", expected, *result.Body)
		}
	})
}
