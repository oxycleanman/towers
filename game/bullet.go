package game

import "github.com/veandco/go-sdl2/sdl"

type Bullet struct {
	Entity
	Velocity
	FiredBy                *Character
	FiredByEnemy           bool
	Damage                 int32
	FlashCounter           float64
	ExplodeCounter         float64
	FireAnimationPlayed    bool
	DestroyAnimationPlayed bool
	IsColliding            bool
}

func (level *Level) InitBullet(texName string) *Bullet {
	bullet := &Bullet{}
	bullet.TextureName = texName
	bullet.Speed = 1000
	bullet.FlashCounter = 0
	bullet.FireAnimationPlayed = false
	bullet.DestroyAnimationPlayed = false
	bullet.Damage = 0
	bullet.IsColliding = false
	bullet.BoundBox = &sdl.Rect{}
	return bullet
}

func (bullet *Bullet) Update(deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	if !bullet.IsColliding {
		var bulletDirRad float64
		if bullet.FiredByEnemy {
			bulletDirRad = DegreeToRad(bullet.Direction + 90)
		} else {
			bulletDirRad = DegreeToRad(bullet.Direction - 90)
		}
		nextX, nextY := FindNextPointInTravel(bullet.Speed * deltaTimeS, bulletDirRad)
		bullet.X += nextX
		bullet.Y += nextY
	}
}
