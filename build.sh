#!/bin/sh

docker build -t alexellis2/kafka-connector:latest .
(cd setup && docker service rm kafka_connector ; docker stack deploy kafka -c setup.yml)
