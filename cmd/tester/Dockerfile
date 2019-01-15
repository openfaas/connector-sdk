FROM golang:1.9.2 as builder
RUN mkdir -p /go/src/github.com/openfaas-incubator/connector
WORKDIR /go/src/github.com/openfaas-incubator/connector

COPY main.go    .

RUN go get -d -v ./...

# Stripping via -ldflags "-s -w" 
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-s -w" -installsuffix cgo -o ./connector

FROM alpine

COPY --from=builder /go/src/github.com/openfaas-incubator/connector/connector /bin/connector

CMD ["/bin/connector"]