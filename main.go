package main

import (
	"github.com/oxycleanman/towers/gui"
	"github.com/oxycleanman/towers/game"
)

func main() {
	game := game.NewGame()
	go func() {
		game.Run()
	}()

	ui := gui.NewUi(game.InputChan, game.LevelChan)
	ui.Run()
}
