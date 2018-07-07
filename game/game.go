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
	W, H, LeftBound, RightBound, TopBound, BottomBound int
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

type Player struct {
	Character
	Currency                         int
	AtTop, AtBottom, AtLeft, AtRight bool
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

func (level *Level) InitBullet(texName string) *Bullet {
	bullet := &Bullet{}
	bullet.TextureName = texName
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
	player.TextureName = "playerShip1_blue"
	player.IsDestroyed = false
	player.Hitpoints = 100
	player.Strength = 10
	player.Speed = 1.0
	player.FireRateTimer = 0
	player.FireRateResetValue = 50
	player.AtBottom = false
	player.AtLeft = false
	player.AtRight = false
	player.AtTop = false
	//player.W = int(w)
	//player.H = int(h)
	//player.X = ui.WinWidth/2 - player.W/2
	//player.Y = ui.WinHeight/2 - player.H/2

	//player.Texture = tex
	level.Player = player
}

func (level *Level) InitEnemy() *Enemy {
	enemy := &Enemy{}
	enemy.TextureName = "ufoGreen"
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
		var bulletDirRad float64
		if bullet.FiredByEnemy {
			bulletDirRad = DegreeToRad(bullet.Direction + 90)
		} else {
			bulletDirRad = DegreeToRad(bullet.Direction - 90)
		}
		nextX, nextY := findNextPointInTravel(bullet.Speed, bulletDirRad)
		bullet.X += nextX
		bullet.Y += nextY
	}
}

func (enemy *Enemy) Update(level *Level) {
	if !enemy.IsDestroyed && enemy.FireRateTimer < enemy.FireRateResetValue {
		enemy.FireRateTimer++
	}
}

func (player *Player) Move(topBound, bottomBound, leftBound, rightBound int) {
	newX := player.X + player.W/2 + player.Xvel
	newY := player.Y + player.H/2 + player.Yvel
	if player.Xvel != 0 && newX <= rightBound && newX >= leftBound {
		player.X += player.Xvel
		player.AtRight = false
		player.AtLeft = false
	} else {
		if newX >= rightBound {
			player.AtRight = true
		}
		if newX <= leftBound {
			player.AtLeft = true
		}
	}
	if player.Yvel != 0 && newY < bottomBound && newY > topBound {
		player.Y += player.Yvel
		player.AtBottom = false
		player.AtTop = false
	} else {
		if newY >= bottomBound {
			player.AtBottom = true
		}
		if newY <= topBound {
			player.AtTop = true
		}
	}
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
