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

	playerState map[*comms.Session]game.PlayerState

	lock *sync.Mutex

	frameRequests map[string]comms.Request
}

func NewDummyGame() *dummyGame {
	return &dummyGame{
		lock: &sync.Mutex{},

		playerState: make(map[*comms.Session]game.PlayerState),

		frameRequests: make(map[string]comms.Request),
	}
}

func (g *dummyGame) IsPlayerSession(sessionID string) bool {
	return g.GetPlayerSession(sessionID) != nil
}

func (g *dummyGame) GetPlayerSession(sessionID string) *comms.Session {
	if g.player1 != nil && g.player1.ID() == sessionID {
		return g.player1
	}
	if g.player2 != nil && g.player2.ID() == sessionID {
		return g.player2
	}
	return nil
}

func (g *dummyGame) HandleMessage(ctx context.Context, min comms.Message, mw comms.MessageWriter) error {

	if min.Type() != "request" {
		return fmt.Errorf("invalid request type: %v", min.Type())
	}
	// log.Printf("received message: %s, game stage: %s", min, g.stage)

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
				g.lock.Lock()
				g.stage = game.GameStageSetup
				g.lock.Unlock()
				mw.WriteMessage(comms.NewEvent("stage:change", game.GameStageSetup))
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

		ships := make([]game.ShipPlacement, 5)
		req.ReadDataTo(&ships)

		// Only allow player to setup their own ships.
		s := g.GetPlayerSession(comms.GetSessionID(ctx))
		if s == nil {
			mw.WriteMessage(comms.NewErrorResponse(
				sessionID,
				req.RequestID(),
				403,
				"error",
				"forbidden",
			))
			return nil
		}

		// Each player can only setup once.
		if _, ok := g.playerState[s]; ok {
			mw.WriteMessage(comms.NewErrorResponse(
				sessionID,
				req.RequestID(),
				400,
				"error",
				"state already set",
			))
			return nil
		}

		// Validate the ship placements.
		shipStates := make(game.ShipStates, len(ships))
		for i, sp := range ships {
			ss, err := sp.ToShipState()
			if err != nil {
				mw.WriteMessage(comms.NewErrorResponse(
					sessionID,
					req.RequestID(),
					400,
					"error",
					err.Error(),
				))
				return nil
			}
			shipStates[i] = *ss
		}

		if err := shipStates.Validate(); err != nil {
			mw.WriteMessage(comms.NewErrorResponse(
				sessionID,
				req.RequestID(),
				400,
				"error",
				err.Error(),
			))
			return nil
		}

		log.Printf("accepted setup: %v", shipStates)
		g.playerState[s] = game.PlayerState{
			Ready: true,
			Ships: shipStates,
		}

		if len(g.playerState) == 2 {

			// Announce stage change
			mw.WriteMessage(comms.NewEvent(
				"stage:change",
				game.GameStagePlaying,
			))

			// Resolve initial frame (frame 0)
			mw.WriteMessage(comms.NewEvent(
				"frame:update",
				nil,
			))

			// Change the game stage to playing
			g.lock.Lock()
			g.stage = game.GameStagePlaying
			g.lock.Unlock()
		}

	case game.GameStagePlaying:
		req := min.(comms.Request)
		switch req.RequestType() {
		case "shot":
			g.lock.Lock()
			if _, ok := g.frameRequests[sessionID]; !ok {
				// Only accept the first shot request.
				g.frameRequests[sessionID] = req
			}
			if len(g.frameRequests) == 2 {
				// Both players has submitted their shot.
				// Resolve the frame.
				log.Printf("here!")
			}
			g.lock.Unlock()
		}
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
	err = comms.StartServer(l, mq)
	if err != nil {
		log.Printf("Server ended with error: %s (%#v)", err, err)
	}
}
