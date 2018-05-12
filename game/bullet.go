package game

type Bullet struct {
	Entity
	Velocity
	IsNew bool
}

func NewBullet(textureName string) *Bullet {
	bullet := &Bullet{}
	bullet.TextureName = textureName
	bullet.Direction = DDown
	bullet.X = 0
	bullet.Y = 0
	bullet.Xvel = 0
	bullet.Yvel = 0
	bullet.W = 0
	bullet.H = 0
	return bullet
}

func (bullet *Bullet) Update() {
	bullet.X += bullet.Xvel
	bullet.Y += bullet.Yvel
}