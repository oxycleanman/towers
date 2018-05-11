package game

type Player struct {
	Character
	Direction Direction
}

func NewPlayer() *Player {
	player := &Player{}
	player.Direction = DDown
	player.Hitpoints = 100
	player.Speed = 1.0
	player.Xvel = 0
	player.Yvel = 0
	player.X = 0
	player.Y = 0
	player.W = 64
	player.H = 64
	return player
}

func (player *Player) Move() {
	player.X += player.Xvel
	player.Y += player.Yvel
}