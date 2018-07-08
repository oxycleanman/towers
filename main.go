package main

import (
	"github.com/oxycleanman/towers/game"
	"github.com/oxycleanman/towers/gui"
)
import (
	_ "net/http/pprof"
	"net/http"
)

func main() {
	game := game.NewGame()
	go func() {
		game.Run()
	}()

	// Start pprof profiling monitor to fix this leaky ass game
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	ui := gui.NewUi(game.InputChan, game.LevelChan)
	ui.Run()
}
