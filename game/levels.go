package game

type Level struct {
	Player                    *Player
	Enemies                   []*Enemy
	Bullets                   []*Bullet
	PrimaryFirePressed        bool
	EnemySpawnTimer           float64
	EnemySpawnFrequency       float64
	MaxNumberEnemies          int
	EnemyDifficultyMultiplier float64
	PointsToComplete          int32
	HasBoss                   bool
	LevelNumber               int
	Complete                  bool
}

func (game *Game) getNewLevel(oldLevel *Level) *Level {
	newLevel := &Level{}
	if oldLevel == nil {
		newLevel.EnemyDifficultyMultiplier = 0.5
		// Lower this number to increase spawn frequency
		newLevel.EnemySpawnFrequency = 200
		newLevel.MaxNumberEnemies = 15
		newLevel.PointsToComplete = 150
		newLevel.LevelNumber = 1
		return newLevel
	}
	newLevel.EnemyDifficultyMultiplier = oldLevel.EnemyDifficultyMultiplier * 1.25
	newLevel.EnemySpawnFrequency = oldLevel.EnemySpawnFrequency * 0.9
	newLevel.MaxNumberEnemies = int(float64(oldLevel.MaxNumberEnemies) * 1.25)
	newLevel.PointsToComplete = int32(float64(oldLevel.PointsToComplete) * 2)
	newLevel.LevelNumber = oldLevel.LevelNumber + 1
	newLevel.Player = oldLevel.Player
	return newLevel
}
