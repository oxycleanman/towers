package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"math/rand"
	"strconv"
	"time"
	"fmt"
)

type uiElement struct {
	game.Pos
	game.Size
	texture *sdl.Texture
}

type uiButton struct {
	uiElement
	boundBox     *sdl.Rect
	textBoundBox *sdl.Rect
	mouseOver    bool
	clicked      bool
	onClick      func()
	textTexture  *sdl.Texture
}

type cursor struct {
	uiElement
}

type hudElement struct {
	horzTiles, vertTiles, totalWidth, horzOffset, vertOffset int
}

type statusBar struct {
	uiElement
	hudElement
	maxTiles int
}

type hud struct {
	uiElement
	hudElement
	healthBar *statusBar
	shieldBar *statusBar
	insideTexture                                *sdl.Texture
}

type ui struct {
	WinWidth             int
	WinHeight            int
	horzTiles, vertTiles int
	backgroundTexture    *sdl.Texture
	cursor               *cursor
	hud                  *hud
	renderer             *sdl.Renderer
	window               *sdl.Window
	font                 *ttf.Font
	textureMap           map[string]*sdl.Texture
	keyboardState        []uint8
	inputChan            chan *game.Input
	levelChan            chan *game.Level
	currentMouseX        int32
	currentMouseY        int32
	playerInit           bool
	fontTextureMap       map[string]*sdl.Texture
	soundFileMap         map[string]*mix.Chunk
	uiElementMap         map[string]*uiButton
	uiSpeedLines	[]*uiElement
	uiSpeedLineTimer int
	meteorTextureNames []string
	enemyTextureNames []string
	mapMoveDelay         int
	mapMoveTimer         int
	muted                bool
	paused               bool
	randNumGen           *rand.Rand
	levelComplete bool
	levelCompleteMessageTimer int
	levelCompleteMessageShowTime int
}

const (
	playerLaserTexture = "laserBlue01"
	enemyLaserTexture  = "laserGreen02"
	explosionTexture   = "laserBlue10"
	playerLaserSound   = "sfx_laser1"
	enemyLaserSound    = "sfx_laser2"
	impactSound        = "boom7"
	cursorTexture      = "cursor_pointer3D"
	hudTexture         = "metalPanel"
	innerHudTexture    = "metalPanel_plate"
	backgroundTexture  = "purple"
	fontSize           = 24
)

var fontColor = sdl.Color{0, 0, 0, 1}
var playerEngineFireTexture *sdl.Texture

func init() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	if err := ttf.Init(); err != nil {
		panic(err)
	}
	if err := mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		panic(err)
	}
	if err := mix.Init(mix.INIT_OGG); err != nil {
		panic(err)
	}
	mix.AllocateChannels(16)
}

