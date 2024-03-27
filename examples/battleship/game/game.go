package game

type GameStage int

const (
	GameStageWaiting GameStage = iota
	GameStageSetup
	GameStagePlaying
	GameStageEnded
)
