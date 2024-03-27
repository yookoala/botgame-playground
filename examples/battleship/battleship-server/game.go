package main

/*

type BoardCellState int

const (
	BoardCellStateUnknown BoardCellState = 0
	BoardCellStateHit     BoardCellState = 1 << iota
	BoardCellStateMiss
	BoardCellStateSunk
)

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
	ShipDirectionHorizontal ShipDirection = 0
	ShipDirectionVertical   ShipDirection = 1
)

type GameStage int

const (
	GameStageWaiting GameStage = 0
	GameStageSetup   GameStage = iota
	GameStagePlaying
	GameStageEnded
)

func (g *dummyGame) Start() {
	messageIn := make(chan comms.Message)

	go func(sess *comms.Session, out chan<- comms.Message) {

	}(g.player1Session, messageIn)
}

type ShipState struct {
	ID          ShipID
	HP          int
	Coordinates [][2]int
}

func NewShipState(id ShipID, cord [][2]int) (state *ShipState, err error) {

	// Check if coordinates are empty
	if len(cord) == 0 {
		err = fmt.Errorf("coordinates are empty")
		return
	}

	// Check if ShipID is valid
	if id.IsValid() {
		err = fmt.Errorf("invalid ShipID: %v", id)
		return
	}

	// Check all coordinates are connected as a straightline
	for i := 1; i < len(cord); i++ {
		if cord[i][0] != cord[i-1][0] && cord[i][1] != cord[i-1][1] {
			err = fmt.Errorf("coordinates are not connected: %v", cord)
			return
		}
	}

	// Check number of coordinates matches ship size
	if len(cord) != id.Size() {
		err = fmt.Errorf("number of coordinates does not match ship size: %v", cord)
		return
	}

	for x := range cord {
		for y := range cord[x] {
			// Check coordinate within bound
			if cord[x][y] < 0 || cord[x][y] > 9 {
				err = fmt.Errorf("coordinate out of bounds: %v", cord[x][y])
				return
			}

			// Check coordinate is not duplicated
			for i := range cord {
				if i != x && cord[i][0] == cord[x][0] && cord[i][1] == cord[x][1] {
					err = fmt.Errorf("coordinate duplicated: %v", cord[x])
					return
				}
			}
		}
	}

	state = &ShipState{
		ID:          id,
		HP:          len(cord),
		Coordinates: cord,
	}
	return
}

type PlayerState struct {
	Ready      bool
	Ships      [5]ShipState
	OwnBoard   [10][10]BoardCellState
	EnemyBoard [10][10]BoardCellState
}

// FindOwnShipAt returns the ship at the given coordinates.
func (p *PlayerState) FindOwnShipAt(x, y int) *ShipState {
	for i := range p.Ships {
		for _, c := range p.Ships[i].Coordinates {
			if c[0] == x && c[1] == y {
				return &p.Ships[i]
			}
		}
	}
	return nil
}

// IsShipSunk returns true if the ship with the given ID is sunk.
func (p *PlayerState) IsShipSunk(shipID ShipID) bool {
	for i := range p.Ships {
		if p.Ships[i].ID == shipID {
			return p.Ships[i].HP == 0
		}
	}
	return false
}

// IsAllShipsSunk returns true if all ships are sunk.
func (p *PlayerState) IsAllShipsSunk() bool {
	for i := range p.Ships {
		if p.Ships[i].HP > 0 {
			return false
		}
	}
	return true
}

// IsAllShipsPlaced returns true if all ships are placed.
func (p *PlayerState) IsAllShipsPlaced() bool {
	for i := range p.Ships {
		if p.Ships[i].HP > 0 {
			return false
		}
	}
	return true
}

func (p *PlayerState) ReceiveHit(x, y int) (s BoardCellState, err error) {
	if p.OwnBoard[x][y] != BoardCellStateUnknown {
		err = fmt.Errorf("cell already hit")
		return
	}

	s = BoardCellStateMiss

shipInspectionLoop:
	for i := range p.Ships {
		for j := range p.Ships[i].Coordinates {
			if p.Ships[i].Coordinates[j][0] == x && p.Ships[i].Coordinates[j][1] == y {
				p.Ships[i].HP--
				if p.Ships[i].HP == 0 {
					s = BoardCellStateSunk
					break shipInspectionLoop
				} else {
					s = BoardCellStateHit
					break shipInspectionLoop
				}
			}
		}
	}

	// Remember on own board.
	p.OwnBoard[x][y] = s
	return
}
*/
