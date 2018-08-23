package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/mix"
	"strconv"
	"github.com/veandco/go-sdl2/sdl"
)

func (ui *ui) mute() {
	muteButton := ui.clickableElementMap["muteButton"]
	if mix.Volume(-1, -1) > 0 {
		mix.Volume(-1, 0)
		ui.muted = true
		muteButton.textTexture = ui.stringToNormalFontTexture("unmute", fontColor)
		_, _, w, h, err := muteButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := muteButton.BoundBox.X + muteButton.BoundBox.W/2 - w/2
		textY := muteButton.BoundBox.Y + muteButton.BoundBox.H/2 - h/2
		muteButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), w, h}
	} else {
		mix.Volume(-1, 128)
		ui.muted = false
		muteButton.textTexture = ui.stringToNormalFontTexture("mute", fontColor)
		_, _, w, h, err := muteButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := muteButton.BoundBox.X + muteButton.BoundBox.W/2 - w/2
		textY := muteButton.BoundBox.Y + muteButton.BoundBox.H/2 - h/2
		muteButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), w, h}
	}
}

func (ui *ui) pause() {
	if !ui.paused {
		ui.paused = true
		//ui.clickableElementMap["pauseButton"].textTexture = ui.stringToNormalFontTexture("unpause", fontColor)
	} else {
		ui.paused = false
		//ui.clickableElementMap["pauseButton"].textTexture = ui.stringToNormalFontTexture("pause", fontColor)
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

func (ui *ui) startGame() {
	ui.gameStarted = !ui.gameStarted
	backgroundMusic := ui.musicFileMap["track1"]
	// TODO: Music volume should be adjustable through the menu
	mix.VolumeMusic(50)
	backgroundMusic.Play(100)
}

func (ui *ui) SpawnEnemies(level *game.Level, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	if level.Player.Points < level.PointsToComplete {
		if level.EnemySpawnTimer >= level.EnemySpawnFrequency && len(level.Enemies) < level.MaxNumberEnemies {
			spawnX := float64(ui.randNumGen.Intn(int(ui.WinWidth)))
			enemyOrMeteor := ui.randNumGen.Intn(2)
			var texName string
			if enemyOrMeteor == 1 {
				texName = "0000"
				//texName = ui.enemyTextureNames[ui.randNumGen.Intn(len(ui.enemyTextureNames))]
			} else {
				texName = ui.meteorTextureNames[ui.randNumGen.Intn(len(ui.meteorTextureNames))]
				//texName = "meteor_large1"
			}
			level.Enemies = append(level.Enemies, level.InitEnemy(spawnX, -100, enemyOrMeteor, texName, false, game.Empty))
			level.EnemySpawnTimer = 0
		} else {
			level.EnemySpawnTimer += ui.AnimationSpeed * deltaTimeS
		}
	} else if level.HasBoss && !level.BossSpawned {
		spawnX := float64(ui.randNumGen.Intn(int(ui.WinWidth)))
		level.Enemies = append(level.Enemies, level.InitEnemy(spawnX, -100, 3, "enemyBlue5", false, game.Empty))
		level.BossSpawned = true
	}
}

func (ui *ui) SpawnPowerUps(level *game.Level, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	// TODO: Should the max number of enemies affect power ups?
	if level.PowerUpSpawnTimer >= level.PowerUpSpawnFrequency && len(level.Enemies) < level.MaxNumberEnemies {
		spawnX := float64(ui.randNumGen.Intn(int(ui.WinWidth)))
		randPowerUp := game.PowerUpType(ui.randNumGen.Intn(4))
		//randPowerUp := game.PowerUpType(2)
		var texName string
		switch randPowerUp {
		case game.Health:
			texName = "pill_green"
			break
		case game.Life:
			texName = "powerupBlue"
			break
		case game.Laser:
			texName = "powerupRed"
			break
		case game.Shield:
			texName = "shield_gold"
			break
		}
		level.Enemies = append(level.Enemies, level.InitEnemy(spawnX, -100, 2, texName, false, randPowerUp))
		level.PowerUpSpawnTimer = 0
	} else {
		level.PowerUpSpawnTimer += ui.AnimationSpeed * deltaTimeS
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

func (ui *ui) CheckFiring(level *game.Level, entity game.Shooter, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	timer, reset, isPlayer := entity.GetFireSettings()
	if timer >= reset {
		var texName string
		var laserFireSound *mix.Chunk
		if isPlayer {
			texName = playerLaserTexture
			laserFireSound = ui.soundFileMap[playerLaserSound]
			for i := 0; i < int(level.Player.LaserLevel); i++ {
				if i == 3 {
					break
				}
				bullet := level.InitBullet(texName)
				tex := ui.textureMap[bullet.TextureName]
				bullet.Texture = tex
				_, _, w, h, err := tex.Query()
				if err != nil {
					panic(err)
				}
				bullet.BoundBox.W = w
				bullet.BoundBox.H = h
				bullet.FiredBy = entity.GetSelf()
				bullet.Direction = bullet.FiredBy.Direction
				// Correctly draw single, double, and triple lasers
				// Offsets used here are to line up lasers with upgraded guns (+0.53, +0.49, etc.)
				if level.Player.LaserLevel == 1 {
					bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2)
				} else if level.Player.LaserLevel == 2 {
					if i == 0 {
						bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2) * (float64(i) + 0.53)
					} else {
						bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2) * (float64(i) + 0.49)
					}
				} else {
					if i == 0 {
						bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2)
					} else if i == 1 {
						bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2)*(float64(i) + 0.49)
					} else {
						bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2)*(float64(i) - 1.475)
					}
				}
				bullet.Y = bullet.FiredBy.Y + float64(bullet.FiredBy.BoundBox.H/2-bullet.BoundBox.H/2)
				bullet.FiredByEnemy = !isPlayer
				bullet.Damage = bullet.FiredBy.Strength
				level.Bullets = append(level.Bullets, bullet)
			}
		} else {
			texName = enemyLaserTexture
			laserFireSound = ui.soundFileMap[enemyLaserSound]
			bullet := level.InitBullet(texName)
			tex := ui.textureMap[bullet.TextureName]
			bullet.Texture = tex
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			bullet.BoundBox.W = w
			bullet.BoundBox.H = h
			bullet.FiredBy = entity.GetSelf()
			bullet.Direction = bullet.FiredBy.Direction
			bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2-bullet.BoundBox.W/2)
			bullet.Y = bullet.FiredBy.Y + float64(bullet.FiredBy.BoundBox.H/2-bullet.BoundBox.H/2)
			bullet.FiredByEnemy = !isPlayer
			bullet.Damage = bullet.FiredBy.Strength
			level.Bullets = append(level.Bullets, bullet)
		}
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

