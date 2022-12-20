package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	proto "whatTime/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Flags allows for user specific arguments/values
var clientName = flag.String("name", "Alice", "Senders name")
var serverPort = flag.String("sPort", "8080", "server port number")

var lamportClock = int64(0)       //the lamport clock
var server proto.ChittyChatClient //the server
var ServerConn *grpc.ClientConn   //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- Welcome to Chitty Chat ---")

	//log to file instead of console
	f := setLog()
	defer f.Close()

	//connect to server and close the connection when program closes
	connectToServer()
	defer ServerConn.Close()
	go joinChat()

	// start allowing user input
	parseAndSendInput()
}

func connectToServer() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	fmt.Printf("client %s: Attempts to dial on port %s\n", *clientName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		fmt.Printf("Fail to Dial : %v", err)
		return
	}

	server = proto.NewChittyChatClient(conn)
	ServerConn = conn
}

func joinChat() {
	joinRequest := &proto.JoinRequest{
		User:         *clientName,
		LamportClock: 0,
	}

	log.Println(*clientName, "is joining the chat")
	stream, _ := server.Join(context.Background(), joinRequest)

	for {
		select {
		case <-stream.Context().Done():
			fmt.Println("Connection to server closed")
			return // stream is done
		default:
		}

		incoming, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Server is done sending messages")
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive message from channel. \nErr: %v", err)
		}

		if incoming.LamportClock > lamportClock {
			lamportClock = incoming.LamportClock + 1
		} else {
			lamportClock++
		}

		log.Printf("%s got message from %s: %s", *clientName, incoming.User, incoming.Message)
		fmt.Printf("\rLamport: %v | %v: %v \n", lamportClock, incoming.User, incoming.Message)
		fmt.Print("-> ")
	}
}

func parseAndSendInput() {
	reader := bufio.NewReader(os.Stdin)

	//Infinite loop to listen for clients input.
	fmt.Print("-> ")
	for {
		//Read user input to the next newline
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim whitespace

		// we are sending a message so we increment the lamport clock
		lamportClock++

		// publish the message in the chat
		response, err := server.Publish(context.Background(), &proto.Message{
			User:         *clientName,
			Message:      input,
			LamportClock: lamportClock,
		})

		if err != nil || response == nil {
			log.Printf("Client %s: something went wrong with the server :(", *clientName)
			continue
		}
	}
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log-file.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
