package game

import (
	"math"
)

type Bullet struct {
	Entity
	Velocity
	IsNew bool
}

func NewBullet(textureName string) *Bullet {
	bullet := &Bullet{}
	bullet.TextureName = textureName
	bullet.Direction = 0.0
	bullet.X = 0
	bullet.Y = 0
	bullet.Xvel = 0
	bullet.Yvel = 0
	bullet.W = 0
	bullet.H = 0
	return bullet
}

func (bullet *Bullet) Update() {
	bulletDirRad := (bullet.Direction + 90) * (math.Pi / 180)
	nextX := 10 * math.Cos(bulletDirRad)
	nextY := 10 * math.Sin(bulletDirRad)
	bullet.X += int(nextX)
	bullet.Y += int(nextY)
}
