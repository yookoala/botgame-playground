package game_test

import (
	"fmt"
	"testing"

	"github.com/yookoala/botgame-playground/examples/battleship/game"
)

func TestShipsState_NoError(t *testing.T) {
	ss := make(game.ShipStates, 0)
	ss = append(ss, game.ShipState{
		ID: game.ShipIDCarrier,
		HP: 0,
		Coordinates: [][2]int{
			{0, 0},
			{0, 1},
			{0, 2},
			{0, 3},
			{0, 4},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDBattleship,
		HP: game.ShipIDBattleship.Size(),
		Coordinates: [][2]int{
			{1, 0},
			{1, 1},
			{1, 2},
			{1, 3},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDCruiser,
		HP: game.ShipIDCruiser.Size(),
		Coordinates: [][2]int{
			{2, 0},
			{2, 1},
			{2, 2},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDSubmarine,
		HP: game.ShipIDSubmarine.Size(),
		Coordinates: [][2]int{
			{3, 0},
			{3, 1},
			{3, 2},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDDestroyer,
		HP: game.ShipIDDestroyer.Size(),
		Coordinates: [][2]int{
			{4, 0},
			{4, 1},
		},
	})

	if err := ss.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestShipsState_MissingCruiser(t *testing.T) {
	ss := make(game.ShipStates, 0)
	ss = append(ss, game.ShipState{
		ID: game.ShipIDCarrier,
		HP: 0,
		Coordinates: [][2]int{
			{0, 0},
			{0, 1},
			{0, 2},
			{0, 3},
			{0, 4},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDBattleship,
		HP: game.ShipIDBattleship.Size(),
		Coordinates: [][2]int{
			{1, 0},
			{1, 1},
			{1, 2},
			{1, 3},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDSubmarine,
		HP: game.ShipIDSubmarine.Size(),
		Coordinates: [][2]int{
			{3, 0},
			{3, 1},
			{3, 2},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDDestroyer,
		HP: game.ShipIDDestroyer.Size(),
		Coordinates: [][2]int{
			{4, 0},
			{4, 1},
		},
	})

	if err := ss.Validate(); err == nil {
		t.Error("expected error but got nil")
	} else if want, have := fmt.Sprintf("ship %s not placed", game.ShipIDCruiser), err.Error(); want != have {
		t.Errorf("unexpected error: want=%q, have=%q", want, have)
	}
}

func TestShipsState_DuplicatedCoordinate(t *testing.T) {
	ss := make(game.ShipStates, 0)
	ss = append(ss, game.ShipState{
		ID: game.ShipIDCarrier,
		HP: 0,
		Coordinates: [][2]int{
			{0, 0},
			{0, 1},
			{0, 2},
			{0, 3},
			{0, 4},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDBattleship,
		HP: game.ShipIDBattleship.Size(),
		Coordinates: [][2]int{
			{1, 0},
			{1, 1},
			{1, 2},
			{1, 3},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDCruiser,
		HP: game.ShipIDCruiser.Size(),
		Coordinates: [][2]int{
			{2, 0},
			{2, 1},
			{2, 2},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDSubmarine,
		HP: game.ShipIDSubmarine.Size(),
		Coordinates: [][2]int{
			{3, 0},
			{3, 1},
			{3, 2},
		},
	})
	ss = append(ss, game.ShipState{
		ID: game.ShipIDDestroyer,
		HP: game.ShipIDDestroyer.Size(),
		Coordinates: [][2]int{
			{0, 4}, // duplicate coordinate with carrier
			{0, 5},
		},
	})

	if err := ss.Validate(); err == nil {
		t.Error("expected error but got nil")
	} else if want, have := fmt.Sprintf("duplicated coordinate: (%d, %d)", 0, 4), err.Error(); want != have {
		t.Errorf("unexpected error: want=%q, have=%q", want, have)
	}
}
