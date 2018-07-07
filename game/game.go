package game

import (
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"os"
)

type Game struct {
	InputChan chan *Input
	LevelChan chan *Level
	Level     *Level
}

type Level struct {
	Player             *Player
	Enemies            []*Enemy
	Bullets            []*Bullet
	PrimaryFirePressed bool
	EnemySpawnTimer    int
}

type InputType int
const (
	None          InputType = iota
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
	Type    InputType
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
	Direction  float64
	Speed      float64
}

type Entity struct {
	Pos
	Size
	TextureName string
	Texture     *sdl.Texture
	FireOffsetX int
	FireOffsetY int
}

type Character struct {
	Entity
	Velocity
	Cost                          int
	Level                         int
	Hitpoints                     int
	Strength                      int
	DestroyedAnimationTextureName string
	DestroyedAnimationPlayed      bool
	DestroyedAnimationCounter     int
	DestroyedSoundPlayed bool
	IsDestroyed                   bool
	FireRateTimer                 int
	FireRateResetValue            int
	IsFiring                      bool
}

type Shooter interface {
	// Should return FireRateTimer, FireRateResetValue, and whether the entity is the player
	GetFireSettings() (int, int, bool)
	SetFireTimer(int)
	GetSelf() *Character
}

func NewGame() *Game {
	game := &Game{}
	game.InputChan = make(chan *Input, 2)
	game.LevelChan = make(chan *Level, 2)

	game.Level = &Level{}
	game.Level.initPlayer()
	game.Level.EnemySpawnTimer = 0

	return game
}

func (game *Game) handleInput(input *Input) {
	if input.Pressed {
		switch input.Type {
		case Up:
			if game.Level.Player.Yvel >= -5 {
				game.Level.Player.Yvel -= 5
			}
			break
		case Down:
			if game.Level.Player.Yvel <= 5 {
				game.Level.Player.Yvel += 5
			}
			break
		case Left:
			if game.Level.Player.Xvel >= -5 {
				game.Level.Player.Xvel -= 5
			}
			break
		case Right:
			if game.Level.Player.Xvel <= 5 {
				game.Level.Player.Xvel += 5
			}
			break
		case FirePrimary:
			game.Level.Player.IsFiring = true
			game.Level.Player.FireRateTimer = game.Level.Player.FireRateResetValue
			break
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
		case FirePrimary:
			game.Level.Player.IsFiring = false
			break
		default:
			//fmt.Println("Some input not pressed")
		}
	}
}

func DegreeToRad(degree float64) float64 {
	return degree * (math.Pi / 180)
}

func (game *Game) Run() {
	game.LevelChan <- game.Level

	for input := range game.InputChan {
		if input.Type == None {
			continue
		}
		if input.Type == Quit {
			close(game.LevelChan)
			close(game.InputChan)
			os.Exit(0)
		}
		game.handleInput(input)

		game.LevelChan <- game.Level
	}
}
