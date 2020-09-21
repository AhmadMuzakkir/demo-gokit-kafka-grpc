.PHONY: protoc-docker protoc run-docker clean-docker

# Build the docker image, then run the docker compose detached
run-docker:
	go mod vendor
	docker-compose down
	docker-compose up --build -d

# Cleanup the docker compose
clean-docker:
	docker-compose down --volumes --rmi all

protoc:
	protoc --gogofaster_out=plugins=grpc:./proto -I=./proto proto/messages.proto

protoc-docker:
	docker run --rm -v `pwd`:/src znly/protoc:0.4.0 --gogofaster_out=plugins=grpc:. -I=. src/proto/messages.proto
