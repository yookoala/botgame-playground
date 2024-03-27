package game

import "fmt"

const (
	BoardSize = 10
)

type GameStage int

const (
	GameStageWaiting GameStage = iota
	GameStageSetup
	GameStagePlaying
	GameStageEnded
)

func (s GameStage) String() string {
	switch s {
	case GameStageWaiting:
		return "Waiting"
	case GameStageSetup:
		return "Setup"
	case GameStagePlaying:
		return "Playing"
	case GameStageEnded:
		return "Ended"
	default:
		return "Unknown"
	}
}

type ShipID int

const (
	ShipIDCarrier ShipID = iota
	ShipIDBattleship
	ShipIDCruiser
	ShipIDSubmarine
	ShipIDDestroyer
)

func (s ShipID) String() string {
	switch s {
	case ShipIDCarrier:
		return "Carrier"
	case ShipIDBattleship:
		return "Battleship"
	case ShipIDCruiser:
		return "Cruiser"
	case ShipIDSubmarine:
		return "Submarine"
	case ShipIDDestroyer:
		return "Destroyer"
	default:
		return "Unknown"
	}
}

func (s ShipID) IsValid() bool {
	return s >= ShipIDCarrier && s <= ShipIDDestroyer
}

func (s ShipID) Size() int {
	switch s {
	case ShipIDCarrier:
		return 5
	case ShipIDBattleship:
		return 4
	case ShipIDCruiser:
		return 3
	case ShipIDSubmarine:
		return 3
	case ShipIDDestroyer:
		return 2
	default:
		return 0
	}
}

type ShipDirection int

const (
	ShipDirectionToRight ShipDirection = iota
	ShipDirectionToDown
)

func (d ShipDirection) String() string {
	switch d {
	case ShipDirectionToRight:
		return "Right"
	case ShipDirectionToDown:
		return "Down"
	default:
		return "Unknown"
	}
}

type ShipState struct {
	ID          ShipID
	HP          int
	Coordinates [][2]int
}

func NewShipState(id ShipID, cord [2]int, dir ShipDirection) (state *ShipState, err error) {
	// Check if ShipID is valid
	if !id.IsValid() {
		err = fmt.Errorf("invalid ShipID: %v", id)
		return
	}

	// Check coordinate within bound
	if cord[0] < 0 || cord[0] >= BoardSize || cord[1] < 0 || cord[1] >= BoardSize {
		err = fmt.Errorf("coordinate out of bounds: %v", cord)
		return
	}

	// Check coordinate with direction. Then generate the cords
	cords := make([][2]int, id.Size())
	if dir == ShipDirectionToRight {
		if cord[0]+id.Size()-1 >= BoardSize {
			err = fmt.Errorf("coordinate out of bounds: cord=%v, dir=%s, size=%d", cord, dir, id.Size())
			return
		}
		for i := 0; i < id.Size(); i++ {
			cords[i] = [2]int{cord[0] + i, cord[1]}
		}
	} else {
		if cord[1]+id.Size()-1 >= BoardSize {
			err = fmt.Errorf("coordinate out of bounds: cord=%v, dir=%s, size=%d", cord, dir, id.Size())
			return
		}
		for i := 0; i < id.Size(); i++ {
			cords[i] = [2]int{cord[0], cord[1] + i}
		}
	}

	// Generate ship state.
	state = &ShipState{
		ID:          id,
		HP:          id.Size(),
		Coordinates: cords,
	}
	return
}
