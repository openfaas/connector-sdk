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
	"time"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

// Sarama currently cannot support latest kafka protocol version 0_11_
var (
	SARAMA_KAFKA_PROTO_VER = sarama.V0_10_2_0
)

func main() {
	var client sarama.Client
	var err error

	broker := "kafka"
	if val, exists := os.LookupEnv("broker_host"); exists {
		broker = val
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

	makeConsumer(client, brokers)
}

func makeConsumer(client sarama.Client, brokers []string) {
	//setup consumer
	cConfig := cluster.NewConfig()
	cConfig.Version = SARAMA_KAFKA_PROTO_VER
	cConfig.Consumer.Return.Errors = true
	cConfig.Consumer.Offsets.Initial = sarama.OffsetNewest //OffsetOldest
	cConfig.Group.Return.Notifications = true
	cConfig.Group.Session.Timeout = 6 * time.Second
	cConfig.Group.Heartbeat.Interval = 2 * time.Second

	group := "faas-kafka-queue-workers"
	topics := []string{"faas-request"}

	consumer, err := cluster.NewConsumer(brokers, group, topics, cConfig)
	if err != nil {
		log.Fatalln("Fail to create Kafka consumer: ", err)
	}
	defer consumer.Close()

	c := makeClient(time.Second * 8)

	mcb := func(msg *sarama.ConsumerMessage) {
		if len(msg.Value) > 0 {
			req := InvocationRequest{}
			err := json.Unmarshal(msg.Value, &req)
			if err != nil {
				log.Println("Invalid JSON on topic:", err)
				return
			}

			log.Printf("faas-request to function: %s data: %s", string(req.Name), req.Data)

			reader := bytes.NewReader([]byte(req.Data))
			httpReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://gateway:8080/function/%s", req.Name), reader)
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

				log.Printf("Response [%d] from %s %s", res.StatusCode, req.Name, stringOutput)
			}

		} else {
			log.Printf("Empty message received.")
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

type InvocationRequest struct {
	Name string `json:"name"`
	Data string `json:"data"`
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
