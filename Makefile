TAG?=latest
.PHONY: build

build:
	docker build -t alexellis2/kafka-connector:$(TAG) .
push:
	docker push alexellis2/kafka-connector:$(TAG)
