package comms

import (
	"fmt"
	"log"
	"net"
	"os"
)

// getNewSessionIDs allocate a new session ID for use.
func getNewSessionIDs() <-chan string {
	ch := make(chan string)
	go func() {
		maxInt := int(^uint(0) >> 1)
		for id := 1; true; id++ {
			ch <- fmt.Sprintf("%d", id)
			if id == maxInt {
				id = 0
			}
		}
	}()
	return ch
}

// StartServer creates a new server loop and start listening to the listener.
func StartServer(listener net.Listener, sh SessionHandler) {
	defer listener.Close()

	newSessionIDs := getNewSessionIDs()
	for {
		sessionID := <-newSessionIDs
		conn, err := listener.Accept()
		if err != nil {
			switch err.(type) {
			case *net.OpError:
				log.Print("Socket closed. Quit")
				// TODO: some clean up to existing sessions?
				os.Exit(0)
			default:
				log.Printf("Socket error: %v", err)
				// TODO: some clean up to existing sessions?
				os.Exit(1)
			}
		}

		go sh.HandleSession(NewSession(sessionID, conn))
	}
}
