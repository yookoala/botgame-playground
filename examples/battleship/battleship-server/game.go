package main

/*

type BoardCellState int

const (
	BoardCellStateUnknown BoardCellState = 0
	BoardCellStateHit     BoardCellState = 1 << iota
	BoardCellStateMiss
	BoardCellStateSunk
)

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
