// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas/gateway/requests"
)

func TestBuildSingleMatchingFunction(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		functions := []requests.Function{}
		annotationMap := make(map[string]string)
		annotationMap["topic"] = "topic1"

		functions = append(functions, requests.Function{
			Name:        "echo",
			Annotations: &annotationMap,
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: ",",
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 1 {
		t.Errorf("Lookup - want: %d items, got: %d", 1, len(lookup))
	}
}
func Test_Build_SingleFunctionNoDelimiter(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		functions := []requests.Function{}
		annotationMap := make(map[string]string)
		annotationMap["topic"] = "topic1"

		functions = append(functions, requests.Function{
			Name:        "echo",
			Annotations: &annotationMap,
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:     client,
		GatewayURL: srv.URL,
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 1 {
		t.Errorf("Lookup - want: %d items, got: %d", 1, len(lookup))
	}
}

func TestBuildMultiMatchingFunction(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		functions := []requests.Function{}
		annotationMap := make(map[string]string)
		annotationMap["topic"] = "topic1,topic2,topic3"

		functions = append(functions, requests.Function{
			Name:        "echo",
			Annotations: &annotationMap,
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: ",",
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 3 {
		t.Errorf("Lookup - want: %d items, got: %d", 3, len(lookup))
	}
}

func TestBuildNoFunctions(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		functions := []requests.Function{}
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: ",",
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 0 {
		t.Errorf("Lookup - want: %d items, got: %d", 0, len(lookup))
	}
}

func Test_Build_JustDelim(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		functions := []requests.Function{}
		annotationMap := make(map[string]string)
		annotationMap["topic"] = ","

		functions = append(functions, requests.Function{
			Name:        "echo",
			Annotations: &annotationMap,
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: ",",
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 0 {
		t.Errorf("Lookup - want: %d items, got: %d", 0, len(lookup))
	}
}

func Test_Build_MultiMatchingFunctionBespokeDelim(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		functions := []requests.Function{}
		annotationMap := make(map[string]string)
		annotationMap["topic"] = "topic1|topic2|topic3,withcomma"

		functions = append(functions, requests.Function{
			Name:        "echo",
			Annotations: &annotationMap,
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: "|",
	}

	lookup, err := builder.Build()
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(lookup) != 3 {
		t.Errorf("Lookup - want: %d items, got: %d", 3, len(lookup))
	}
}

func Test_appendServiceMap(t *testing.T) {
	var TestCases = []struct {
		Name               string
		Key                string
		Function           string
		InputServiceMap    map[string][]string
		ExpectedServiceMap map[string][]string
	}{
		{
			Name:            "Empty starting map - key with length",
			Key:             "newKey",
			Function:        "fnName",
			InputServiceMap: map[string][]string{},
			ExpectedServiceMap: map[string][]string{
				"newKey": {"fnName"},
			},
		},
		{
			Name:               "Empty starting map - zero key length",
			Key:                "",
			Function:           "fnName",
			InputServiceMap:    map[string][]string{},
			ExpectedServiceMap: map[string][]string{},
		},
		{
			Name:            "Populated starting map - key with length",
			Key:             "theKey",
			Function:        "newName",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName", "newName"},
			},
		},
		{
			Name:            "Populated starting map - zero key length",
			Key:             "",
			Function:        "newName",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName"},
			},
		},
		{
			Name:            "Populated starting map - new key with length",
			Key:             "newKey",
			Function:        "newName",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName"},
				"newKey": {"newName"},
			},
		},
		{
			Name:            "Populated starting map - existing key new function",
			Key:             "newKey",
			Function:        "secondName",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}, "newKey": {"newName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName"},
				"newKey": {"newName", "secondName"},
			},
		},
	}

	for _, test := range TestCases {

		serviceMap := appendServiceMap(test.Key, test.Function, test.InputServiceMap)

		if len(serviceMap) != len(test.ExpectedServiceMap) {
			t.Errorf("Testcase %s failed on serviceMap size. want - %d, got - %d", test.Name, len(test.ExpectedServiceMap), len(serviceMap))
		}

		for key := range serviceMap {

			if _, exists := test.ExpectedServiceMap[key]; !exists {
				t.Errorf("Testcase %s failed on serviceMap keys. found value - %s doesnt exist in expected", test.Name, key)
			}

			if len(serviceMap[key]) != len(test.ExpectedServiceMap[key]) {
				t.Errorf("Testcase %s failed on key slice size. want - %d, got - %d", test.Name, len(test.ExpectedServiceMap[key]), len(serviceMap[key]))
			}

			lookupMap := make(map[string]bool)
			for _, fn := range serviceMap[key] {
				lookupMap[fn] = true
			}

			for _, v := range test.ExpectedServiceMap[key] {
				if _, found := lookupMap[v]; !found {
					t.Errorf("Testcase %s failed on key slice values. found value - %s doesnt exist in expected", test.Name, v)
				}
			}
		}
	}
}
