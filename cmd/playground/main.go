package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/yookoala/botgame-playground/playground/comms"
)

type GameStage int

const (
	GameStageWaiting GameStage = iota
	GameStageSetup
	GameStagePlaying
	GameStageEnded
)

type dummyGame struct {
	stage GameStage

	player1 *comms.Session
	player2 *comms.Session

	sc comms.SessionCollection

	lock *sync.Mutex
}

func NewDummyGame(sc comms.SessionCollection) *dummyGame {
	return &dummyGame{
		sc:   sc,
		lock: &sync.Mutex{},
	}
}

func (g *dummyGame) HandleMessage(ctx context.Context, min comms.Message, mw comms.MessageWriter) error {

	if min.Type() != "request" {
		return fmt.Errorf("invalid request type: %v", min.Type())
	}
	log.Printf("received message: %s", min)

	// Resolve session id from context.
	sessionID := comms.GetSessionID(ctx)

	switch g.stage {
	case GameStageWaiting:
		// TODO: more sophisticated player joinning request / response.
		if g.player1 == nil && g.sc.Has(sessionID) {
			log.Printf("adding session as player 1: %s", sessionID)
			g.lock.Lock()
			g.player1 = g.sc.Get(sessionID)
			g.lock.Unlock()
			resp, err := comms.NewMessageFromJSONString(fmt.Sprintf(
				`{
					"type": "response",
					"sessionID": %#v,
					"playerID": "player1",
					"response": "success",
					"code": 200,
					"message": "You have joined the game."
				}`,
				sessionID,
			))
			if err != nil {
				log.Printf("error creating response message: %s", err)
				return err
			}

			err = mw.WriteMessage(resp)
			if err != nil {
				log.Printf("error sending response message: %s", err)
				g.lock.Lock()
				g.player1 = nil // unset player1
				g.lock.Unlock()
				return err
			}

			log.Printf("response send to player 1: %s", resp)
		}

		if g.player2 == nil && g.sc.Has(sessionID) {
			log.Printf("adding session as player 2: %s", sessionID)
			g.lock.Lock()
			g.player2 = g.sc.Get(sessionID)
			g.lock.Unlock()
			resp, err := comms.NewMessageFromJSONString(fmt.Sprintf(
				`{
					"type": "response",
					"sessionID": %#v,
					"playerID": "player2",
					"response": "success",
					"code": 200,
					"message": "You have joined the game."
				}`,
				sessionID,
			))
			if err != nil {
				log.Printf("error creating response message: %s", err)
				return err
			}

			err = mw.WriteMessage(resp)
			if err != nil {
				log.Printf("error sending response message: %s", err)
				g.lock.Lock()
				g.player2 = nil // unset player1
				g.lock.Unlock()
				return err
			}

			log.Printf("response send to player 1: %s", resp)
		}

		log.Printf("still here: %#v, %#v", g.player1, g.player2)

		// After both player has joinned and all setup done
		// start accepting game setup request.
		if g.player1 != nil && g.player2 != nil {
			log.Print("move on to setup stage")
			mw.WriteMessage(comms.MustMessage(comms.NewMessageFromJSONString(`{
				"type": "event",
				"event": "accept_setup"
			}`)))
			g.stage = GameStageSetup
		}

	case GameStageSetup:
		// TODO: implement me
		// only echoing the message for now
		log.Printf("setup: received message: %s", min)
		mw.WriteMessage(comms.MustMessage(comms.NewMessageFromJSONString(fmt.Sprintf(`{
			"type": "response",
			"sessionID": %#v,
			"response": "pong"
		}`, sessionID))))
	}
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

	sc := comms.NewSessionCollection()

	// Prepare the input (mq) and output (mw) ends of the game.
	mq := comms.NewSimpleMessageQueue(sc, 0) // Fan-in session messages
	mw := comms.NewSimpleMessageBroker(sc)   // Broke messages to sessions

	// Compose the game with the input and output ends.
	mq.Start(NewDummyGame(sc), mw)

	// Start passing socket request to the message queue.
	comms.StartService(l, mq)
}
