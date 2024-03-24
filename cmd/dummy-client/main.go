package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func waitErrorOnce(fn func() error) <-chan error {
	ch := make(chan error)
	go func() {
		ch <- fn()
	}()
	return ch
}

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

	sess, m, err := comms.NewSessionFromConn(conn)
	if err != nil {
		log.Fatal(err)
	}

	// Print the received greeting message
	log.Printf("Received: %v\n", m)

	// Listen to OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case sig := <-c:
			// Check if sig is interrupt (Ctrl+C)
			if sig.String() == "interrupt" {
				conn.Close()
			} else {
				fmt.Printf("Received signal: %s\n", sig.String())
			}
		case err := <-waitErrorOnce(func() error {
			return sess.WriteMessage(comms.MustMessage(comms.NewMessageFromJSONString(
				fmt.Sprintf(`{"sessionID": "%s", "type":"message","data":"Hello, server!"}`, sess.ID()),
			)))
		}):
			if err == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			switch err.(type) {
			case *net.OpError:
				log.Print("Socket closed. Quit")
				os.Exit(0)
			default:
				log.Printf("Socket error: %v", err)
				os.Exit(1)
			}
		}
	}
}
