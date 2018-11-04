// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/openfaas-incubator/kafka-connector/types"
	auth "github.com/openfaas/faas-provider/auth"
)

// Sarama currently cannot support latest kafka protocol version V0_10_2_0
var saramaKafkaProtocolVersion = sarama.V0_10_2_0

type connectorConfig struct {
	gatewayURL      string
	upstreamTimeout time.Duration
	topics          []string
	printResponse   bool
	rebuildInterval time.Duration
	broker          string
}

func main() {
	var client sarama.Client
	var err error

	var credentials *auth.BasicAuthCredentials

	if val, ok := os.LookupEnv("basic_auth"); ok && len(val) > 0 {
		if val == "true" || val == "1" {

			reader := auth.ReadBasicAuthFromDisk{}

			if val, ok := os.LookupEnv("secret_mount_path"); ok && len(val) > 0 {
				reader.SecretMountPath = os.Getenv("secret_mount_path")
			}

			res, err := reader.Read()
			if err != nil {
				panic(err)
			}
			credentials = res
		}
	}

	config := buildConnectorConfig()

	topicMap := types.NewTopicMap()

	lookupBuilder := types.FunctionLookupBuilder{
		GatewayURL:  config.gatewayURL,
		Client:      types.MakeClient(config.upstreamTimeout),
		Credentials: credentials,
	}

	ticker := time.NewTicker(config.rebuildInterval)
	go synchronizeLookups(ticker, &lookupBuilder, &topicMap)

	brokers := []string{config.broker + ":9092"}
	for {

		client, err = sarama.NewClient(brokers, nil)
		if client != nil && err == nil {
			break
		}
		if client != nil {
			client.Close()
		}
		fmt.Println("Wait for brokers ("+config.broker+") to come up.. ", brokers)

		time.Sleep(1 * time.Second)
	}

	makeConsumer(client, brokers, config, &topicMap)
}

func synchronizeLookups(ticker *time.Ticker,
	lookupBuilder *types.FunctionLookupBuilder,
	topicMap *types.TopicMap) {

	for {
		<-ticker.C
		lookups, err := lookupBuilder.Build()
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("Syncing topic map")
		topicMap.Sync(&lookups)
	}
}

func makeConsumer(client sarama.Client, brokers []string, config connectorConfig, topicMap *types.TopicMap) {
	//setup consumer
	cConfig := cluster.NewConfig()
	cConfig.Version = saramaKafkaProtocolVersion
	cConfig.Consumer.Return.Errors = true
	cConfig.Consumer.Offsets.Initial = sarama.OffsetNewest //OffsetOldest
	cConfig.Group.Return.Notifications = true
	cConfig.Group.Session.Timeout = 6 * time.Second
	cConfig.Group.Heartbeat.Interval = 2 * time.Second

	group := "faas-kafka-queue-workers"

	topics := config.topics
	log.Printf("Binding to topics: %v", config.topics)

	consumer, err := cluster.NewConsumer(brokers, group, topics, cConfig)
	if err != nil {
		log.Fatalln("Fail to create Kafka consumer: ", err)
	}

	defer consumer.Close()

	mcb := makeMessageHandler(topicMap, config)

	num := 0

	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				num = (num + 1) % math.MaxInt32
				fmt.Printf("[#%d] Received on [%v,%v]: '%s'\n",
					num,
					msg.Topic,
					msg.Partition,
					string(msg.Value))

				mcb(msg)

				consumer.MarkOffset(msg, "") // mark message as processed
			}
		case err = <-consumer.Errors():

			fmt.Println("consumer error: ", err)

		case ntf := <-consumer.Notifications():

			fmt.Printf("Rebalanced: %+v\n", ntf)

		}
	}
}

func makeMessageHandler(topicMap *types.TopicMap, config connectorConfig) func(msg *sarama.ConsumerMessage) {

	invoker := types.Invoker{
		PrintResponse: config.printResponse,
		Client:        types.MakeClient(config.upstreamTimeout),
		GatewayURL:    config.gatewayURL,
	}

	mcb := func(msg *sarama.ConsumerMessage) {
		invoker.Invoke(topicMap, msg.Topic, &msg.Value)
	}
	return mcb
}

func buildConnectorConfig() connectorConfig {

	broker := "kafka"
	if val, exists := os.LookupEnv("broker_host"); exists {
		broker = val
	}

	topics := []string{}
	if val, exists := os.LookupEnv("topics"); exists {
		for _, topic := range strings.Split(val, ",") {
			if len(topic) > 0 {
				topics = append(topics, topic)
			}
		}
	}
	if len(topics) == 0 {
		log.Fatal(`Provide a list of topics i.e. topics="payment_published,slack_joined"`)
	}

	gatewayURL := "http://gateway:8080"
	if val, exists := os.LookupEnv("gateway_url"); exists {
		gatewayURL = val
	}

	upstreamTimeout := time.Second * 30
	rebuildInterval := time.Second * 3

	if val, exists := os.LookupEnv("upstream_timeout"); exists {
		parsedVal, err := time.ParseDuration(val)
		if err == nil {
			upstreamTimeout = parsedVal
		}
	}

	if val, exists := os.LookupEnv("rebuild_interval"); exists {
		parsedVal, err := time.ParseDuration(val)
		if err == nil {
			rebuildInterval = parsedVal
		}
	}

	printResponse := false
	if val, exists := os.LookupEnv("print_response"); exists {
		printResponse = (val == "1" || val == "true")
	}

	return connectorConfig{
		gatewayURL:      gatewayURL,
		upstreamTimeout: upstreamTimeout,
		topics:          topics,
		rebuildInterval: rebuildInterval,
		broker:          broker,
		printResponse:   printResponse,
	}
}