func NewUi(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.randNumGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.WinHeight = 1080
	ui.WinWidth = 1920
	ui.horzTiles = 0
	ui.vertTiles = 0
	ui.textureMap = make(map[string]*sdl.Texture)
	ui.fontTextureMap = make(map[string]*sdl.Texture)
	ui.soundFileMap = make(map[string]*mix.Chunk)
	ui.uiElementMap = make(map[string]*uiButton)
	ui.uiSpeedLineTimer = 0
	ui.playerInit = false
	ui.mapMoveTimer = 0
	ui.mapMoveDelay = 5
	ui.muted = false
	ui.levelCompleteMessageShowTime = 150

	var err error
	ui.window, err = sdl.CreateWindow("Some Shitty Space Game", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, int32(ui.WinWidth), int32(ui.WinHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.font, err = ttf.OpenFont("gui/assets/fonts/kenvector_future.ttf", fontSize)
	if err != nil {
		panic(err)
	}

	ui.currentMouseX = int32(ui.WinWidth / 2)
	ui.currentMouseY = int32(ui.WinHeight / 2)

	renderer, err := sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	ui.renderer = renderer
	sdl.ShowCursor(0)
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.loadTextures("gui/assets/images/")
	ui.loadSounds("gui/assets/sounds/")
	ui.loadUiElements()
	return ui
}

// TODO: Add draw func and logic to add Power-Ups

// TODO: Add draw func and logic to add non-enemy objects like meteors

func (ui *ui) DrawBackground(level *game.Level) {
	// TODO: Draw better background to create illusion of motion
	if ui.backgroundTexture == nil {
		ui.backgroundTexture = ui.textureMap[backgroundTexture]
	}
	if ui.horzTiles == 0 && ui.vertTiles == 0 {
		_, _, w, h, err := ui.backgroundTexture.Query()
		if err != nil {
			panic(err)
		}
		ui.horzTiles = ui.WinWidth / int(w)
		ui.vertTiles = ui.WinHeight / int(h)
	}
	//destRect := &sdl.Rect{0, 0, int32(ui.WinWidth), int32(ui.WinHeight)}
	//ui.renderer.Copy(ui.backgroundTexture, nil, destRect)
}

func (ui *ui) DrawSpeedLines() {
	// TODO: Maybe the number of lines changes as the player moves towards the top of the screen/accelerates?
	for _, line := range ui.uiSpeedLines {
		if line.Y > ui.WinHeight {
			spawnX := ui.randNumGen.Intn(ui.WinWidth)
			spawnY := -ui.randNumGen.Intn(ui.WinHeight)
			line.X = spawnX
			line.Y = spawnY
		}
		line.Y += 5
		ui.renderer.Copy(line.texture, nil, &sdl.Rect{int32(line.X), int32(line.Y), int32(line.W), int32(line.H)})
	}
}

func (ui *ui) DrawCursor() {
	if ui.cursor == nil {
		ui.cursor = &cursor{}
		ui.cursor.texture = ui.textureMap[cursorTexture]
		_, _, w, h, err := ui.cursor.texture.Query()
		if err != nil {
			panic(err)
		}
		ui.cursor.W = int(w)
		ui.cursor.H = int(h)
	}
	ui.renderer.Copy(ui.cursor.texture, nil, &sdl.Rect{ui.currentMouseX, ui.currentMouseY, int32(ui.cursor.W), int32(ui.cursor.H)})
}

func (ui *ui) DrawUiElements(level *game.Level) {
	// Draw Player Hitpoints, eventually the entire HUD
	p := level.Player
	hpTex := ui.stringToTexture(strconv.Itoa(p.Hitpoints)+" HP", sdl.Color{255, 255, 255, 1})
	_, _, hpW, hpH, err := hpTex.Query()
	if err != nil {
		panic(err)
	}
	pTex := ui.stringToTexture(strconv.Itoa(p.Points), sdl.Color{255, 255, 255, 1})
	_, _, pW, pH, err := pTex.Query()
	if err != nil {
		panic(err)
	}
	levTex := ui.stringToTexture("Level " + strconv.Itoa(level.LevelNumber), sdl.Color{255, 255, 255, 1})
	_, _, levW, levH, err := levTex.Query()
	if err != nil {
		panic(err)
	}

	// Initialize and Draw HUD
	if ui.hud == nil {
		ui.hud = &hud{}
		ui.hud.texture = ui.textureMap[hudTexture]
		ui.hud.insideTexture = ui.textureMap[innerHudTexture]
		_, _, w, h, err := ui.hud.texture.Query()
		if err != nil {
			panic(err)
		}
		ui.hud.horzTiles = (ui.WinWidth/2) / int(w)
		ui.hud.vertTiles = 4
		ui.hud.W = int(w)
		ui.hud.H = int(h)
		ui.hud.totalWidth = ui.hud.W * ui.hud.horzTiles
		ui.hud.horzOffset = (ui.WinWidth - ui.hud.totalWidth) / 2
		ui.hud.healthBar = &statusBar{}
		ui.hud.shieldBar = &statusBar{}
		ui.hud.healthBar.maxTiles = 20
		ui.hud.healthBar.horzTiles = ui.hud.healthBar.maxTiles
		ui.hud.healthBar.vertTiles = 1
		ui.hud.healthBar.horzOffset = ui.hud.horzOffset + ui.hud.totalWidth - 316
		ui.hud.healthBar.vertOffset = ui.WinHeight - ui.hud.H + 35
		ui.hud.shieldBar.maxTiles = 20
		ui.hud.shieldBar.horzTiles = ui.hud.shieldBar.maxTiles
		ui.hud.shieldBar.vertTiles = 1
		ui.hud.shieldBar.horzOffset = ui.hud.horzOffset + ui.hud.totalWidth - 316
		ui.hud.shieldBar.vertOffset = ui.WinHeight - ui.hud.H + 66
		ui.hud.healthBar.W = 16
		ui.hud.shieldBar.W = 16
	}
	for i := 0; i < ui.hud.horzTiles; i++ {
		var tex *sdl.Texture
		if i == 0 {
			tex = ui.textureMap["metalPanel_blueCorner_noBorder"]
			ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(i*ui.hud.W + ui.hud.horzOffset), int32(ui.WinHeight - ui.hud.H), int32(ui.hud.W), int32(ui.hud.H)}, 0, nil, sdl.FLIP_HORIZONTAL)
		} else if i == ui.hud.horzTiles-1 {
			tex = ui.textureMap["metalPanel_blueCorner_noBorder"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i*ui.hud.W + ui.hud.horzOffset), int32(ui.WinHeight - ui.hud.H), int32(ui.hud.W), int32(ui.hud.H)})
		} else {
			tex = ui.textureMap["metalPanel_blue_noBorder"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i*ui.hud.W + ui.hud.horzOffset), int32(ui.WinHeight - ui.hud.H), int32(ui.hud.W), int32(ui.hud.H)})
		}
	}

	// Draw Health Bar Container and Badge
	for i := 0; i < ui.hud.healthBar.maxTiles; i++ {
		var tex *sdl.Texture
		if i == 0 {
			tex = ui.textureMap["barHorizontal_shadow_left"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i*ui.hud.healthBar.W + ui.hud.healthBar.horzOffset), int32(ui.hud.healthBar.vertOffset), 6, 26})
		} else if i == ui.hud.healthBar.maxTiles - 1 {
			tex = ui.textureMap["barHorizontal_shadow_right"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i*ui.hud.healthBar.W + ui.hud.healthBar.horzOffset - 10), int32(ui.hud.healthBar.vertOffset), 6, 26})
		} else {
			tex = ui.textureMap["barHorizontal_shadow_mid"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i*ui.hud.healthBar.W + ui.hud.healthBar.horzOffset - 10), int32(ui.hud.healthBar.vertOffset), 16, 26})
		}
		tex = ui.textureMap["pill_green"]
		ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.hud.healthBar.horzOffset - 32), int32(ui.hud.healthBar.vertOffset), 22, 22})
	}


	//Draw Shield Bar Container and Badge
	for i := 0; i < ui.hud.shieldBar.maxTiles; i++ {
		var tex *sdl.Texture
		if i == 0 {
			tex = ui.textureMap["barHorizontal_shadow_left"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.shieldBar.W + ui.hud.shieldBar.horzOffset), int32(ui.hud.shieldBar.vertOffset), 6, 26})
		} else if i == ui.hud.shieldBar.maxTiles - 1 {
			tex = ui.textureMap["barHorizontal_shadow_right"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.shieldBar.W + ui.hud.shieldBar.horzOffset - 10), int32(ui.hud.shieldBar.vertOffset), 6, 26})
		} else {
			tex = ui.textureMap["barHorizontal_shadow_mid"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.shieldBar.W + ui.hud.shieldBar.horzOffset - 10), int32(ui.hud.shieldBar.vertOffset), 16, 26})
		}
		tex = ui.textureMap["shield_gold"]
		ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.hud.shieldBar.horzOffset - 32), int32(ui.hud.shieldBar.vertOffset), 22, 22})
	}

	// Draw Health Bar
	for i := 0; i < ui.hud.healthBar.horzTiles; i++ {
		var tex *sdl.Texture
		if i == 0 {
			tex = ui.textureMap["barHorizontal_green_left"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.healthBar.W + ui.hud.healthBar.horzOffset), int32(ui.hud.healthBar.vertOffset), 6, 26})
		} else if i == ui.hud.healthBar.horzTiles - 1 {
			tex = ui.textureMap["barHorizontal_green_right"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.healthBar.W + ui.hud.healthBar.horzOffset - 10), int32(ui.hud.healthBar.vertOffset), 6, 26})
		} else {
			tex = ui.textureMap["barHorizontal_green_mid"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.healthBar.W + ui.hud.healthBar.horzOffset - 10), int32(ui.hud.healthBar.vertOffset), 16, 26})
		}
	}

	//Draw Shield Bar
	for i := 0; i < ui.hud.shieldBar.horzTiles; i++ {
		var tex *sdl.Texture
		if i == 0 {
			tex = ui.textureMap["barHorizontal_yellow_left"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.shieldBar.W + ui.hud.shieldBar.horzOffset), int32(ui.hud.shieldBar.vertOffset), 6, 26})
		} else if i == ui.hud.shieldBar.horzTiles - 1 {
			tex = ui.textureMap["barHorizontal_yellow_right"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.shieldBar.W + ui.hud.shieldBar.horzOffset - 10), int32(ui.hud.shieldBar.vertOffset), 6, 26})
		} else {
			tex = ui.textureMap["barHorizontal_yellow_mid"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.shieldBar.W + ui.hud.shieldBar.horzOffset - 10), int32(ui.hud.shieldBar.vertOffset), 16, 26})
		}
	}

	// Draw Other UI elements
	for _, element := range ui.uiElementMap {
		if element.mouseOver && !element.clicked {
			element.texture.SetBlendMode(sdl.BLENDMODE_ADD)
			ui.renderer.Copy(element.texture, nil, element.boundBox)
		} else if element.clicked {
			element.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
			element.texture.SetColorMod(220, 220, 220)
			ui.renderer.CopyEx(element.texture, nil, element.boundBox, 0, nil, sdl.FLIP_VERTICAL)
		} else {
			element.texture.SetColorMod(255, 255, 255)
			element.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
			ui.renderer.Copy(element.texture, nil, element.boundBox)
		}
		ui.renderer.Copy(element.textTexture, nil, element.textBoundBox)
	}

	// Copy HP and Point elements to the renderer
	ui.renderer.Copy(hpTex, nil, &sdl.Rect{0, 0, hpW, hpH})
	ui.renderer.Copy(pTex, nil, &sdl.Rect{hpW + 20, 0, pW, pH})
	ui.renderer.Copy(levTex, nil, &sdl.Rect{int32(ui.WinWidth - 20) - levW, 0, levW, levH})
}

