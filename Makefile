TAG?=latest
.PHONY: build

build:
	docker build -t functions/kafka-connector:$(TAG) .
push:
	docker push functions/kafka-connector:$(TAG)
