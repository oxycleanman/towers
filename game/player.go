package game

type Player struct {
	Character
	Direction float64
}

func NewPlayer(textureName string) *Player {
	player := &Player{}
	player.Direction = 0.0
	player.Hitpoints = 100
	player.Speed = 1.0
	player.Xvel = 0
	player.Yvel = 0
	player.X = 0
	player.Y = 0
	player.W = 0
	player.H = 0
	player.FireOffsetX = -50
	player.FireOffsetY = -50
	player.TextureName = textureName
	return player
}

func (player *Player) Move() {
	player.X += player.Xvel
	player.Y += player.Yvel
}
