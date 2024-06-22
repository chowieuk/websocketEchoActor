# Websocket Echo Actor

An adaptation of https://github.com/gorilla/websocket/tree/main/examples/echo that uses Proto.Actor for responding to messages

To build and run the example, you must have the protoc compiler installed, and the Protobuffer package dependencies for go and JS installed.

## Protoc compiler

Follow the guide at https://grpc.io/docs/protoc-installation/

## Protobuf package dependencies

### Go

```sh
go get -u google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

NOTE: You should also ensure protoc-gen-go is included in your PATH

```sh
PATH="${PATH}:${HOME}/go/bin"
```

### JS

```sh
npm install -g protoc-gen-js
npm install
```

## Build and run

```sh
make build
make run
```

The server includes a simple web client. To use the client, open http://127.0.0.1:8080 in the browser and follow the instructions on the page.
