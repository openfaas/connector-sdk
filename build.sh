#!/bin/sh

docker build -t functions/kafka-connector:latest .
(cd setup && docker service rm kafka_connector ; docker stack deploy kafka -c connector-swarm.yml)
