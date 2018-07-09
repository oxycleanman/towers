package game

type Player struct {
	Character
	Currency                                         int
	Lives int
	Points	int
	AtTop, AtBottom, AtLeft, AtRight, IsAccelerating bool
}

// Implement Shooter Interface
func (player *Player) GetFireSettings() (int, int, bool) {
	return player.FireRateTimer, player.FireRateResetValue, true
}

func (player *Player) SetFireTimer(value int) {
	player.FireRateTimer = value
}

func (player *Player) GetSelf() *Character {
	return &player.Character
}

func (level *Level) initPlayer() {
	player := &Player{}
	player.TextureName = "playerShip1_blue"
	player.IsDestroyed = false
	player.Hitpoints = 100
	player.ShieldHitpoints = 100
	player.Points = 0
	player.Strength = 10
	player.Speed = 1.0
	player.FireRateTimer = 0
	player.FireRateResetValue = 50
	player.AtBottom = false
	player.AtLeft = false
	player.AtRight = false
	player.AtTop = false
	player.EngineFireAnimationCounter = 1
	player.IsAccelerating = false
	level.Player = player
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
		if player.Yvel < 0 {
			player.IsAccelerating = true
		} else {
			player.IsAccelerating = false
		}
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
