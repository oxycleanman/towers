package game

type Enemy struct {
	Character
}

type EnemyType int
const (
	Basic EnemyType = iota
	Medium
	//Large
	//XLarge
)

func NewEnemy(etype EnemyType, textureName string) *Enemy {
	enemy := &Enemy{}
	enemy.TextureName = textureName
	switch etype {
	case Basic:
		enemy.Hitpoints = 10
		enemy.Speed = 0.5
		enemy.W = 32
		enemy.H = 32
	case Medium:
		enemy.Hitpoints = 20
		enemy.Speed = 0.8
		enemy.W = 32
		enemy.H = 32
	}

	return enemy
}
