package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/yookoala/botgame-playground/comms"
	"github.com/yookoala/botgame-playground/examples/battleship/game"
)

type dummyGame struct {
	stage game.GameStage

	player1 *comms.Session
	player2 *comms.Session

	lock *sync.Mutex
}

func NewDummyGame() *dummyGame {
	return &dummyGame{
		lock: &sync.Mutex{},
	}
}

func (g *dummyGame) HandleMessage(ctx context.Context, min comms.Message, mw comms.MessageWriter) error {

	if min.Type() != "request" {
		return fmt.Errorf("invalid request type: %v", min.Type())
	}
	//log.Printf("received message: %s", min)

	// Resolve context variables.
	sc := comms.GetSessionCollection(ctx)
	sessionID := comms.GetSessionID(ctx)

	if min.Type() == "signal" {
		// Ignore signal for now.
		return nil
	}

	switch g.stage {
	case game.GameStageWaiting:

		req := min.(comms.Request)
		switch req.RequestType() {
		case "join":
			// TODO: more sophisticated player joinning request / response.
			if g.player1 == nil && sc.Has(sessionID) {
				if g.player2 != nil && g.player2.ID() == sessionID {
					// player 1 cannot join again.
					// ignore for now.
					return nil
				}

				log.Printf("adding session as player 1: %s", sessionID)
				g.lock.Lock()
				g.player1 = sc.Get(sessionID)
				g.lock.Unlock()

				resp := comms.NewResponse(
					sessionID,
					req.RequestID(),
					req.RequestType(),
					200,
					"success",
					"player1",
				)

				err := mw.WriteMessage(resp)
				if err != nil {
					log.Printf("error sending response message: %s", err)
					g.lock.Lock()
					g.player1 = nil // unset player1
					g.lock.Unlock()
					return err
				}

				log.Printf("response send to player 1: %s", resp)
			} else if g.player2 == nil && sc.Has(sessionID) {
				if g.player1 != nil && g.player1.ID() == sessionID {
					// player 1 cannot join again.
					// ignore for now.
					return nil
				}
				log.Printf("adding session as player 2: %s", sessionID)
				g.lock.Lock()
				g.player2 = sc.Get(sessionID)
				g.lock.Unlock()

				resp := comms.NewResponse(
					sessionID,
					req.RequestID(),
					req.RequestType(),
					200,
					"success",
					"player2",
				)

				err := mw.WriteMessage(resp)
				if err != nil {
					log.Printf("error sending response message: %s", err)
					g.lock.Lock()
					g.player2 = nil // unset player1
					g.lock.Unlock()
					return err
				}

				log.Printf("response send to player 2: %s", resp)
			}

			// After both player has joinned and all setup done
			// start accepting game setup request.
			if g.player1 != nil && g.player2 != nil {
				log.Print("move on to setup stage")
				g.stage = game.GameStageSetup
				mw.WriteMessage(comms.NewEvent(sessionID, "stage:change", game.GameStageSetup))
			}

		case "subscribe":
			resp := comms.NewResponse(
				sessionID,
				req.RequestID(),
				req.RequestType(),
				200,
				"success",
				"subscribed",
			)
			err := mw.WriteMessage(resp)
			if err != nil {
				log.Printf("error sending response message: %s", err)
				g.lock.Lock()
				g.player2 = nil // unset player1
				g.lock.Unlock()
				return err
			}

		}

	case game.GameStageSetup:
		req := min.(comms.Request)
		if req.RequestType() != "setup" {
			return fmt.Errorf("invalid request type: %v", req.RequestType())
		}

		ships := make([]game.ShipState, 5)
		req.ReadDataTo(&ships)

		log.Printf("received setup request: %v", ships)
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

	// Listen to OS signals
	comms.CloseOnSignal(l, os.Interrupt)

	// Prepare the input (mq) and output (mw) ends of the game.
	sc := comms.NewSessionCollection()
	sc.OnAdd(func(s *comms.Session) {
		log.Printf("session added: %s, current len=%d", s.ID(), sc.Len())
	})
	sc.OnRemove(func(s *comms.Session) {
		log.Printf("session remove: %s, current len=%d", s.ID(), sc.Len())
	})
	mq := comms.NewSimpleMessageQueue(sc, 0) // Fan-in session messages
	mw := comms.NewSimpleMessageBroker(sc)   // Broke messages to sessions

	// Compose the game with the input and output ends.
	mq.Start(NewDummyGame(), mw)

	// Start passing socket request to the message queue.
	err = comms.StartListen(l, mq)
	if err != nil {
		log.Printf("Server ended with error: %s (%#v)", err, err)
	}
}
