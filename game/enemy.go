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

func (level *Level) InitEnemy() *Enemy {
	enemy := &Enemy{}
	enemy.TextureName = "ufoGreen"
	enemy.IsDestroyed = false
	enemy.Hitpoints = 50
	enemy.Strength = 5
	enemy.Speed = 1.0
	enemy.FireRateTimer = 0
	enemy.FireRateResetValue = 100
	enemy.X = 300
	enemy.Y = 300
	return enemy
}

func (enemy *Enemy) Update(level *Level) {
	if !enemy.IsDestroyed && enemy.FireRateTimer < enemy.FireRateResetValue {
		enemy.FireRateTimer++
	}
}
