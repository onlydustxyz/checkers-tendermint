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

type StoredGame struct {
	Creator string // A stringified address for the creator of the game.
	Black   string // A stringified address for the player playing blacks.
	Red     string // A stringified address for the player playing reds.
}

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

func readTx(tx []byte) (checkers.Pos, checkers.Pos) {
	return checkers.Pos{X: int(tx[0]), Y: int(tx[1])}, checkers.Pos{X: int(tx[2]), Y: int(tx[3])}
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

func (app *Application) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	start, end := readTx(req.Tx)
	valid := start.X >= 0 && start.X < 8 && start.Y >= 0 && start.Y < 8 && end.X >= 0 && end.X < 8 && end.Y >= 0 && end.Y < 8
	if valid {
		return types.ResponseCheckTx{Code: types.CodeTypeOK}
	}
	return types.ResponseCheckTx{Code: 1, Log: "Invalid move"}
}

func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	start, end := readTx(req.Tx)
	captured_checker, err := app.state.game.Move(start, end)
	if err != nil {
		return types.ResponseDeliverTx{Code: 1, Log: err.Error()}
	}
	var hasCaptured byte
	if captured_checker != checkers.NO_POS {
		hasCaptured = 1
	} else {
		hasCaptured = 0
	}
	events := []types.Event{
		{
			Type: "move",
			Attributes: []types.EventAttribute{
				{Key: []byte("has-captured"), Value: []byte{hasCaptured}, Index: true},
			},
		},
	}
	return types.ResponseDeliverTx{Code: types.CodeTypeOK, Events: events, GasUsed: 1}
}