func (ui *ui) DrawPlayer(level *game.Level) {
	if level.Player.IsFiring {
		level.Player.FireRateTimer++
	}
	if level.Player.Texture == nil {
		tex := ui.textureMap[level.Player.TextureName]
		level.Player.Texture = tex
		_, _, w, h, err := tex.Query()
		if err != nil {
			panic(err)
		}
		level.Player.W = int(w)
		level.Player.H = int(h)
		level.Player.X = ui.WinWidth/2 - level.Player.W/2
		level.Player.Y = ui.WinHeight/2 - level.Player.H/2
		level.Player.FireOffsetX = 0
		//player.Direction = game.FindDegreeRotation(int32(player.Y+player.H/2), int32(player.X+player.W/2), ui.currentMouseY, ui.currentMouseX) - 90
		// Arbitrary number 5 here to slightly move fire point forward of texture
		level.Player.FireOffsetY = int(h/2) + 5
	}
	player := level.Player
	if player.IsDestroyed {
		// TODO: Lose Scenario... Some kind of modal? Loss of life?
	}
	tex := player.Texture

	// Draw Player Shield
	shieldTex := ui.textureMap["shield"]
	_, _, sw, sh, err := shieldTex.Query()
	if err != nil {
		panic(err)
	}
	if player.ShieldHitpoints > 0 {
		if player.ShieldHitpoints > 25 && player.ShieldHitpoints < 75 {
			shieldTex.SetColorMod(255, 255, 0)
		} else if player.ShieldHitpoints <= 25 {
			shieldTex.SetColorMod(255, 0, 0)
		} else {
			shieldTex.SetColorMod(255, 255, 255)
		}
		// Subtract 10 from Y pos to move shield further forward from player
		ui.renderer.Copy(shieldTex, nil, &sdl.Rect{int32(player.X + player.W/2) - sw/2, int32(player.Y + player.H/2) - sh/2 - 10, sw, sh})
	}
	ui.renderer.Copy(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), int32(player.W), int32(player.H)})

	// TODO: Need to add draw logic here to account for player upgrades (better guns, better shield, etc)

	// Engine Fire Animation
	if player.EngineFireAnimationCounter > 5 {
		playerEngineFireTexture = ui.textureMap["fire0"+strconv.Itoa(ui.randNumGen.Intn(3) + 1)]
		player.EngineFireAnimationCounter = 1
	} else {
		if playerEngineFireTexture == nil {
			playerEngineFireTexture = ui.textureMap["fire0"+strconv.Itoa(ui.randNumGen.Intn(3) + 1)]
		}
		player.EngineFireAnimationCounter++
	}
	_, _, w, h, err := playerEngineFireTexture.Query()
	if err != nil {
		panic(err)
	}
	if player.IsAccelerating {
		ui.renderer.Copy(playerEngineFireTexture, nil, &sdl.Rect{int32(player.X+player.W/2) - w/2, int32(player.Y + player.H + 5), w, h})
	}
}

