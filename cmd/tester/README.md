# OpenFaas Sample Connector

This folder contains a sample OpenFaas connector. You can use this as a base for creating your own connectors.
For a more complex example checkout [kafka-connector](https://github.com/openfaas-incubator/kafka-connector)

## How to Use

1. Clone this repository: `git clone https://github.com/openfaas-incubator/connector-sdk.git`
2. Go into the directory: `cd ./connector-sdk/cmd/tester/yaml`
3. For OpenFaas deployed on Docker Swarm do: `docker stack deploy func -c ./docker-compose.yml`
4. For OpenFaas deployed on kubernetes do: `kubectl create -f ./kubernetes --namespace openfaas`

To check if it actually works and triggers a function, deploy any function with annotation `topic=faas-request`.
You can also run this command to deploy a sample function and see `trigger-func` invocation count growing in ui.

```bash
faas-cli deploy --image=functions/nodeinfo --name=trigger-func --annotation topic=faas-request
```