package game

type Bullet struct {
	Entity
	Velocity
	FiredBy                *Character
	FiredByEnemy           bool
	Damage                 int
	FlashCounter           int
	ExplodeCounter         int
	FireAnimationPlayed    bool
	DestroyAnimationPlayed bool
	IsColliding            bool
}

func (level *Level) InitBullet(texName string) *Bullet {
	bullet := &Bullet{}
	bullet.TextureName = texName
	bullet.Speed = 10.0
	bullet.FlashCounter = 0
	bullet.FireAnimationPlayed = false
	bullet.DestroyAnimationPlayed = false
	bullet.Damage = 0
	bullet.IsColliding = false
	return bullet
}

func (bullet *Bullet) Update() {
	if !bullet.IsColliding {
		var bulletDirRad float64
		if bullet.FiredByEnemy {
			bulletDirRad = DegreeToRad(bullet.Direction + 90)
		} else {
			bulletDirRad = DegreeToRad(bullet.Direction - 90)
		}
		nextX, nextY := FindNextPointInTravel(bullet.Speed, bulletDirRad)
		bullet.X += nextX
		bullet.Y += nextY
	}
}