func (ui *ui) DrawEnemies(level *game.Level) {
	for _, enemy := range level.Enemies {
		if enemy.Texture == nil {
			enemy.Texture = ui.textureMap[enemy.TextureName]
			_, _, w, h, err := enemy.Texture.Query()
			if err != nil {
				panic(err)
			}
			enemy.W = int(w)
			enemy.H = int(h)
			enemy.FireOffsetX = 0
			// Arbitrary number 5 here to slightly move fire point forward of texture
			enemy.FireOffsetY = int(h/2) + 5
		}
		if !enemy.IsDestroyed {
			if enemy.ShouldSpin {
				ui.renderer.CopyEx(enemy.Texture, nil, &sdl.Rect{int32(enemy.X), int32(enemy.Y), int32(enemy.W), int32(enemy.H)}, enemy.SpinAngle * enemy.SpinSpeed, nil, sdl.FLIP_NONE)
				enemy.SpinAngle++
				if enemy.SpinAngle > 360 {
					enemy.SpinAngle = 0
				}
			} else {
				ui.renderer.Copy(enemy.Texture, nil, &sdl.Rect{int32(enemy.X), int32(enemy.Y), int32(enemy.W), int32(enemy.H)})
			}
		}
	}
}

func (ui *ui) DrawExplosions(level *game.Level) {
	index := 0
	for _, enemy := range level.Enemies {
		if enemy.IsDestroyed && !enemy.DestroyedAnimationPlayed {
			if !enemy.DestroyedSoundPlayed {
				explosionSound := "boom" + strconv.Itoa(int(ui.randNumGen.Intn(9)+1))
				enemyDestroyedSound := ui.soundFileMap[explosionSound]
				enemyDestroyedSound.Play(-1, 0)
				enemy.DestroyedSoundPlayed = true
			}
			seconds := enemy.DestroyedAnimationCounter
			frameNumber := seconds % 64
			colNumber := frameNumber % 8
			rowNumber := frameNumber / 8
			if enemy.DestroyedAnimationTextureName == "" {
				enemy.DestroyedAnimationTextureName = "explosion" + strconv.Itoa(int(ui.randNumGen.Intn(4)+1))
			}
			tex := ui.textureMap[enemy.DestroyedAnimationTextureName]
			srcRect := &sdl.Rect{int32(colNumber * 256), int32(rowNumber * 256), 256, 256}
			ui.renderer.Copy(tex, srcRect, &sdl.Rect{int32(enemy.X - enemy.W/2), int32(enemy.Y - enemy.H/2), int32(enemy.W * 2), int32(enemy.H * 2)})
			enemy.DestroyedAnimationCounter++
			if enemy.DestroyedAnimationCounter == 64 {
				enemy.DestroyedAnimationPlayed = true
			}
			// WHY DOES THIS NOT WORK THE SAME AS THE BULLETS???!!!
			if !enemy.DestroyedAnimationPlayed {
				level.Enemies[index] = enemy
				index++
			}
		} else {
			if !enemy.IsDestroyed && !enemy.DestroyedAnimationPlayed {
				level.Enemies[index] = enemy
				index++
			} else {
				enemy = nil
			}
		}
	}
	level.Enemies = append(level.Enemies[:index])
}

