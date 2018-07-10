package game

type Level struct {
	Player                    *Player
	Enemies                   []*Enemy
	Bullets                   []*Bullet
	PrimaryFirePressed        bool
	EnemySpawnTimer           int
	EnemySpawnFrequency       int
	MaxNumberEnemies          int
	EnemyDifficultyMultiplier float32
	PointsToComplete          int
	HasBoss                   bool
	LevelNumber               int
	Complete                  bool
}

func (game *Game) initLevels() {
	// Level 1
	lev1 := &Level{}
	lev1.EnemyDifficultyMultiplier = 0.5
	// Lower this number to increase spawn frequency
	lev1.EnemySpawnFrequency = 150
	lev1.MaxNumberEnemies = 10
	lev1.PointsToComplete = 100
	lev1.LevelNumber = 1
	lev1.Complete = false
	lev1.PrimaryFirePressed = false
	game.Levels = append(game.Levels, lev1)

	// Level 2
	lev2 := &Level{}
	lev2.EnemyDifficultyMultiplier = 0.7
	lev2.EnemySpawnFrequency = 125
	lev2.MaxNumberEnemies = 15
	lev2.PointsToComplete = 250
	lev2.LevelNumber = 2
	lev2.Complete = false
	lev2.PrimaryFirePressed = false
	game.Levels = append(game.Levels, lev2)

	// Level 3
	lev3 := &Level{}
	lev3.EnemyDifficultyMultiplier = 0.85
	lev3.EnemySpawnFrequency = 110
	lev3.MaxNumberEnemies = 18
	lev3.PointsToComplete = 500
	lev3.LevelNumber = 3
	lev3.Complete = false
	lev3.PrimaryFirePressed = false
	game.Levels = append(game.Levels, lev3)
}
