package game

import "strings"

type Enemy struct {
	Character
	CanFire bool
	ConstantMotion bool
	ShouldSpin bool
	SpinSpeed float64
	SpinAngle float64
	IsBoss bool
}

// Implement Shooter Interface
func (enemy *Enemy) GetFireSettings() (int, int, bool) {
	return enemy.FireRateTimer, enemy.FireRateResetValue, false
}

func (enemy *Enemy) SetFireTimer(value int) {
	enemy.FireRateTimer = value
}

func (enemy *Enemy) GetSelf() *Character {
	return &enemy.Character
}

// TODO: Need an enemy factory of some kind here to generate different enemy types
func (level *Level) InitEnemy(initX, initY int, enemyOrMeteor int, texName string) *Enemy {
	enemy := &Enemy{}
	// 0 = meteor, 1 = enemy
	if enemyOrMeteor == 1 {
		enemy.TextureName = texName
		enemy.Hitpoints = int(50 * level.EnemyDifficultyMultiplier)
		enemy.PointValue = 25
		enemy.Strength = int(10 * level.EnemyDifficultyMultiplier)
		enemy.Speed = float64(2.0 * level.EnemyDifficultyMultiplier)
		enemy.ConstantMotion = false
		if strings.Contains(enemy.TextureName, "ufo") {
			enemy.ShouldSpin = true
		}
		enemy.SpinAngle = 0
		enemy.SpinSpeed = 3.0
		enemy.CanFire = true
		enemy.IsBoss = false
	} else {
		enemy.TextureName = texName
		enemy.Hitpoints = 10
		enemy.PointValue = 5
		enemy.Strength = 5
		if strings.Contains(enemy.TextureName, "big") {
			enemy.Hitpoints = 20
			enemy.PointValue = 15
			enemy.Strength = 15
		}
		enemy.Speed = 2.0
		enemy.ConstantMotion = true
		enemy.ShouldSpin = true
		enemy.SpinAngle = 0
		enemy.SpinSpeed = 0.8
		enemy.CanFire = false
	}
	enemy.IsDestroyed = false
	enemy.FireRateTimer = 0
	enemy.FireRateResetValue = 150
	enemy.X = initX
	enemy.Y = initY
	return enemy
}

func (enemy *Enemy) Update(level *Level) {
	if !enemy.IsDestroyed && enemy.FireRateTimer < enemy.FireRateResetValue && enemy.CanFire {
		enemy.FireRateTimer++
	}
}

// TODO: This logic should be better to make enemies more difficult
func (enemy *Enemy) Move(level *Level) {
	if enemy.ConstantMotion {
		enemy.Y += int(enemy.Speed)
	} else {
		player := level.Player
		if player.X > enemy.X {
			enemy.X += int(enemy.Speed)
		} else {
			enemy.X -= int(enemy.Speed)
		}
		enemy.Y += int(enemy.Speed * 1.5)
	}
}
