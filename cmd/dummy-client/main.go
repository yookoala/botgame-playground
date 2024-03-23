package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	// Connect to ./echo.sock
	// Listen to single-line JSON messages from the server.
	// Send a single-line JSON message with a greeting message.
	// Close the connection

	conn, err := net.Dial("unix", "./echo.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create a reader from the connection
	reader := bufio.NewReader(conn)

	// Read a line from the connection
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	// Print the received message
	log.Println("Received:", message)

	// Send a message to the socket
	conn.Write([]byte("Hello, server!\n"))
	conn.Close()
}
