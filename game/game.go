package game

import (
	"os"
)

type Game struct {
	InputChan chan *Input
	LevelChan chan *Level
	Level *Level
}

type Level struct {
	Player *Player
	Enemies []*Enemy
	Bullets []*Bullet
}

type Direction float64
const (
	DDown Direction = 0.0
	DUp Direction = 180.0
	DLeft Direction = 90.0
	DRight Direction = 270.0
)

type InputType int
const (
	None InputType = iota
	Up
	Down
	Left
	Right
	Quit
	FirePrimary
	FireSecondary
	Pause
)

type Input struct {
	Pos
	Type InputType
	Pressed bool
}

type Pos struct {
	X, Y int
}

type Size struct {
	W, H int
}

type Velocity struct {
	Xvel, Yvel int
	Direction Direction
}

type Entity struct {
	Pos
	Size
	TextureName string
	FireOffsetX int
	FireOffsetY int
}

type Character struct {
	Entity
	Velocity
	Hitpoints int
	Speed float64
}

type Tile struct {
	Entity
	Rune rune
}

func NewGame() *Game {
	game := &Game{}
	game.InputChan = make(chan *Input, 2)
	game.LevelChan = make(chan *Level, 2)

	game.Level = &Level{}
	game.Level.Player = NewPlayer("tank_huge")
	return game
}

func (game *Game) handleInput(input *Input) {
	if input.Pressed {
		switch input.Type {
		case Up:
			if game.Level.Player.Yvel > -5 {
				game.Level.Player.Yvel--
			}
			game.Level.Player.Direction = DUp
			break
		case Down:
			if game.Level.Player.Yvel < 5 {
				game.Level.Player.Yvel++
			}
			game.Level.Player.Direction = DDown
			break
		case Left:
			if game.Level.Player.Xvel > -5 {
				game.Level.Player.Xvel--
			}
			game.Level.Player.Direction = DLeft
			break
		case Right:
			if game.Level.Player.Xvel < 5 {
				game.Level.Player.Xvel++
			}
			game.Level.Player.Direction = DRight
			break
		//case FirePrimary:
		//	game.leftButtonDown = true
		//	break
		//case FireSecondary:
		//	game.rightButtonDown = true
		//	break
		default:
			//fmt.Println("Some input pressed")
		}
	} else {
		switch input.Type {
		case Up:
			if game.Level.Player.Yvel < 0 {
				game.Level.Player.Yvel = 0
			}
			break
		case Down:
			if game.Level.Player.Yvel > 0 {
				game.Level.Player.Yvel = 0
			}
			break
		case Left:
			if game.Level.Player.Xvel < 0 {
				game.Level.Player.Xvel = 0
			}
			break
		case Right:
			if game.Level.Player.Xvel > 0 {
				game.Level.Player.Xvel = 0
			}
			break
		//case FirePrimary:
		//	game.leftButtonDown = false
		//	break
		//case FireSecondary:
		//	game.rightButtonDown = false
		//	break
		default:
			//fmt.Println("Some input not pressed")
		}
	}
}

func (game *Game) Run() {
	game.LevelChan <- game.Level

	for input := range game.InputChan {
		if input.Type == Quit {
			close(game.LevelChan)
			close(game.InputChan)
			os.Exit(0)
		}
		game.handleInput(input)

		game.LevelChan <- game.Level
	}

}