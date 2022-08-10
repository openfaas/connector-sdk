## tester example

Test and emit to the "payment.received" topic

```sh
go build
export PASSWORD="Your gateway password"
./tester \
    -username=admin \
    -password=$PASSWORD
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
