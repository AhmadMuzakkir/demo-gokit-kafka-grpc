## Go-Kit, Kafka, gRPC

This project shows an example how to use Go-Kit, Kafka, and gRPC to implement a simple messaging server.

The system should provide two services:
1. Store messages coming from the stream and gRPC service.
2. Get messages.

BadgerDB is used as the database. 
I choose BadgerDB, because it's a simple embedded key-value database.

### Improvements 
There are some improvements that can be made.

- Support more than 1 partition. Currently, only 1 partition is supported.
- Batch the database inserts.
- Add more unit tests.

### How to run

Prerequisites:
- Docker Compose
- Docker Engine

You can use docker compose to run the server and all the required services.

Note on the Dockerfile: <br />
The Dockerfile will run `go build` with `-mod vendor`. 
<br /> So, please make sure your run `go mod vendor` first, before building the image.

To simply things, you can use provided commands in the Makefile, as explained below. 

To run the docker-compose.
```bash
$ make run-docker
```
Assuming default configurations, these addresses will be available in the host:
- GRPC: localhost:8001
- Kafka: localhost:9092
- Zookeper: localhost:2181

To clean up, you can do:
```bash
$ make clean-docker 
```

### Using CLI

To interact with the server, you can use the provided CLI, located in
`cmd/cli`.

The following examples assume default configurations, you can change as you need:

Get the first 10 messages using gRPC.
```bash
$ go run cmd/cli/main.go grpc -a localhost:8001 get -limit 10
```

Send a message with `username = alice, message = 'hello bob'` using gRPC.
```bash
$ go run cmd/cli/main.go grpc -a localhost:8001 send -u alice hello bob
```

Send a message with `username = bob, message = 'hello alice'` to Kafka.
```bash
$ go run cmd/cli/main.go kafka -a localhost:9092 send -u bob hello alice
```
