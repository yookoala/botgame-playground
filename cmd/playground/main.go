package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func echoServer(conn net.Conn) {
	defer conn.Close()

	// Create a session ID for the client.
	// Send the client a single-line JSON message with the session ID and a greeting message.
	// Listen for singlie-line JSON messages from the client.
	// When a message is received, parse the JSON message and log.
	// When the client disconnects, log the session ID and a goodbye message.

	respEncoder := json.NewEncoder(conn)
	respEncoder.Encode(comms.NewGreeting("123", "Hello, client!"))

	bufSize := 1024
	buf := make([]byte, bufSize)
	for b, err := conn.Read(buf); err == nil; b, err = conn.Read(buf) {
		log.Printf("Received: %s", buf[:b])
	}
	fmt.Println("Client disconnected")
}

func main() {
	// Create a socket for connection.
	// When a client connects that socket:
	// 1. Create a session ID for the client.
	// 2. Send the client a single-line JSON message with the session ID and a greeting message.
	// 3. Listen for singlie-line JSON messages from the client.
	// 4. When a message is received, parse the JSON message and log.
	// 5. When the client disconnects, log the session ID and a goodbye message.

	// Start the server.
	// Listen for incoming connections.
	// When a connection is received, handle the connection in a goroutine.
	// When the server is stopped, close the socket.

	// Create a socket
	l, err := net.Listen("unix", "./echo.sock")
	if err != nil {
		println("listen error", err.Error())
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// Check if sig is interrupt (Ctrl+C)
			if sig.String() == "interrupt" {
				l.Close()
			} else {
				fmt.Printf("Received signal: %s\n", sig.String())
			}
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			switch err.(type) {
			case *net.OpError:
				log.Print("Socket closed. Quit")
				os.Exit(0)
			default:
				log.Printf("Socket error: %v", err)
				os.Exit(1)
			}
		}

		go echoServer(conn)
	}
}
