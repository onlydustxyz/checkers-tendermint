package main

import (
    "fmt"
    "os"
    "github.com/tendermint/tendermint/libs/log"
    tmos "github.com/tendermint/tendermint/libs/os"
    "github.com/tendermint/tendermint/abci/server"
    checkers "tuto/game/app"
)

var logger log.Logger

func main() {
    fmt.Println("Starting the app")
    var app = checkers.NewApplication()
    logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

    // Start the listener
    srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
    if err != nil {
        return
    }
    srv.SetLogger(logger.With("module", "abci-server"))
    if err := srv.Start(); err != nil {
        return
    }

    // Stop upon receiving SIGTERM or CTRL-C.
    tmos.TrapSignal(logger, func() {
        // Cleanup
        srv.Stop()
    })

    // Run forever.
    select {}
}
