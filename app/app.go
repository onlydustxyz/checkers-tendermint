package app

import (
	"github.com/tendermint/tendermint/abci/types"
	"github.com/batkinson/checkers-go/checkers"
)

type Application struct {
	types.BaseApplication

	game *checkers.Game
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

func NewApplication() *Application {
	return &Application{game: checkers.New()}
}

func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	var path = string(reqQuery.Data)
	switch path {
	case "/store/board":
		resQuery.Log = "found board"
		resQuery.Value = []byte(app.game.String())
	case "/store/turn":
		resQuery.Log = "found turn"
		resQuery.Value = []byte(app.game.Turn.Color)
	default:
		resQuery.Log = "path not found"
	}
	return resQuery
}
