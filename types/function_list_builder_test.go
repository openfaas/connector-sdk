// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas-provider/types"
)

func TestBuildSingleMatchingFunction(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			annotationMap := make(map[string]string)
			annotationMap["topic"] = "topic1"

			functions = append(functions, types.FunctionStatus{
				Name:        "echo",
				Annotations: &annotationMap,
				Namespace:   "openfaas-fn",
			})
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
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

		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			annotationMap := make(map[string]string)
			annotationMap["topic"] = "topic1"

			functions = append(functions, types.FunctionStatus{
				Name:        "echo",
				Annotations: &annotationMap,
			})
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
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
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			annotationMap := make(map[string]string)
			annotationMap["topic"] = "topic1,topic2,topic3"

			functions = append(functions, types.FunctionStatus{
				Name:        "echo",
				Annotations: &annotationMap,
			})
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
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
		functions := []types.FunctionStatus{}
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
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {

			functions := []types.FunctionStatus{}
			annotationMap := make(map[string]string)
			annotationMap["topic"] = ","

			functions = append(functions, types.FunctionStatus{
				Name:        "echo",
				Annotations: &annotationMap,
			})
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
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
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			annotationMap := make(map[string]string)
			annotationMap["topic"] = "topic1|topic2|topic3,withcomma"

			functions = append(functions, types.FunctionStatus{
				Name:        "echo",
				Annotations: &annotationMap,
			})
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
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
		Namespace          string
		InputServiceMap    map[string][]string
		ExpectedServiceMap map[string][]string
	}{
		{
			Name:            "Empty starting map - key with length",
			Key:             "newKey",
			Function:        "fnName",
			Namespace:       "openfaas-fn",
			InputServiceMap: map[string][]string{},
			ExpectedServiceMap: map[string][]string{
				"newKey": {"fnName.openfaas-fn"},
			},
		},
		{
			Name:               "Empty starting map - zero key length",
			Key:                "",
			Function:           "fnName",
			Namespace:          "openfaas-fn",
			InputServiceMap:    map[string][]string{},
			ExpectedServiceMap: map[string][]string{},
		},
		{
			Name:            "Populated starting map - key with length",
			Key:             "theKey",
			Function:        "newName",
			Namespace:       "fn",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName", "newName.fn"},
			},
		},
		{
			Name:            "Populated starting map - zero key length",
			Key:             "",
			Function:        "newName",
			Namespace:       "fn",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName"},
			},
		},
		{
			Name:            "Populated starting map - new key with length",
			Key:             "newKey",
			Function:        "newName",
			Namespace:       "fn",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName"},
				"newKey": {"newName.fn"},
			},
		},
		{
			Name:            "Populated starting map - existing key new function",
			Key:             "newKey",
			Function:        "secondName",
			Namespace:       "openfaas-fn",
			InputServiceMap: map[string][]string{"theKey": {"fnName"}, "newKey": {"newName"}},
			ExpectedServiceMap: map[string][]string{
				"theKey": {"fnName"},
				"newKey": {"newName", "secondName.openfaas-fn"},
			},
		},
	}

	for _, test := range TestCases {

		serviceMap := appendServiceMap(test.Key, test.Function, test.Namespace, test.InputServiceMap)

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

func Test_BuildMultipleNamespaceFunction(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn", "fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			annotationMap := make(map[string]string)
			annotationMap["topic"] = "topic1"

			functions = append(functions, types.FunctionStatus{
				Name:        "echo",
				Annotations: &annotationMap,
				Namespace:   "openfaas-fn",
			})
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
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
	functions, ok := lookup["topic1"]
	if !ok {
		t.Errorf("Topic %s does not exists", "topic1")
	}
	if len(functions) != 2 {
		t.Errorf("Topic %s - want: %d functions, got: %d", "topic1", 2, len(functions))
	}
}

func Test_GetNamespaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		namespaces := []string{"openfaas-fn", "fn"}
		bytesOut, _ := json.Marshal(namespaces)
		w.Write(bytesOut)
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:     client,
		GatewayURL: srv.URL,
	}

	namespaces, err := builder.getNamespaces()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if len(namespaces) != 2 {
		t.Errorf("Namespaces - want: %d, got: %d", 2, len(namespaces))
	}
}

func Test_GetNamespaces_ProviderGives404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not available"))
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:     client,
		GatewayURL: srv.URL,
	}

	namespaces, err := builder.getNamespaces()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	want := 0
	got := len(namespaces)
	if len(namespaces) != 0 {
		t.Errorf("Namespaces when 404, want %d, but got: %d", want, got)
	}
}

func Test_GetFunctions(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			if r.URL.Query().Get("namespace") == "openfaas-fn" {
				annotationMap := make(map[string]string)
				annotationMap["topic"] = "topic1"

				functions = append(functions, types.FunctionStatus{
					Name:        "echo",
					Annotations: &annotationMap,
					Namespace:   "openfaas-fn",
				})
			}
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: ",",
	}

	functions, err := builder.getFunctions("openfaas-fn")
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(functions) != 1 {
		t.Errorf("Functions - want: %d items, got: %d", 1, len(functions))
	}
}

func Test_GetEmptyFunctions(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/system/namespaces" {
			namespaces := []string{"openfaas-fn"}
			bytesOut, _ := json.Marshal(namespaces)
			w.Write(bytesOut)
		} else {
			functions := []types.FunctionStatus{}
			if r.URL.Query().Get("namespace") == "openfaas-fn" {
				annotationMap := make(map[string]string)
				annotationMap["topic"] = "topic1"

				functions = append(functions, types.FunctionStatus{
					Name:        "echo",
					Annotations: &annotationMap,
					Namespace:   "openfaas-fn",
				})

			}
			bytesOut, _ := json.Marshal(functions)
			w.Write(bytesOut)
		}
	}))

	client := srv.Client()
	builder := FunctionLookupBuilder{
		Client:         client,
		GatewayURL:     srv.URL,
		TopicDelimiter: ",",
	}

	functions, err := builder.getFunctions("fn")
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(functions) != 0 {
		t.Errorf("Functions - want: %d items, got: %d", 0, len(functions))
	}
}