func (ui *ui) DrawBullet(level *game.Level) {
	index := 0
	for i, bullet := range level.Bullets {
		if bullet.Texture == nil {
			tex := ui.textureMap[bullet.TextureName]
			bullet.Texture = tex
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			bullet.W = int(w)
			bullet.H = int(h)
			bullet.Direction = bullet.FiredBy.Direction
			bullet.X = (bullet.FiredBy.X + bullet.FiredBy.W/2) - bullet.W/2
			bullet.Y = (bullet.FiredBy.Y + bullet.FiredBy.H/2) - bullet.H/2
		}
		tex := bullet.Texture

		// Fire Animation
		if bullet.FlashCounter < 5 && !bullet.FireAnimationPlayed {
			fireTex := ui.textureMap[explosionTexture]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			var posX, posY int
			if bullet.FiredByEnemy {
				posX = (bullet.FiredBy.X + bullet.FiredBy.W/2) - int(w/4)
				posY = (bullet.FiredBy.Y + bullet.FiredBy.H/2) - int(h/4) + bullet.FiredBy.FireOffsetY
			} else {
				posX = (bullet.FiredBy.X + bullet.FiredBy.W/2) - int(w/4)
				posY = (bullet.FiredBy.Y + bullet.FiredBy.H/2) - int(h/4) - bullet.FiredBy.FireOffsetY
			}
			ui.renderer.Copy(fireTex, nil, &sdl.Rect{int32(posX), int32(posY), w / 2, h / 2})
			bullet.FlashCounter++
		}
		if bullet.FlashCounter >= 5 && !bullet.FireAnimationPlayed {
			bullet.FlashCounter = 0
			bullet.FireAnimationPlayed = true
		}

		// Collision Animation && Normal Travel
		if bullet.ExplodeCounter < 5 && bullet.IsColliding && !bullet.DestroyAnimationPlayed {
			fireTex := ui.textureMap[explosionTexture]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			bullet.Xvel = 0
			bullet.Yvel = 0
			ui.renderer.Copy(fireTex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), w / 2, h / 2})
			bullet.ExplodeCounter++
		} else {
			if bullet.ExplodeCounter >= 5 && bullet.IsColliding && !bullet.DestroyAnimationPlayed {
				bullet.DestroyAnimationPlayed = true
				bullet.ExplodeCounter = 0
			}
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), int32(bullet.W), int32(bullet.H)})
		}
		// Keep bullets in the slice that aren't out of bounds (drop the bullets that go off screen so they aren't redrawn)
		if !ui.checkOutOfBounds(bullet.X, bullet.Y, int32(bullet.W), int32(bullet.H)) && !bullet.DestroyAnimationPlayed {
			if index != i {
				level.Bullets[index] = bullet
			}
			index++
		} else {
			bullet = nil
		}
	}
	level.Bullets = level.Bullets[:index]
}

