// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/pkg/errors"
)

// FunctionLookupBuilder builds a list of OpenFaaS functions
type FunctionLookupBuilder struct {
	GatewayURL     string
	Client         *http.Client
	Credentials    *auth.BasicAuthCredentials
	TopicDelimiter string
}

//getNamespaces get openfaas namespaces
func (s *FunctionLookupBuilder) getNamespaces() ([]string, error) {
	var (
		err        error
		namespaces []string
	)
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/system/namespaces", s.GatewayURL), nil)
	if err != nil {
		return namespaces, err
	}

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return namespaces, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusNotFound {
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return namespaces, err
		}

		if len(bytesOut) == 0 {
			return namespaces, nil
		}

		err = json.Unmarshal(bytesOut, &namespaces)
		if err != nil {
			return namespaces, err
		}
	}

	return namespaces, err
}

func (s *FunctionLookupBuilder) getFunctions(namespace string) ([]types.FunctionStatus, error) {
	gateway := fmt.Sprintf("%s/system/functions", s.GatewayURL)
	gatewayURL, err := url.Parse(gateway)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("invalid gateway URL: %s", err.Error())
	}
	if len(namespace) > 0 {
		query := gatewayURL.Query()
		query.Set("namespace", namespace)
		gatewayURL.RawQuery = query.Encode()
	}

	req, _ := http.NewRequest(http.MethodGet, gatewayURL.String(), nil)
	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, reqErr := s.Client.Do(req)

	if reqErr != nil {
		return []types.FunctionStatus{}, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, _ := ioutil.ReadAll(res.Body)

	functions := []types.FunctionStatus{}
	marshalErr := json.Unmarshal(bytesOut, &functions)

	if marshalErr != nil {
		return []types.FunctionStatus{}, errors.Wrap(marshalErr, fmt.Sprintf("unable to unmarshal value: %q", string(bytesOut)))
	}

	return functions, nil
}

// Build compiles a map of topic names and functions that have
// advertised to receive messages on said topic
func (s *FunctionLookupBuilder) Build() (map[string][]string, error) {
	var (
		err error
	)

	namespaces, err := s.getNamespaces()
	if err != nil {
		return map[string][]string{}, err
	}
	serviceMap := make(map[string][]string)

	if len(namespaces) == 0 {
		namespace := ""
		functions, err := s.getFunctions(namespace)
		if err != nil {
			return map[string][]string{}, err
		}
		serviceMap = buildServiceMap(&functions, s.TopicDelimiter, namespace, serviceMap)
	} else {
		for _, namespace := range namespaces {
			functions, err := s.getFunctions(namespace)
			if err != nil {
				return map[string][]string{}, err
			}
			serviceMap = buildServiceMap(&functions, s.TopicDelimiter, namespace, serviceMap)
		}
	}

	return serviceMap, err
}

func buildServiceMap(functions *[]types.FunctionStatus, topicDelimiter, namespace string, serviceMap map[string][]string) map[string][]string {
	for _, function := range *functions {

		if function.Annotations != nil {

			annotations := *function.Annotations

			if topicNames, exist := annotations["topic"]; exist {

				if len(topicDelimiter) > 0 && strings.Count(topicNames, topicDelimiter) > 0 {

					topicSlice := strings.Split(topicNames, topicDelimiter)

					for _, topic := range topicSlice {
						serviceMap = appendServiceMap(topic, function.Name, namespace, serviceMap)
					}
				} else {
					serviceMap = appendServiceMap(topicNames, function.Name, namespace, serviceMap)
				}
			}
		}
	}
	return serviceMap
}

func appendServiceMap(key, function, namespace string, sm map[string][]string) map[string][]string {

	key = strings.TrimSpace(key)

	if len(key) > 0 {

		if sm[key] == nil {
			sm[key] = []string{}
		}
		sep := ""
		if len(namespace) > 0 {
			sep = "."
		}

		functionPath := fmt.Sprintf("%s%s%s", function, sep, namespace)
		sm[key] = append(sm[key], functionPath)
	}

	return sm
}
