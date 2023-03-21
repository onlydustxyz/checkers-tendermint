package app

import (
	"encoding/json"

	"github.com/batkinson/checkers-go/checkers"
	"github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	stateKey = []byte("stateKey")
)

type Application struct {
	types.BaseApplication

	db           dbm.DB
	state        State
	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

type State struct {
	game   *checkers.Game
	height int64
}

func loadState(db dbm.DB) State {
	var state State
	stateBytes, err := db.Get(stateKey)
	if err != nil {
		panic(err)
	}
	if len(stateBytes) == 0 {
		state.game = checkers.New()
		state.height = 0
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func saveState(state State, db dbm.DB) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	err = db.Set(stateKey, stateBytes)
	if err != nil {
		panic(err)
	}
}

func NewApplication() *Application {
	db := dbm.NewMemDB()
	state := loadState(db)

	return &Application{state: state, db: db}
}

func (app *Application) Info(req types.RequestInfo) types.ResponseInfo {
	return types.ResponseInfo{Data: string(app.state.game.String()), Version: "0.0.1", LastBlockHeight: app.state.height}
}

func (app *Application) Commit() types.ResponseCommit {
	saveState(app.state, app.db)
	app.state.height++

	return types.ResponseCommit{}
}

func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	var path = string(reqQuery.Data)
	resQuery.Height = app.state.height
	switch path {
	case "/store/board":
		resQuery.Log = "found board"
		resQuery.Value = []byte(app.state.game.String())
	case "/store/turn":
		resQuery.Log = "found turn"
		resQuery.Value = []byte(app.state.game.Turn.Color)
	default:
		resQuery.Log = "path not found"
	}
	return
}

func (app *Application) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	game, err := checkers.Parse(string(req.AppStateBytes))
	if err != nil {
		panic(err)
	}
	app.state.game = game
	app.state.height = req.InitialHeight
	app.state.game.Turn = checkers.BLACK_PLAYER

	return types.ResponseInitChain{
		AppHash:    req.AppStateBytes,
		Validators: req.Validators,
	}
}
