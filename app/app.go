package app

import (
	"github.com/tendermint/tendermint/abci/types"
)

type Application struct {
	types.BaseApplication

	RetainBlocks int64 // blocks to retain after commit (via ResponseCommit.RetainHeight)
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	resQuery.Log = "yoloooo"
	return resQuery
}
