// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/sdk"
	"github.com/openfaas/faas-provider/types"
)

// FunctionLookupBuilder builds a list of OpenFaaS functions
type FunctionLookupBuilder struct {
	GatewayURL     string
	Client         *http.Client
	Credentials    *auth.BasicAuthCredentials
	TopicDelimiter string
	sdk            *sdk.SDK
}

func NewFunctionLookupBuilder(gatewayURL, topicDelimiter string, client *http.Client, credentials *auth.BasicAuthCredentials) *FunctionLookupBuilder {
	u, _ := url.Parse(gatewayURL)
	return &FunctionLookupBuilder{
		sdk:            sdk.NewSDK(u, credentials, client),
		TopicDelimiter: topicDelimiter,
	}
}

// Build compiles a map of topic names and functions that have
// advertised to receive messages on said topic
func (s *FunctionLookupBuilder) Build() (map[string][]string, error) {
	var err error

	namespaces, err := s.sdk.GetNamespaces()
	if err != nil {
		return map[string][]string{}, err
	}

	serviceMap := make(map[string][]string)

	if len(namespaces) == 0 {
		namespaces = []string{""}
	}

	for _, namespace := range namespaces {
		functions, err := s.sdk.GetFunctions(namespace)
		if err != nil {
			return map[string][]string{}, fmt.Errorf("unable to get functions in: %s, error: %w", namespace, err)
		}
		serviceMap = buildServiceMap(&functions, s.TopicDelimiter, namespace, serviceMap)
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
