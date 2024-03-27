package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/yookoala/botgame-playground/comms"
	"github.com/yookoala/botgame-playground/examples/battleship/game"
)

type gameClient struct {
	stage game.GameStage
}

func (c *gameClient) HandleMessage(ctx context.Context, m comms.Message, mw comms.MessageWriter) (err error) {

	switch c.stage {
	case game.GameStageWaiting:
		switch m.Type() {
		case "signal":
			sig := m.(comms.Signal)
			if sig.Signal() != "client:init" {
				return fmt.Errorf("received unexpected message in setup stage. expected signal of client:init, got: %s", m)
			}

			// Annonce join game
			err = mw.WriteMessage(comms.NewRequest("", "join", nil))
			if err != nil {
				log.Fatal(err)
			}
			return
		case "event":
			evt := m.(comms.Event)
			if evt.EventType() == "stage:change" {
				log.Printf("received stage change message: %s", m)
				evt.ReadDataTo(&c.stage)
				log.Printf("stage changed to %s", c.stage)
			}
			if c.stage != game.GameStageSetup {
				log.Printf("received unexpected event: %s", m)
				return
			}

			// Handle setup event
			return c.HandleMessage(ctx, comms.NewSignal("client:setup", nil), mw)
		}

	case game.GameStageSetup:
		if m.Type() != "signal" {
			return fmt.Errorf("received unexpected message in setup stage. expected signal, got: %s", m)
		}
		sig := m.(comms.Signal)
		if sig.Signal() != "client:setup" {
			return fmt.Errorf("received unexpected message in setup stage. expected signal of client:setup, got: %s", m)
		}

		// send the ship allocations to game server then wait.
		ships := make([]*game.ShipState, 5)
		ships[0], _ = game.NewShipState(game.ShipIDCarrier, [2]int{0, 0}, game.ShipDirectionToRight)
		ships[1], _ = game.NewShipState(game.ShipIDBattleship, [2]int{0, 1}, game.ShipDirectionToRight)
		ships[2], _ = game.NewShipState(game.ShipIDCruiser, [2]int{0, 2}, game.ShipDirectionToRight)
		ships[3], _ = game.NewShipState(game.ShipIDSubmarine, [2]int{0, 3}, game.ShipDirectionToRight)
		ships[4], _ = game.NewShipState(game.ShipIDDestroyer, [2]int{0, 4}, game.ShipDirectionToRight)
		err = mw.WriteMessage(comms.NewRequest("", "setup", ships))
		return
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
	defer conn.Close()

	// Listen to OS signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for {
			sig := <-c

			// Check if sig is interrupt (Ctrl+C)
			if sig.String() == "interrupt" {
				conn.Close()
			} else {
				fmt.Printf("received system signal: %s\n", sig.String())
			}
		}
	}()

	// Create a game client
	cli := NewGameClient()
	comms.StartClient(cli, conn)
}
