## tester example

Test and emit to the "payment.received" topic

```sh
go build
export PASSWORD="Your gateway password"
./tester \
    -username=admin \
    -password=$PASSWORD
```

Deploy an example function to be triggered by the topic:

```bash
faas-cli store deploy printer --annotation topic=payment.received
```

Emit a custom topic:

```sh
go build
export PASSWORD="Your gateway password"

./tester \
    -username=admin \
    -password=$PASSWORD \
    -topic "custom/topic/1"
```

Deploy an example function to be triggered by the topic:

```bash
faas-cli store deploy printer --annotation topic=custom/topic/1
```
