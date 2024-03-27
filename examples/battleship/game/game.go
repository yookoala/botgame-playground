package game

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
