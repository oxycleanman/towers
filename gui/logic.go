package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/mix"
)

func (ui *ui) mute() {
	if mix.Volume(-1, -1) > 0 {
		mix.Volume(-1, 0)
		ui.muted = true
		ui.clickableElementMap["muteButton"].textTexture = ui.stringToTexture("unmute", fontColor)
	} else {
		mix.Volume(-1, 128)
		ui.muted = false
		ui.clickableElementMap["muteButton"].textTexture = ui.stringToTexture("mute", fontColor)
	}
}

func (ui *ui) pause() {
	if !ui.paused {
		ui.paused = true
		ui.clickableElementMap["pauseButton"].textTexture = ui.stringToTexture("unpause", fontColor)
	} else {
		ui.paused = false
		ui.clickableElementMap["pauseButton"].textTexture = ui.stringToTexture("pause", fontColor)
	}
}

func (ui *ui) openCloseMenu() {
	if ui.menuOpen {
		ui.menuOpen = false
		ui.paused = false
	} else {
		ui.menuOpen = true
		ui.paused = true
	}
}

func (ui *ui) SpawnEnemies(level *game.Level) {
	if level.EnemySpawnTimer >= level.EnemySpawnFrequency && len(level.Enemies) < level.MaxNumberEnemies {
		spawnX := float64(ui.randNumGen.Intn(int(ui.WinWidth)))
		enemyOrMeteor := ui.randNumGen.Intn(2)
		var texName string
		if enemyOrMeteor == 1 {
			texName = "ufo3d"
			//texName = ui.enemyTextureNames[ui.randNumGen.Intn(len(ui.enemyTextureNames))]
		} else {
			texName = ui.meteorTextureNames[ui.randNumGen.Intn(len(ui.meteorTextureNames))]
		}
		level.Enemies = append(level.Enemies, level.InitEnemy(spawnX, -100, enemyOrMeteor, texName))
		level.EnemySpawnTimer = 0
	} else {
		level.EnemySpawnTimer++
	}
}

func (ui *ui) UpdateEnemies(level *game.Level, deltaTime uint32) {
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			if int32(enemy.Y) > ui.WinHeight {
				//Enemy is outside of the screen bounds, destroy it
				enemy.IsDestroyed = true
				enemy.DestroyedAnimationPlayed = true
			}
			if enemy.CanFire {
				ui.CheckFiring(level, enemy, deltaTime)
			}
			enemy.Move(level, deltaTime)
		}
	}
}

// TODO: Make it so that clicking fires as slow as holding down the mouse button (Fire Speed will be an upgrade)
func (ui *ui) CheckFiring(level *game.Level, entity game.Shooter, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	timer, reset, isPlayer := entity.GetFireSettings()
	if timer >= reset {
		var texName string
		var laserFireSound *mix.Chunk
		if isPlayer {
			texName = playerLaserTexture
			laserFireSound = ui.soundFileMap[playerLaserSound]
		} else {
			texName = enemyLaserTexture
			laserFireSound = ui.soundFileMap[enemyLaserSound]
		}
		bullet := level.InitBullet(texName)
		bullet.FiredByEnemy = !isPlayer
		bullet.FiredBy = entity.GetSelf()
		bullet.Damage = bullet.FiredBy.Strength
		level.Bullets = append(level.Bullets, bullet)
		entity.SetFireTimer(0)
		laserFireSound.Play(-1, 0)
	} else if !isPlayer {
		entity.SetFireTimer(timer + ui.AnimationSpeed * deltaTimeS)
	}
}

func (ui *ui) UpdateBullets(level *game.Level, deltaTime uint32) {
	for _, bullet := range level.Bullets {
		bullet.Update(deltaTime)
	}
}

func (ui *ui) UpdatePlayer(level *game.Level, deltaTime uint32) {
	level.Player.Move(0, ui.WinHeight, 0, ui.WinWidth, deltaTime)
}

func (ui *ui) checkOutOfBounds(x, y float64, w, h int32) bool {
	if int32(x) > ui.WinWidth + w || int32(x) < 0 - w || int32(y) > ui.WinHeight + h || int32(y) < 0 - h {
		return true
	}
	return false
}

func (ui *ui) checkCollisions(level *game.Level) {
	// Bullet Collisions
	for _, bullet := range level.Bullets {
		if !bullet.IsColliding {
			if !bullet.FiredByEnemy {
				for _, enemy := range level.Enemies {
					if !enemy.IsDestroyed {
						if enemy.BoundBox.HasIntersection(bullet.BoundBox) {
							bullet.IsColliding = true
							enemy.Hitpoints -= bullet.Damage
							if enemy.Hitpoints <= 0 {
								enemy.IsDestroyed = true
								level.Player.Points += enemy.PointValue
								if level.Player.Points >= level.PointsToComplete {
									level.Complete = true
								}
							}
							bulletImpactSound := ui.soundFileMap[impactSound]
							bulletImpactSound.Volume(45)
							bulletImpactSound.Play(-1, 0)
						}
					}
				}
			} else {
				if level.Player.BoundBox.HasIntersection(bullet.BoundBox) {
					bullet.IsColliding = true
					if level.Player.ShieldHitpoints > 0 {
						level.Player.ShieldHitpoints -= bullet.Damage
						ui.hud.shieldBar.horzTiles = level.Player.ShieldHitpoints/5
					} else {
						level.Player.Hitpoints -= bullet.Damage
						ui.hud.healthBar.horzTiles = level.Player.Hitpoints/5
					}
					if level.Player.Hitpoints <= 0 {
						level.Player.IsDestroyed = true
					}
				}
			}
		}
	}

	// Enemy Collisions With Player
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			if enemy.BoundBox.HasIntersection(level.Player.BoundBox) {
				if !enemy.IsBoss {
					enemy.Hitpoints = 0
					enemy.IsDestroyed = true
					if level.Player.ShieldHitpoints > 0 {
						level.Player.ShieldHitpoints -= enemy.Strength * 2
						ui.hud.shieldBar.horzTiles = level.Player.ShieldHitpoints / 5
					} else {
						level.Player.Hitpoints -= enemy.Strength * 2
						ui.hud.healthBar.horzTiles = level.Player.Hitpoints / 5
					}
					if level.Player.Hitpoints <= 0 {
						level.Player.IsDestroyed = true
					}
				} else {
					level.Player.Hitpoints = 0
					level.Player.IsDestroyed = true
				}
			}
		}
	}

	// Player collisions with power-ups
}

func (ui *ui) Update(level *game.Level, deltaTime uint32) {
	ui.UpdatePlayer(level, deltaTime)
	if !level.Complete {
		ui.UpdateBullets(level, deltaTime)
		ui.UpdateEnemies(level, deltaTime)
		ui.checkCollisions(level)
		ui.SpawnEnemies(level)
		ui.CheckFiring(level, level.Player, deltaTime)
	}
}
