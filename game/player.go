package game

import "github.com/veandco/go-sdl2/sdl"

type Player struct {
	Character
	Currency                                         int
	Lives                                            int
	Points                                           int32
	AtTop, AtBottom, AtLeft, AtRight, IsAccelerating bool
	AnimationCounter                                 float64
	ReverseAnimationCounter	float64
	TurnAnimationCounter                              int
	TurnAnimationPlayed                              bool
	LaserLevel int32
}

// Implement Shooter Interface
func (player *Player) GetFireSettings() (float64, float64, bool) {
	return player.FireRateTimer, player.FireRateResetValue, true
}

func (player *Player) SetFireTimer(value float64) {
	player.FireRateTimer = value
}

func (player *Player) GetSelf() *Character {
	return &player.Character
}

func (level *Level) InitPlayer(isNewPlayer bool) {
	if isNewPlayer {
		player := &Player{}
		player.TextureName = "player"
		player.IsDestroyed = false
		player.Hitpoints = 100
		player.ShieldHitpoints = 100
		player.Points = 0
		player.Speed = 800.0
		player.FireRateTimer = 0
		player.FireRateResetValue = 10
		player.AtBottom = false
		player.AtLeft = false
		player.AtRight = false
		player.AtTop = false
		player.EngineFireAnimationCounter = 1
		player.TurnAnimationCounter = 20
		player.IsAccelerating = false
		player.Lives = 3
		player.LaserLevel = 1
		player.Strength = 10
		player.BoundBox = &sdl.Rect{}
		level.Player = player
	} else {
		level.Player.TextureName = "player"
		level.Player.Hitpoints = 100
		level.Player.ShieldHitpoints = 100
		level.Player.AtBottom = false
		level.Player.AtLeft = false
		level.Player.AtRight = false
		level.Player.AtTop = false
		level.Player.IsAccelerating = false
		level.Player.DestroyedAnimationPlayed = false
		level.Player.DestroyedAnimationPlayed = false
		level.Player.DestroyedAnimationTextureName = ""
		level.Player.DestroyedAnimationCounter = 0
		level.Player.IsDestroyed = false
		level.Player.Lives--
	}
}

func (player *Player) Move(topBound, bottomBound, leftBound, rightBound int32, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	if !player.IsDestroyed {
		newX := int32(player.X + player.Xvel*deltaTimeS + float64(player.BoundBox.W/2))
		newY := int32(player.Y + player.Yvel*deltaTimeS + float64(player.BoundBox.H/2))
		if player.Xvel != 0 && newX <= rightBound && newX >= leftBound {
			player.X += player.Xvel * deltaTimeS
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
			player.Y += player.Yvel * deltaTimeS
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
}
