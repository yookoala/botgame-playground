package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func echoServer(sess *comms.Session) error {
	defer sess.Close()
	log.Printf("Session started: sessionID=%s", sess.ID())
	sess.WriteMessage(comms.NewGreeting(sess.ID()))
	for {
		m, err := sess.ReadMessage()
		if err != nil {
			// Check if error is eof
			if err == io.EOF {
				log.Print("Client disconnecting")
			} else {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
		log.Printf("Received: %v", m)
	}
	log.Println("Client disconnected")
	return nil
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
				close(c)
			} else {
				fmt.Printf("Received signal: %s\n", sig.String())
			}
		}
	}()

	comms.StartServer(l, comms.SessionHandlerFunc(echoServer))
}
