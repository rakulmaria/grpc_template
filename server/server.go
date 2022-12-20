package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	proto "whatTime/proto"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedChittyChatServer // Necessary
	port                                string
	LamportClock                        int64
	streams                             map[string]*proto.ChittyChat_JoinServer
}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var port = flag.String("port", "8080", "server port number")

func main() {
	f := setLog() // enables the logger
	defer f.Close()
	// Get the port from the command line when the server is run
	flag.Parse()

	// Start the server
	go startServer()

	// Keep the server running until it is manually quit
	for {

	}
}

func startServer() {
	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		fmt.Printf("Failed to listen on port %s: %v", *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &Server{
		port:         *port,
		LamportClock: 0,
		streams:      make(map[string]*proto.ChittyChat_JoinServer),
	}

	proto.RegisterChittyChatServer(grpcServer, server) //Registers the server to the gRPC server.

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
}

func (s *Server) Join(request *proto.JoinRequest, stream proto.ChittyChat_JoinServer) error {
	log.Printf("Server: Join request from %s\n", request.User)

	// adds the stream to the streams map
	s.streams[request.User] = &stream

	// sends a message to the client
	sendToAll(s.streams, &proto.Message{
		User:         "Server",
		Message:      "Welcome " + request.User + " to the Chitty Chat!",
		LamportClock: s.LamportClock,
	})

	// waits for the stream to be closed -- happens when the client stops
	// then removes the stream from the streams map
	// and sends a message to the other clients
	<-stream.Context().Done()
	delete(s.streams, request.User)
	log.Println(request.User, "disconnected")

	sendToAll(s.streams, &proto.Message{
		User:         "Server",
		Message:      request.User + " has left the chat",
		LamportClock: s.LamportClock,
	})
	return nil
}

func (s *Server) Publish(ctx context.Context, message *proto.Message) (*proto.Empty, error) {
	if message.LamportClock < s.LamportClock {
		message.LamportClock = s.LamportClock + 1
	}

	// update the Lamport time of the server
	s.LamportClock = message.LamportClock

	sendToAll(s.streams, message)

	return &proto.Empty{}, nil
}

// sends a message to all streams in the streams map
func sendToAll(streams map[string]*proto.ChittyChat_JoinServer, message *proto.Message) {
	for _, stream := range streams {
		(*stream).Send(message)
	}
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log-file.log", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log-file.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
