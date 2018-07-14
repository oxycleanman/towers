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
	Levels    []*Level
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
	LevelComplete
	PlayerDeath
)

type Input struct {
	Pos
	Type    InputType
	Pressed bool
}

type Pos struct {
	X, Y float64
	BoundBox *sdl.Rect
}

type Velocity struct {
	Xvel, Yvel float64
	Direction  float64
	Speed      float64
}

type Entity struct {
	Pos
	TextureName string
	Texture     *sdl.Texture
	FireOffsetX, FireOffsetY float64
}

type Character struct {
	Entity
	Velocity
	PointValue                    int32
	Hitpoints                     int32
	Strength                      int32
	DestroyedAnimationTextureName string
	DestroyedAnimationPlayed      bool
	DestroyedAnimationCounter     int
	DestroyedSoundPlayed          bool
	IsDestroyed                   bool
	FireRateTimer                 int
	FireRateResetValue            int
	IsFiring                      bool
	ShieldHitpoints               int32
	EngineFireAnimationCounter    int
}

type Shooter interface {
	// Should return FireRateTimer, FireRateResetValue, and whether the entity is the player
	GetFireSettings() (int, int, bool)
	SetFireTimer(int)
	GetSelf() *Character
}

// TODO: Which structs can have logic consolidated using interfaces like Shooter? AI interface for enemy/meteor?

func NewGame() *Game {
	game := &Game{}
	game.InputChan = make(chan *Input, 8)
	game.LevelChan = make(chan *Level, 2)

	game.initLevels()

	game.Level = game.Levels[0]
	game.Level.initPlayer()

	return game
}

// TODO: Fix this so that multiple key presses work 100%
func (game *Game) handleInput(input *Input) {
	if input.Pressed {
		switch input.Type {
		case Up:
			if game.Level.Player.Yvel >= 0 {
				game.Level.Player.Yvel = -game.Level.Player.Speed
			}
			game.Level.Player.Yvel -= game.Level.Player.Speed
			break
		case Down:
			if game.Level.Player.Yvel <= 0 {
				game.Level.Player.Yvel = game.Level.Player.Speed
			}
			game.Level.Player.Yvel += game.Level.Player.Speed
			break
		case Left:
			if game.Level.Player.Xvel >= 0 {
				game.Level.Player.Xvel = -game.Level.Player.Speed
			}
			game.Level.Player.Xvel -= game.Level.Player.Speed
			break
		case Right:
			if game.Level.Player.Xvel <= 0 {
				game.Level.Player.Xvel = game.Level.Player.Speed
			}
			game.Level.Player.Xvel += game.Level.Player.Speed
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

func FindNextPointInTravel(dist, rotationRad float64) (float64, float64) {
	nextX := dist * math.Cos(rotationRad)
	nextY := dist * math.Sin(rotationRad)
	return nextX, nextY
}

func DegreeToRad(degree float64) float64 {
	return degree * (math.Pi / 180)
}

func FindDegreeRotation(originY, originX, pointY, pointX int32) float64 {
	return math.Atan2(float64(pointY-originY), float64(pointX-originX)) * (180.0 / math.Pi)
}

func (game *Game) Run() {
	game.LevelChan <- game.Level

	for input := range game.InputChan {
		switch input.Type {
		case None:
			break
		case Quit:
			close(game.LevelChan)
			close(game.InputChan)
			os.Exit(0)
			break
		case LevelComplete:
			// Move to the next level
			currentLevel := game.Level.LevelNumber
			currentPlayer := game.Level.Player
			if !(currentLevel >= len(game.Levels)) {
				game.Level = game.Levels[currentLevel]
				game.Level.Player = currentPlayer
			} else {
				// TODO: Need some end game, or just generate levels and track points?
			}
			break
		default:
			game.handleInput(input)
			game.LevelChan <- game.Level
		}
		//if input.Type == None {
		//	continue
		//}
		//if input.Type == Quit {
		//	close(game.LevelChan)
		//	close(game.InputChan)
		//	os.Exit(0)
		//}
		//if input.Type == LevelComplete {
		//
		//}

	}
}
