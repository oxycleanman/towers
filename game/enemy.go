package game

import (
	"strings"
	"github.com/veandco/go-sdl2/sdl"
)

type PowerUpType int
const (
	Health PowerUpType = iota
	Shield
	Laser
	Life
	Empty
)

type Enemy struct {
	Character
	CanFire        bool
	ConstantMotion bool
	SpinTimer      float64
	ShouldSpin     bool
	IsUfo          bool
	IsMeteor       bool
	SpinSpeed      float64
	SpinAngle      float64
	IsBoss         bool
	IsFractured    bool
	IsPowerUp      bool
	PowerUpType    PowerUpType
}

// Implement Shooter Interface
func (enemy *Enemy) GetFireSettings() (float64, float64, bool) {
	return enemy.FireRateTimer, enemy.FireRateResetValue, false
}

func (enemy *Enemy) SetFireTimer(value float64) {
	enemy.FireRateTimer = value
}

func (enemy *Enemy) GetSelf() *Character {
	return &enemy.Character
}

// TODO: Need an enemy factory of some kind here to generate different enemy types
func (level *Level) InitEnemy(initX, initY float64, spawnType int, texName string, isFractured bool, powerUpType PowerUpType) *Enemy {
	enemy := &Enemy{}
	// spawnType 0 = meteor, 1 = enemy, >=2 = power up
	switch spawnType {
	case 0:
		enemy.TextureName = texName
		enemy.Hitpoints = 10
		enemy.PointValue = 5
		enemy.Strength = 5
		if strings.Contains(enemy.TextureName, "large") {
			enemy.Hitpoints = 20
			enemy.PointValue = 15
			enemy.Strength = 15
		}
		enemy.Speed = 150
		enemy.ConstantMotion = true
		enemy.ShouldSpin = true
		enemy.SpinAngle = 0
		enemy.SpinSpeed = 5
		enemy.CanFire = false
		enemy.IsMeteor = true
		if strings.Contains(enemy.TextureName, "small") {
			enemy.IsFractured = true
		} else {
			enemy.IsFractured = isFractured
		}
		break
	case 1:
		enemy.TextureName = texName
		enemy.Hitpoints = int32(50 * level.EnemyDifficultyMultiplier)
		enemy.PointValue = 25
		enemy.Strength = int32(15 * level.EnemyDifficultyMultiplier)
		enemy.Speed = 150 * level.EnemyDifficultyMultiplier
		enemy.ConstantMotion = false
		if strings.Contains(enemy.TextureName, "00") {
			enemy.ShouldSpin = true
			enemy.IsUfo = true
		}
		enemy.SpinAngle = 0
		enemy.SpinSpeed = 5
		enemy.CanFire = true
		enemy.IsBoss = false
		break
	case 2:
		enemy.TextureName = texName
		enemy.PointValue = 50
		enemy.Hitpoints = 1
		enemy.Strength = int32(50 * level.EnemyDifficultyMultiplier)
		enemy.Speed = 180 * level.EnemyDifficultyMultiplier
		enemy.ConstantMotion = true
		enemy.ShouldSpin = false
		enemy.CanFire = false
		enemy.IsMeteor = false
		enemy.IsFractured = false
		enemy.IsPowerUp = true
		enemy.IsBoss = false
		enemy.PowerUpType = powerUpType
		break
	}
	enemy.IsDestroyed = false
	enemy.FireRateTimer = 0
	enemy.FireRateResetValue = 150
	enemy.X = initX
	enemy.Y = initY
	enemy.BoundBox = &sdl.Rect{X:int32(enemy.X), Y:int32(enemy.Y)}
	return enemy
}

func (enemy *Enemy) Update(level *Level) {
	if !enemy.IsDestroyed && enemy.FireRateTimer < enemy.FireRateResetValue && enemy.CanFire {
		enemy.FireRateTimer++
	}
}

// TODO: This logic should be better to make enemies more difficult
func (enemy *Enemy) Move(level *Level, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	if enemy.ConstantMotion {
		enemy.Y += enemy.Speed * deltaTimeS
	} else {
		player := level.Player
		if player.X > enemy.X {
			enemy.X += enemy.Speed * deltaTimeS
		} else {
			enemy.X -= enemy.Speed * deltaTimeS
		}
		// Should they move faster on Y than X?
		enemy.Y += enemy.Speed * deltaTimeS * 1.5
	}
}