func (ui *ui) fractureMeteor(enemy *game.Enemy, level *game.Level) {
	for i := 0; i < 3; i++ {
		x := enemy.BoundBox.X + int32(i) * int32(ui.randNumGen.Intn(40 - -40)) + -40
		y := enemy.BoundBox.Y + int32(1) * int32(ui.randNumGen.Intn(40 - -40)) + -40
		texNum := ui.randNumGen.Intn(48) + 1
		texName := "meteor_small" + strconv.Itoa(texNum)
		smallMeteor := level.InitEnemy(float64(x), float64(y), 0, texName, true, game.Empty)
		smallMeteor.DestroyedAnimationCounter = float64(texNum)
		level.Enemies = append(level.Enemies, smallMeteor)
	}
}

func (ui *ui) checkCollisions(level *game.Level) {
	// Bullet Collisions
	for _, bullet := range level.Bullets {
		if !bullet.IsColliding {
			if !bullet.FiredByEnemy { // Player Bullets hitting enemies
				for _, enemy := range level.Enemies {
					if !enemy.IsDestroyed && !enemy.IsPowerUp {
						if enemy.BoundBox.HasIntersection(bullet.BoundBox) {
							bullet.IsColliding = true
							enemy.Hitpoints -= bullet.Damage
							if enemy.Hitpoints <= 0 {
								enemy.IsDestroyed = true
								level.Player.Points += enemy.PointValue
								if level.Player.Points >= level.PointsToComplete && !level.HasBoss {
									level.Complete = true
								}
								if enemy.IsMeteor && !enemy.IsFractured {
									ui.fractureMeteor(enemy, level)
								}
								if enemy.IsBoss {
									level.Complete = true
								}
							}
							bulletImpactSound := ui.soundFileMap[impactSound]
							bulletImpactSound.Volume(45)
							bulletImpactSound.Play(-1, 0)
						}
					}
				}
			} else { // Enemy bullets hitting player
				if level.Player.BoundBox.HasIntersection(bullet.BoundBox) && !level.Player.IsDestroyed {
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

	// Enemy Collisions With Player
	if !level.Player.IsDestroyed {
		for _, enemy := range level.Enemies {
			if !enemy.IsDestroyed {
				if enemy.BoundBox.HasIntersection(level.Player.BoundBox) {
					if !enemy.IsBoss && !enemy.IsPowerUp { // Normal enemy, not boss or power up
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
					} else if !enemy.IsBoss && enemy.IsPowerUp { // Power Up
						enemy.Hitpoints = 0
						enemy.IsDestroyed = true
						//enemy.DestroyedAnimationPlayed = true
						level.Player.Points += enemy.PointValue
						switch enemy.PowerUpType {
						case game.Health:
							level.Player.Hitpoints += enemy.Strength
							if level.Player.Hitpoints > 100 {
								level.Player.Hitpoints = 100
							}
							break
						case game.Life:
							level.Player.Lives++
							break
						case game.Laser:
							level.Player.LaserLevel++
							level.Player.Strength += level.Player.LaserLevel
							if level.Player.TextureName == "player" {
								level.Player.TextureName = "player_guns"
							}
							weaponUpgradeSound := ui.soundFileMap["weapload"]
							weaponUpgradeSound.Play(-1, 0)
							break
						case game.Shield:
							level.Player.ShieldHitpoints += enemy.Strength
							if level.Player.ShieldHitpoints > 100 {
								level.Player.ShieldHitpoints = 100
							}
							break
						}
					} else { // Enemy is a boss, kill player
						level.Player.Hitpoints = 0
						level.Player.IsDestroyed = true
					}
				}
			}
		}
	}
}

func (ui *ui) checkPlayerDeath(level *game.Level) {
	if level.Player.IsDestroyed && level.Player.DestroyedAnimationPlayed && level.Player.Lives > 0{
		level.InitPlayer(false)
		level.Player.X = float64(ui.WinWidth/2 - level.Player.BoundBox.W/2)
		level.Player.Y = float64(ui.WinHeight/2 - level.Player.BoundBox.H/2)
	}
	if level.Player.Lives == 0 {
		ui.gameOver = true
	}
}

func (ui *ui) Update(level *game.Level, deltaTime uint32) {
	if ! ui.gameOver {
		ui.UpdatePlayer(level, deltaTime)
		ui.checkPlayerDeath(level)
		if !level.Complete && !level.Player.IsDestroyed {
			ui.UpdateBullets(level, deltaTime)
			ui.UpdateEnemies(level, deltaTime)
			ui.checkCollisions(level)
			ui.SpawnEnemies(level, deltaTime)
			ui.SpawnPowerUps(level, deltaTime)
			ui.CheckFiring(level, level.Player, deltaTime)
		}
	}
}
