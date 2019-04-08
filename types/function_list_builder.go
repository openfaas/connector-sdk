// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas/gateway/requests"
)

// FunctionLookupBuilder builds a list of OpenFaaS functions
type FunctionLookupBuilder struct {
	GatewayURL     string
	Client         *http.Client
	Credentials    *auth.BasicAuthCredentials
	TopicDelimiter string
}

// Build compiles a map of topic names and functions that have
// advertised to receive messages on said topic
func (s *FunctionLookupBuilder) Build() (map[string][]string, error) {
	var err error
	serviceMap := make(map[string][]string)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/system/functions", s.GatewayURL), nil)

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, reqErr := s.Client.Do(req)

	if reqErr != nil {
		return serviceMap, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, _ := ioutil.ReadAll(res.Body)

	functions := []requests.Function{}
	marshalErr := json.Unmarshal(bytesOut, &functions)

	if marshalErr != nil {
		return serviceMap, marshalErr
	}

	for _, function := range functions {

		if function.Annotations != nil {

			annotations := *function.Annotations

			if topicNames, exist := annotations["topic"]; exist {

				if len(s.TopicDelimiter) > 0 && strings.Count(topicNames, s.TopicDelimiter) > 0 {

					topicSlice := strings.Split(topicNames, s.TopicDelimiter)

					for _, topic := range topicSlice {
						serviceMap = appendServiceMap(topic, function.Name, serviceMap)
					}
				} else {
					serviceMap = appendServiceMap(topicNames, function.Name, serviceMap)
				}
			}
		}
	}
	return serviceMap, err
}

func appendServiceMap(key string, function string, sm map[string][]string) map[string][]string {

	key = strings.TrimSpace(key)

	if len(key) > 0 {

		if sm[key] == nil {
			sm[key] = []string{}
		}
		sm[key] = append(sm[key], function)
	}

	return sm
}
