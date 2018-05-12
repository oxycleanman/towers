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
	Player  *Player
	Enemies []*Enemy
	Bullets []*Bullet
}

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
	Hitpoints                 int
	DestroyedAnimationPlayed  bool
	DestroyedAnimationCounter int
	IsDestroyed               bool
}

type Dimensional interface {
	GetDimensionalData() (int, int, int, int)
}

type Player struct {
	Character
}

type Enemy struct {
	Character
	FireCounter int
}

type Bullet struct {
	Entity
	Velocity
	Damage                 int
	FlashCounter           int
	ExplodeCounter         int
	FireAnimationPlayed    bool
	DestroyAnimationPlayed bool
	IsColliding            bool
}

func (bullet *Bullet) GetDimensionalData() (int, int, int, int) {
	return bullet.X, bullet.Y, bullet.W, bullet.H
}

func (player *Player) GetDimensionalData() (int, int, int, int) {
	return player.X, player.Y, player.W, player.H
}

func (enemy *Enemy) GetDimensionalData() (int, int, int, int) {
	return enemy.X, enemy.Y, enemy.W, enemy.H
}

func getObjMinMax(obj Dimensional) (int, int, int, int) {
	x, y, w, h := obj.GetDimensionalData()
	xMin := x - w/2
	xMax := x + w/2
	yMin := y - h/2
	yMax := y + h/2
	return xMin, yMin, xMax, yMax
}

func CheckCollision(obj1, obj2 Dimensional) bool {
	obj1MinX, obj1MinY, obj1MaxX, obj1MaxY := getObjMinMax(obj1)
	obj2MinX, obj2MinY, obj2MaxX, obj2MaxY := getObjMinMax(obj2)
	if obj2MinX >= obj1MinX && obj2MinX <= obj1MaxX && obj2MinY >= obj1MinY && obj2MinY <= obj1MaxY {
		return true
	}
	if obj2MinX >= obj1MinX && obj2MinX <= obj1MaxX && obj2MaxY >= obj1MinY && obj2MaxY <= obj1MaxY {
		return true
	}
	if obj2MaxX >= obj1MinX && obj2MaxX <= obj1MaxX && obj2MinY >= obj1MinY && obj2MinY <= obj1MaxY {
		return true
	}
	if obj2MaxX >= obj1MinX && obj2MaxX <= obj1MaxX && obj2MaxY >= obj1MinY && obj2MaxY <= obj1MaxY {
		return true
	}
	return false
}

func (bullet *Bullet) Update() {
	if !bullet.IsColliding {
		bulletDirRad := DegreeToRad(bullet.Direction + 90)
		nextX, nextY := findNextPointInTravel(bullet.Speed, bulletDirRad)
		bullet.X += int(nextX)
		bullet.Y += int(nextY)
	}
}

func (player *Player) Move() {
	player.X += player.Xvel
	player.Y += player.Yvel
}

func NewGame() *Game {
	game := &Game{}
	game.InputChan = make(chan *Input, 2)
	game.LevelChan = make(chan *Level, 2)

	game.Level = &Level{}
	return game
}

func findNextPointInTravel(dist, rotationRad float64) (int, int) {
	nextX := dist * math.Cos(rotationRad)
	nextY := dist * math.Sin(rotationRad)
	return int(nextX), int(nextY)
}

func (game *Game) handleInput(input *Input) {
	if input.Pressed {
		switch input.Type {
		case Up:
			if game.Level.Player.Yvel >= -5 {
				game.Level.Player.Yvel -= 5
			}
			//game.Level.Player.Direction = DUp
			break
		case Down:
			if game.Level.Player.Yvel <= 5 {
				game.Level.Player.Yvel += 5
			}
			//game.Level.Player.Direction = DDown
			break
		case Left:
			if game.Level.Player.Xvel >= -5 {
				game.Level.Player.Xvel -= 5
			}
			//game.Level.Player.Direction = DLeft
			break
		case Right:
			if game.Level.Player.Xvel <= 5 {
				game.Level.Player.Xvel += 5
			}
			//game.Level.Player.Direction = DRight
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
		default:
			//fmt.Println("Some input not pressed")
		}
	}
}

//func DetectCollision()

func FindDegreeRotation(originY, originX, pointY, pointX int32) float64 {
	return math.Atan2(float64(pointY-originY), float64(pointX-originX)) * (180 / math.Pi)
}

func DegreeToRad(degree float64) float64 {
	return degree * (math.Pi / 180)
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
