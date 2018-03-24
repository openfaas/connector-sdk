This producer can be used to generate load for the Kafka connector.

Example:

* Ingest 15k messages:

```shell
 $ kubectl run -t -i --image alexellis2/kafka-producer producer bash
./producer -messages=15000
```

