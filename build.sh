#!/bin/sh

docker run -ti --network=kafka_streaming -v `pwd`:/root alpine:3.6 sh