func (ui *ui) DrawLevelComplete(level *game.Level) {
	fmt.Println(ui.levelCompleteMessageTimer)
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			enemy.Hitpoints = 0
			enemy.IsDestroyed = true
		}
	}
	level.Bullets = nil
	level.Player.IsFiring = false
	tex := ui.stringToTexture("Level " + strconv.Itoa(level.LevelNumber) + " Complete", sdl.Color{255, 255, 255, 1})
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.WinWidth/2) - w/2, int32(ui.WinHeight/2) - h/2, w, h})
}

// Remember to always draw from the ground up
func (ui *ui) Draw(level *game.Level) {
	ui.renderer.Clear()
	if !ui.paused {
		ui.Update(level)
	}
	ui.DrawBackground(level)
	ui.DrawSpeedLines()
	ui.DrawBullet(level)
	ui.DrawPlayer(level)
	ui.DrawEnemies(level)
	ui.DrawExplosions(level)
	ui.DrawUiElements(level)
	if level.Complete {
		ui.DrawLevelComplete(level)
	}
	ui.DrawCursor()
	ui.renderer.Present()
}

func (ui *ui) Run() {
	// Aiming for 120 FPS
	var targetFrameTime = 1.0 / 120.0 * 1000
	var frameStart time.Time
	var elapsedTime float64
	var level *game.Level

	for {
		frameStart = time.Now()

		select {
		case newLevel := <-ui.levelChan:
			if level != nil && newLevel.LevelNumber > level.LevelNumber {
				ui.levelCompleteMessageTimer = 0
			}
			level = newLevel
			if level.Complete {
				if ui.levelCompleteMessageTimer >= ui.levelCompleteMessageShowTime {
					ui.inputChan <- &game.Input{Type: game.LevelComplete}
				}
				ui.levelCompleteMessageTimer++
			}
			ui.Draw(level)
			break
		default:
			if level.Complete {
				if ui.levelCompleteMessageTimer >= ui.levelCompleteMessageShowTime {
					ui.inputChan <- &game.Input{Type: game.LevelComplete}
				}
				ui.levelCompleteMessageTimer++
			}
			ui.Draw(level)
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				input := determineInputType(e)
				if !ui.paused {
					ui.inputChan <- input
				}
			case *sdl.MouseButtonEvent:
				input := ui.determineMouseButtonInput(e)
				if input.Type != game.None && !level.Complete {
					ui.inputChan <- input
				}
			case *sdl.MouseMotionEvent:
				ui.checkMouseHover(e)
				ui.currentMouseX = e.X
				ui.currentMouseY = e.Y
			default:
				ui.inputChan <- &game.Input{Type: game.None}
			}
		}

		elapsedTime = time.Since(frameStart).Seconds() * 1000
		if elapsedTime < targetFrameTime {
			sdl.Delay(uint32(targetFrameTime - elapsedTime))
			elapsedTime = time.Since(frameStart).Seconds()
		}

		//m := &runtime.MemStats{}
		//runtime.ReadMemStats(m)
		//fmt.Println(m.HeapObjects, m.HeapInuse, m.HeapReleased, m.HeapSys)
	}
}
