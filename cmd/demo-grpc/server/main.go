package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

// This is a skeleton placeholder. Add generated protobuf code under api/ and
// register a real service when turning the demo into a runnable gRPC app.
func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	log.Println("demo gRPC server listening on :9090")
	log.Fatal(server.Serve(ln))
}
