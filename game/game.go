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

type Level struct {
	Player                    *Player
	Enemies                   []*Enemy
	Bullets                   []*Bullet
	PrimaryFirePressed        bool
	EnemySpawnTimer           int
	EnemySpawnFrequency       int
	MaxNumberEnemies          int
	EnemyDifficultyMultiplier float32
	PointsToComplete          int
	HasBoss                   bool
	LevelNumber               int
	Complete bool
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
	PointValue int
	Hitpoints                     int
	Strength                      int
	DestroyedAnimationTextureName string
	DestroyedAnimationPlayed      bool
	DestroyedAnimationCounter     int
	DestroyedSoundPlayed          bool
	IsDestroyed                   bool
	FireRateTimer                 int
	FireRateResetValue            int
	IsFiring                      bool
	ShieldHitpoints               int
	EngineFireAnimationCounter    int
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

	game.initLevels()

	game.Level = game.Levels[0]
	game.Level.initPlayer()

	return game
}

func (game *Game) initLevels() {
	// Level 1
	lev1 := &Level{}
	lev1.EnemyDifficultyMultiplier = 0.5
	// Lower this number to increase spawn frequency
	lev1.EnemySpawnFrequency = 250
	lev1.MaxNumberEnemies = 5
	lev1.PointsToComplete = 100
	lev1.LevelNumber = 1
	lev1.EnemySpawnTimer = 0
	lev1.Complete = false
	game.Levels = append(game.Levels, lev1)

	// Level 2
	lev2 := &Level{}
	lev2.EnemyDifficultyMultiplier = 0.7
	lev2.EnemySpawnFrequency = 200
	lev2.MaxNumberEnemies = 10
	lev2.PointsToComplete = 250
	lev2.LevelNumber = 2
	lev2.EnemySpawnTimer = 0
	lev2.Complete = false
	game.Levels = append(game.Levels, lev2)

	// Level 3
	lev3 := &Level{}
	lev3.EnemyDifficultyMultiplier = 0.85
	lev3.EnemySpawnFrequency = 150
	lev3.MaxNumberEnemies = 10
	lev3.PointsToComplete = 500
	lev3.LevelNumber = 3
	lev3.EnemySpawnTimer = 0
	lev3.Complete = false
	game.Levels = append(game.Levels, lev3)
}

// TODO: Fix this so that multiple key presses work 100%
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
		if input.Type == LevelComplete {
			// Move to the next level
			currentLevel := game.Level.LevelNumber
			currentPlayer := game.Level.Player
			if !(currentLevel >= len(game.Levels)) {
				game.Level = game.Levels[currentLevel]
				game.Level.Player = currentPlayer
			} else {
				// TODO: Need some end game, or just generate levels and track points?
			}
		}
		game.handleInput(input)

		game.LevelChan <- game.Level
	}
}
