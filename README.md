Deploy

```
cd setup
docker service rm kafka_connector ; docker stack deploy kafka -c setup.yml
```

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
