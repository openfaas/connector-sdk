package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/openfaas/faas/gateway/requests"
)

// Sarama currently cannot support latest kafka protocol version 0_11_
var (
	SARAMA_KAFKA_PROTO_VER = sarama.V0_10_2_0
)

type connectorConfig struct {
	gatewayURL      string
	upstreamTimeout time.Duration
	topics          []string
}

type TopicMap struct {
	lookup *map[string][]string
	lock   sync.Mutex
}

func (t *TopicMap) Match(topicName string) []string {
	t.lock.Lock()

	var values []string

	for key, val := range *t.lookup {
		if key == topicName {
			values = val
			break
		}
	}

	t.lock.Unlock()

	return values
}

func (t *TopicMap) Sync(updated *map[string][]string) {
	t.lock.Lock()

	t.lookup = updated

	t.lock.Unlock()
}

func main() {
	var client sarama.Client
	var err error

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

	config := connectorConfig{
		gatewayURL:      gatewayURL,
		upstreamTimeout: upstreamTimeout,
		topics:          topics,
	}

	brokers := []string{broker + ":9092"}
	for {

		client, err = sarama.NewClient(brokers, nil)
		if client != nil && err == nil {
			break
		}
		if client != nil {
			client.Close()
		}
		fmt.Println("Wait for brokers ("+broker+") to come up.. ", brokers)

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Topics")
	fmt.Println(client.Topics())

	makeConsumer(client, brokers, config)
}

type ServiceListBroker struct {
	gatewayURL string
	client     *http.Client
}

func (s *ServiceListBroker) Build() (map[string][]string, error) {
	var err error
	serviceMap := make(map[string][]string)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/system/functions", s.gatewayURL), nil)
	res, reqErr := s.client.Do(req)

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
		if *function.Labels != nil {
			labels := *function.Labels

			if topic, pass := labels["topic"]; pass {

				if serviceMap[topic] == nil {
					serviceMap[topic] = []string{}
				}
				serviceMap[topic] = append(serviceMap[topic], function.Name)
			}
		}
	}

	return serviceMap, err
}

func makeConsumer(client sarama.Client, brokers []string, config connectorConfig) {
	//setup consumer
	cConfig := cluster.NewConfig()
	cConfig.Version = SARAMA_KAFKA_PROTO_VER
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

	c := makeClient(time.Second * 8)

	lookup := make(map[string][]string)
	topicMap := TopicMap{
		lookup: &lookup,
	}

	listBroker := ServiceListBroker{
		gatewayURL: config.gatewayURL,
		client:     makeClient(time.Second * 3),
	}

	ticker := time.NewTicker(time.Second * 3)

	go func() {
		for {
			<-ticker.C
			lookups, err := listBroker.Build()
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("Syncing topic map")
			topicMap.Sync(&lookups)
		}
	}()

	mcb := func(msg *sarama.ConsumerMessage) {
		if len(msg.Value) > 0 {

			matchedFunctions := topicMap.Match(msg.Topic)
			for _, matchedFunction := range matchedFunctions {

				log.Printf("Invoke function: %s", matchedFunction)

				reader := bytes.NewReader([]byte(msg.Value))
				httpReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/function/%s", config.gatewayURL, matchedFunction), reader)
				defer httpReq.Body.Close()

				res, doErr := c.Do(httpReq)
				if doErr != nil {
					log.Println("Invalid response:", doErr)
					return
				}
				if res.Body != nil {
					defer res.Body.Close()

					bytesOut, readErr := ioutil.ReadAll(res.Body)
					if readErr != nil {
						log.Printf("Error reading body")
					}
					stringOutput := string(bytesOut)

					log.Printf("Response [%d] from %s %s", res.StatusCode, matchedFunction, stringOutput)
				}

			}
		}
	}

	num := 0

	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				num = (num + 1) % math.MaxInt32
				fmt.Printf("[#%d] Received on [%v,%v]: '%s'\n", num, msg.Topic, msg.Partition, string(msg.Value))

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

func makeClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			// DisableKeepAlives:     false,
			IdleConnTimeout: 120 * time.Millisecond,
			// ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
}
