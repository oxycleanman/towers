package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

func (ui *ui) mute() {
	if mix.Volume(-1, -1) > 0 {
		mix.Volume(-1, 0)
		ui.muted = true
		ui.uiElementMap["muteButton"].textTexture = ui.stringToTexture("unmute", fontColor)
	} else {
		mix.Volume(-1, 128)
		ui.muted = false
		ui.uiElementMap["muteButton"].textTexture = ui.stringToTexture("mute", fontColor)
	}
}

func (ui *ui) pause() {
	if !ui.paused {
		ui.paused = true
		ui.uiElementMap["pauseButton"].textTexture = ui.stringToTexture("unpause", fontColor)
	} else {
		ui.paused = false
		ui.uiElementMap["pauseButton"].textTexture = ui.stringToTexture("pause", fontColor)
	}
}

func (ui *ui) SpawnEnemies(level *game.Level) {
	if level.EnemySpawnTimer >= level.EnemySpawnFrequency && len(level.Enemies) < 10 {
		spawnX := ui.randNumGen.Intn(ui.WinWidth)
		level.Enemies = append(level.Enemies, level.InitEnemy(spawnX, -100))
		level.EnemySpawnTimer = 0
	} else {
		level.EnemySpawnTimer++
	}
}

func (ui *ui) UpdateEnemies(level *game.Level) {
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			if enemy.Y > ui.WinHeight {
				//Enemy is outside of the screen bounds, destroy it
				enemy.IsDestroyed = true
				enemy.DestroyedAnimationPlayed = true
			}
			ui.CheckFiring(level, enemy)
			enemy.Move(level)
		}
	}
}

func (ui *ui) CheckFiring(level *game.Level, entity game.Shooter) {
	timer, reset, isPlayer := entity.GetFireSettings()
	if timer >= reset {
		var texName string
		//var laserFireSound *mix.Chunk
		if isPlayer {
			texName = playerLaserTexture
			//laserFireSound = ui.soundFileMap[playerLaserSound]
		} else {
			texName = enemyLaserTexture
			//laserFireSound = ui.soundFileMap[enemyLaserSound]
		}
		bullet := level.InitBullet(texName)
		bullet.FiredByEnemy = !isPlayer
		bullet.FiredBy = entity.GetSelf()
		bullet.Damage = bullet.FiredBy.Strength
		level.Bullets = append(level.Bullets, bullet)
		entity.SetFireTimer(0)
		//laserFireSound.Play(-1, 0)
	} else if !isPlayer {
		entity.SetFireTimer(timer + 1)
	}
}

func (ui *ui) UpdateBullets(level *game.Level) {
	for _, bullet := range level.Bullets {
		bullet.Update()
	}
}

func (ui *ui) UpdatePlayer(level *game.Level) {
	level.Player.Move(0, ui.WinHeight, 0, ui.WinWidth)
}

func (ui *ui) checkOutOfBounds(x, y int, w, h int32) bool {
	if x > ui.WinWidth+int(w) || x < int(0-w) || y > ui.WinHeight+int(h) || y < int(0-h) {
		return true
	}
	return false
}

func (ui *ui) checkCollisions(level *game.Level) {
	// Bullet Collisions
	for _, bullet := range level.Bullets {
		if !bullet.IsColliding {
			bulletRect := &sdl.Rect{int32(bullet.X), int32(bullet.Y), int32(bullet.W), int32(bullet.H)}
			if !bullet.FiredByEnemy {
				for _, enemy := range level.Enemies {
					if !enemy.IsDestroyed {
						enemyRect := &sdl.Rect{int32(enemy.X), int32(enemy.Y), int32(enemy.W), int32(enemy.H)}
						if enemyRect.HasIntersection(bulletRect) {
							bullet.IsColliding = true
							enemy.Hitpoints -= bullet.Damage
							if enemy.Hitpoints <= 0 {
								enemy.IsDestroyed = true
								level.Player.Points += enemy.PointValue
								if level.Player.Points >= level.PointsToComplete {
									level.Complete = true
								}
							}
							//bulletImpactSound := ui.soundFileMap[impactSound]
							//bulletImpactSound.Volume(45)
							//bulletImpactSound.Play(-1, 0)
						}
					}
				}
			} else {
				playerRect := &sdl.Rect{int32(level.Player.X), int32(level.Player.Y), int32(level.Player.W), int32(level.Player.H)}
				if playerRect.HasIntersection(bulletRect) {
					bullet.IsColliding = true
					if level.Player.ShieldHitpoints > 0 {
						level.Player.ShieldHitpoints -= bullet.Damage
					} else {
						level.Player.Hitpoints -= bullet.Damage
					}
					if level.Player.Hitpoints <= 0 {
						level.Player.IsDestroyed = true
					}
				}
			}
		}
	}

	// Enemy Collisions
	//for _, enemy := range level.Enemies {
	//	if !enemy.IsDestroyed {
	//
	//	}
	//}
}

func (ui *ui) Update(level *game.Level) {
	ui.UpdatePlayer(level)
	if !level.Complete {
		ui.UpdateBullets(level)
		ui.UpdateEnemies(level)
		ui.checkCollisions(level)
		ui.SpawnEnemies(level)
		ui.CheckFiring(level, level.Player)
	}
}
