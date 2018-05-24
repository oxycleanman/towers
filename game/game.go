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
	Strength                  int
	DestroyedAnimationPlayed  bool
	DestroyedAnimationCounter int
	IsDestroyed               bool
	FireRateTimer             int
	FireRateResetValue        int
	IsFiring                  bool
}

type Dimensional interface {
	GetDimensionalData() (int, int, int, int)
}

type Shooter interface {
	// Should return FireRateTimer, FireRateResetValue, and whether the entity is the player
	GetFireSettings() (int, int, bool)
	SetFireTimer(int)
	GetSelf() *Character
}

type Player struct {
	Character
}

type Enemy struct {
	Character
}

type Bullet struct {
	Entity
	Velocity
	FiredBy                *Character
	FiredByEnemy           bool
	Damage                 int
	FlashCounter           int
	ExplodeCounter         int
	FireAnimationPlayed    bool
	DestroyAnimationPlayed bool
	IsColliding            bool
}

// Player and Enemy implement Shooter
func (player *Player) GetFireSettings() (int, int, bool) {
	return player.FireRateTimer, player.FireRateResetValue, true
}

func (player *Player) SetFireTimer(value int) {
	player.FireRateTimer = value
}

func (player *Player) GetSelf() *Character {
	return &player.Character
}

func (enemy *Enemy) GetFireSettings() (int, int, bool) {
	return enemy.FireRateTimer, enemy.FireRateResetValue, false
}

func (enemy *Enemy) SetFireTimer(value int) {
	enemy.FireRateTimer = value
}

func (enemy *Enemy) GetSelf() *Character {
	return &enemy.Character
}

// Bullet, Player, and Enemy implement Dimensional
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

func (level *Level) InitBullet() *Bullet {
	bullet := &Bullet{}
	bullet.TextureName = "bulletDark1"
	bullet.Speed = 10.0
	//bullet.Texture = tex
	bullet.FlashCounter = 0
	bullet.FireAnimationPlayed = false
	bullet.DestroyAnimationPlayed = false
	bullet.Damage = 0
	bullet.IsColliding = false
	//bullet.W = int(w)
	//bullet.H = int(h)
	//bullet.Direction = level.Player.Direction
	//bullet.X = (level.Player.X + level.Player.FireOffsetX) - bullet.W/2
	//bullet.Y = (level.Player.Y + level.Player.FireOffsetY) - bullet.H/2
	return bullet
}

func (level *Level) initPlayer() {
	player := &Player{}
	player.TextureName = "tank_huge"
	player.IsDestroyed = false
	player.Hitpoints = 100
	player.Strength = 10
	player.Speed = 1.0
	player.FireRateTimer = 0
	player.FireRateResetValue = 50
	//player.W = int(w)
	//player.H = int(h)
	//player.X = ui.WinWidth/2 - player.W/2
	//player.Y = ui.WinHeight/2 - player.H/2

	//player.Texture = tex
	level.Player = player
}

func (level *Level) InitEnemy() *Enemy {
	enemy := &Enemy{}
	enemy.TextureName = "tank_dark"
	enemy.IsDestroyed = false
	enemy.Hitpoints = 50
	enemy.Strength = 5
	enemy.Speed = 1.0
	enemy.FireRateTimer = 0
	enemy.FireRateResetValue = 100
	//enemy.W = int(w)
	//enemy.H = int(h)
	enemy.X = 300
	enemy.Y = 300
	//enemy.FireOffsetX = enemy.W / 2
	//enemy.FireOffsetY = enemy.H / 2
	//enemy.Texture = tex
	return enemy
}

func (bullet *Bullet) Update() {
	if !bullet.IsColliding {
		bulletDirRad := DegreeToRad(bullet.Direction + 90)
		nextX, nextY := findNextPointInTravel(bullet.Speed, bulletDirRad)
		bullet.X += nextX
		bullet.Y += nextY
	}
}

func (level *Level) CheckBulletCollisions() {
	for _, bullet := range level.Bullets {
		for _, enemy := range level.Enemies {
			if CheckCollision(enemy, bullet) && !bullet.IsColliding && !enemy.IsDestroyed && !bullet.FiredByEnemy {
				bullet.IsColliding = true
				enemy.Hitpoints -= bullet.Damage
				if enemy.Hitpoints <= 0 {
					enemy.IsDestroyed = true
				}
			}
			if CheckCollision(level.Player, bullet) && !bullet.IsColliding && bullet.FiredByEnemy {
				bullet.IsColliding = true
				level.Player.Hitpoints -= bullet.Damage
			}
		}
	}
}

func (enemy *Enemy) Update(level *Level) {
	if !enemy.IsDestroyed && enemy.FireRateTimer < enemy.FireRateResetValue {
		enemy.FireRateTimer++
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
	game.Level.initPlayer()
	game.Level.EnemySpawnTimer = 0

	return game
}

func findNextPointInTravel(dist, rotationRad float64) (int, int) {
	nextX := dist * math.Cos(rotationRad)
	nextY := dist * math.Sin(rotationRad)
	return int(math.Round(nextX)), int(math.Round(nextY))
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
		case FirePrimary:
			game.Level.Player.IsFiring = true
			game.Level.Player.FireRateTimer = game.Level.Player.FireRateResetValue
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
		default:
			//fmt.Println("Some input not pressed")
		}
	}
}

func FindDegreeRotation(originY, originX, pointY, pointX int32) float64 {
	return math.Atan2(float64(pointY-originY), float64(pointX-originX)) * (180.0 / math.Pi)
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
