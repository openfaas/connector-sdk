## Kafka-connector

This is a Proof-of-Concept for connecting a Kakfa topic to the OpenFaaS API Gateway. 

This sample is setup for use on Swarm, but the application code will also work on Kubernetes.

Usage:

* Deploy this sample
* Publish messages in JSON format to the `faas-request` topic:

```
{"name": "func_echoit", "data": "echo this message"}
```

This causes the `"func_echoit"` function to be triggered synchronously with the data `"echo this message"`.

* The HTTP response from the function is written to the logs of the `connector` service.

Todo:
- [] Write function reponse back to a separate topic or Callback-URL

### Development

Build the connector image:

```
./build.sh
```

Deploy the Kafka stack and the connector

```
(cd setup && docker service rm kafka_connector ; docker stack deploy kafka -c setup.yml)
```

> Note: If the broker has a different name from `kafka` you can pass the `broker_host` environmental variable. This exclude the port.

Trigger a function via the topic "faas-request"

```
SERVICE_NAME=kafka_kafka
TASK_ID=$(docker service ps --filter 'desired-state=running' $SERVICE_NAME -q)
CONTAINER_ID=$(docker inspect --format '{{ .Status.ContainerStatus.ContainerID }}' $TASK_ID)
docker exec -it $CONTAINER_ID kafka-console-producer --broker-list kafka:9092 --topic faas-request

{"name": "func_echoit", "data": "test"}
```

Grab logs:

```
docker service logs kafka_connector -f
kafka_connector.1.93e5xd7bk47h@moby    | Topics
kafka_connector.1.93e5xd7bk47h@moby    | [echo faas-request __consumer_offsets] <nil>
kafka_connector.1.93e5xd7bk47h@moby    | Rebalanced: &{Type:rebalance start Claimed:map[] Released:map[] Current:map[]}
kafka_connector.1.93e5xd7bk47h@moby    | Rebalanced: &{Type:rebalance OK Claimed:map[faas-request:[0]] Released:map[] Current:map[faas-request:[0]]}
kafka_connector.1.93e5xd7bk47h@moby    | 2017/11/14 18:30:48 faas-request to function: func_echoit
kafka_connector.1.93e5xd7bk47h@moby    | [#5] Received on [faas-request,0]: '{"name": "func_echoit", "data": "test"}'
```
