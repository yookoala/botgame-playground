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
	ShipUndefined ShipID = iota
	ShipIDCarrier
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

// ShipPlacement represents a ship's placement.
type ShipPlacement struct {
	ID         ShipID
	Coordinate [2]int
	Direction  ShipDirection
}

// NewShipPlacement creates a new ShipPlacement.
func NewShipPlacement(id ShipID, cord [2]int, dir ShipDirection) (sp *ShipPlacement, err error) {
	sp = &ShipPlacement{
		ID:         id,
		Coordinate: cord,
		Direction:  dir,
	}
	return
}

func (sp ShipPlacement) String() string {
	return fmt.Sprintf("%s at (%d, %d) %s", sp.ID, sp.Coordinate[0], sp.Coordinate[1], sp.Direction)
}

// ToShipState converts ShipPlacement to ShipState.
func (sp ShipPlacement) ToShipState() (state *ShipState, err error) {
	return NewShipState(sp.ID, sp.Coordinate, sp.Direction)
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

type ShipStates []ShipState

func (s ShipStates) Validate() error {
	if s == nil {
		return fmt.Errorf("nil pointer")
	}

	// Check if all ships are placed.
	for i := ShipIDCarrier; i <= ShipIDDestroyer; i++ {
		found := false
		for j := range s {
			if s[j].ID == i {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("ship %v not placed", i)
		}
	}

	// Check if ships have duplicated coordinate(s)
	coords := make(map[[2]int]bool)
	for i := range s {
		for j := range s[i].Coordinates {
			if coords[s[i].Coordinates[j]] {
				return fmt.Errorf("duplicated coordinate: (%d, %d)",
					s[i].Coordinates[j][0], s[i].Coordinates[j][1])
			}
			coords[s[i].Coordinates[j]] = true
		}
	}
	return nil
}

func (s ShipStates) Initialize() error {
	if s == nil {
		return fmt.Errorf("nil pointer")
	}
	for i, l := 0, len(s); i < l; i++ {
		s[i].HP = s[i].ID.Size()
	}
	return nil
}

type BoardCellState int

const BoardCellStateUnknown BoardCellState = 0

const (
	BoardCellStateHit BoardCellState = 1 << iota
	BoardCellStateMiss
	BoardCellStateSunk
)

func (s BoardCellState) String() string {
	switch s {
	case BoardCellStateUnknown:
		return "Unknown"
	case BoardCellStateHit:
		return "Hit"
	case BoardCellStateMiss:
		return "Miss"
	case BoardCellStateSunk:
		return "Sunk"
	default:
		return "Unknown"
	}
}
