package game

type Enemy struct {
	Character
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

func (level *Level) InitEnemy(initX, initY int) *Enemy {
	enemy := &Enemy{}
	enemy.TextureName = "ufoGreen"
	enemy.IsDestroyed = false
	enemy.Hitpoints = int(50 * level.EnemyDifficultyMultiplier)
	enemy.PointValue = 25
	enemy.Strength = int(10 * level.EnemyDifficultyMultiplier)
	enemy.Speed = float64(2.0 * level.EnemyDifficultyMultiplier)
	enemy.FireRateTimer = 0
	enemy.FireRateResetValue = 150
	enemy.X = initX
	enemy.Y = initY
	return enemy
}

func (enemy *Enemy) Update(level *Level) {
	if !enemy.IsDestroyed && enemy.FireRateTimer < enemy.FireRateResetValue {
		enemy.FireRateTimer++
	}
}

// TODO: This logic should be better to make enemies more difficult
func (enemy *Enemy) Move(level *Level) {
	player := level.Player
	if player.X > enemy.X {
		enemy.X += int(enemy.Speed)
	} else {
		enemy.X -= int(enemy.Speed)
	}
	enemy.Y += int(enemy.Speed * 1.5)
}
