package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/yookoala/botgame-playground/comms"
	"github.com/yookoala/botgame-playground/examples/battleship/game"
)

type gameClient struct {
	stage game.GameStage
}

func (c *gameClient) HandleMessage(ctx context.Context, m comms.Message, mw comms.MessageWriter) (err error) {

	// Handle stage change event first.
	if m.Type() == "event" {
		evt := m.(comms.Event)
		if evt.EventType() == "stage:change" {
			log.Printf("received stage change message: %s", m)
			evt.ReadDataTo(&c.stage)
			log.Printf("stage changed to %s", c.stage)
		}
	}

	switch c.stage {
	case game.GameStageWaiting:
		if m.Type() != "signal" {
			return fmt.Errorf("invalid message type: %v", m.Type())
		}
		sig := m.(comms.Signal)
		if sig.Signal() != "client:init" {
			return fmt.Errorf("invalid signal type: %v", sig.Signal())
		}

		// Annonce join game
		err = mw.WriteMessage(comms.NewRequest("", "join", nil))
		if err != nil {
			log.Fatal(err)
		}
		return
	case game.GameStageSetup:
		// send the ship allocations to game server then wait.
		ships := make([]*game.ShipPlacement, 5)
		ships[0], _ = game.NewShipPlacement(game.ShipIDCarrier, [2]int{0, 0}, game.ShipDirectionToRight)
		ships[1], _ = game.NewShipPlacement(game.ShipIDBattleship, [2]int{0, 1}, game.ShipDirectionToRight)
		ships[2], _ = game.NewShipPlacement(game.ShipIDCruiser, [2]int{0, 2}, game.ShipDirectionToRight)
		ships[3], _ = game.NewShipPlacement(game.ShipIDSubmarine, [2]int{0, 3}, game.ShipDirectionToRight)
		ships[4], _ = game.NewShipPlacement(game.ShipIDDestroyer, [2]int{0, 4}, game.ShipDirectionToRight)
		err = mw.WriteMessage(comms.NewRequest("", "setup", ships))
		return
	case game.GameStagePlaying:
		if m.Type() == "event" {
			evt := m.(comms.Event)
			log.Printf("playing stage. Event: %s", evt.EventType())
			if evt.EventType() == "frame:update" {
				// It's our turn, send a shot
				log.Printf("frame update in playing stage")
				err = mw.WriteMessage(comms.NewRequest("", "shot", [2]int{0, 0}))
				return
			}
		}
	}

	return nil
}

func NewGameClient() *gameClient {
	return &gameClient{}
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

	// Listen to OS signals
	//comms.CloseOnSignal(conn, os.Interrupt)

	// Create a game client
	cli := NewGameClient()
	comms.StartClient(cli, conn)
}
