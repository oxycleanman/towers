package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"math/rand"
	"strconv"
	"strings"
	"fmt"
	"time"
)

type uiElement struct {
	game.Pos
	texture *sdl.Texture
}

type uiButton struct {
	uiElement
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
	horzTiles, vertTiles, totalWidth int32
	horzOffset, vertOffset float64
}

type statusBar struct {
	uiElement
	hudElement
	maxTiles int32
}

type menu struct {
	uiElement
}

type hud struct {
	uiElement
	hudElement
	healthBar *statusBar
	shieldBar *statusBar
	insideTexture                                *sdl.Texture
}

type ui struct {
	WinWidth             int32
	WinHeight            int32
	horzTiles, vertTiles int
	backgroundTexture            *sdl.Texture
	cursor                       *cursor
	hud                          *hud
	menu	*menu
	renderer                     *sdl.Renderer
	window                       *sdl.Window
	font                         *ttf.Font
	fontLarge *ttf.Font
	textureMap                   map[string]*sdl.Texture
	keyboardState                []uint8
	inputChan                    chan *game.Input
	levelChan                    chan *game.Level
	currentMouseX                int32
	currentMouseY                int32
	playerInit                   bool
	fontTextureMap               map[string]*sdl.Texture
	largeFontTextureMap               map[string]*sdl.Texture
	soundFileMap                 map[string]*mix.Chunk
	clickableElementMap          map[string]*uiButton
	uiSpeedLines                 []*uiElement
	uiSpeedLineTimer             int
	meteorTextureNames           []string
	enemyTextureNames            []string
	mapMoveDelay                 int
	mapMoveTimer                 int
	muted                        bool
	paused                       bool
	menuOpen bool
	randNumGen                   *rand.Rand
	levelComplete                bool
	levelCompleteMessageTimer    float64
	levelCompleteMessageShowTime int
	AnimationSpeed float64
	gameOver bool
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
	largeFontSize = 60
	texturePath = "gui/assets/images/"
	soundFilePath = "gui/assets/sounds/"
	fontPath = "gui/assets/fonts/kenvector_future.ttf"
	gameTitle = "Some Shitty Space Game"
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
	ui.WinHeight = 1080
	ui.WinWidth = 1920
	var err error
	// SDL set up
	ui.window, err = sdl.CreateWindow(gameTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, ui.WinWidth, ui.WinHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.font, err = ttf.OpenFont(fontPath, fontSize)
	if err != nil {
		panic(err)
	}
	ui.fontLarge, err = ttf.OpenFont(fontPath, largeFontSize)
	if err != nil {
		panic(err)
	}
	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED | sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}
	sdl.ShowCursor(0)
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	// Initialize UI
	ui.randNumGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.horzTiles = 0
	ui.vertTiles = 0
	ui.textureMap = make(map[string]*sdl.Texture)
	ui.fontTextureMap = make(map[string]*sdl.Texture)
	ui.largeFontTextureMap = make(map[string]*sdl.Texture)
	ui.soundFileMap = make(map[string]*mix.Chunk)
	ui.clickableElementMap = make(map[string]*uiButton)
	ui.uiSpeedLineTimer = 0
	ui.playerInit = false
	ui.mapMoveTimer = 0
	ui.mapMoveDelay = 5
	ui.muted = false
	ui.levelCompleteMessageShowTime = 150
	ui.AnimationSpeed = 100

	ui.currentMouseX = int32(ui.WinWidth / 2)
	ui.currentMouseY = int32(ui.WinHeight / 2)

	// Load Assets
	ui.loadTextures(texturePath)
	ui.loadSounds(soundFilePath)

	return ui
}

// TODO: Add draw func and logic to add Power-Ups

func (ui *ui) DrawBackground(level *game.Level) {
	// TODO: Draw better background to create illusion of motion
	//if ui.backgroundTexture == nil {
	//	ui.backgroundTexture = ui.textureMap[backgroundTexture]
	//}
	//if ui.horzTiles == 0 && ui.vertTiles == 0 {
	//	_, _, w, h, err := ui.backgroundTexture.Query()
	//	if err != nil {
	//		panic(err)
	//	}
	//	ui.horzTiles = ui.WinWidth / int(w)
	//	ui.vertTiles = ui.WinHeight / int(h)
	//}
	//destRect := &sdl.Rect{0, 0, int32(ui.WinWidth), int32(ui.WinHeight)}
	//ui.renderer.Copy(ui.backgroundTexture, nil, destRect)
}

func (ui *ui) DrawSpeedLines(deltaTime uint32) {
	// TODO: Maybe the number of lines changes as the player moves towards the top of the screen/accelerates?
	deltaTimeS := float64(deltaTime)/1000
	for _, line := range ui.uiSpeedLines {
		if line.BoundBox.Y > ui.WinHeight {
			spawnX := float64(ui.randNumGen.Intn(int(ui.WinWidth)))
			spawnY := -float64(ui.randNumGen.Intn(int(ui.WinHeight)))
			line.X = spawnX
			line.Y = spawnY
		}
		//This should be in Logic, not GUI (movement)
		line.Y += 5 * ui.AnimationSpeed * deltaTimeS
		line.BoundBox.X = int32(line.X)
		line.BoundBox.Y = int32(line.Y)
		ui.renderer.Copy(line.texture, nil, line.BoundBox)
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
		ui.cursor.BoundBox = &sdl.Rect{}
		ui.cursor.BoundBox.W = w
		ui.cursor.BoundBox.H = h
	}
	ui.cursor.BoundBox.X = ui.currentMouseX
	ui.cursor.BoundBox.Y = ui.currentMouseY
	ui.renderer.Copy(ui.cursor.texture, nil, ui.cursor.BoundBox)
}

func (ui *ui) DrawMenu() {
	if ui.menu == nil {
		ui.menu = &menu{}
		ui.menu.texture = ui.textureMap["glassPanel_cornerBR"]
		_, _, w, h, err := ui.menu.texture.Query()
		if err != nil {
			panic(err)
		}
		xPos := ui.WinWidth/2 - w * 2
		yPos := ui.WinHeight/2 - h * 2
		ui.menu.BoundBox = &sdl.Rect{int32(xPos), int32(yPos), w * 4, h * 4}
	}
	if ui.menuOpen {
		ui.renderer.Copy(ui.menu.texture, nil, ui.menu.BoundBox)
	}
}

func (ui *ui) DrawUiElements(level *game.Level) {
	// Draw Player Hitpoints, eventually the entire HUD
	p := level.Player
	hpTex := ui.stringToNormalFontTexture(strconv.Itoa(int(p.Hitpoints)) + " HP", sdl.Color{255, 255, 255, 1})
	_, _, hpW, hpH, err := hpTex.Query()
	if err != nil {
		panic(err)
	}
	pTex := ui.stringToNormalFontTexture(strconv.Itoa(int(p.Points)), sdl.Color{255, 255, 255, 1})
	_, _, pW, pH, err := pTex.Query()
	if err != nil {
		panic(err)
	}
	levTex := ui.stringToNormalFontTexture("Level " + strconv.Itoa(level.LevelNumber), sdl.Color{255, 255, 255, 1})
	_, _, levW, levH, err := levTex.Query()
	if err != nil {
		panic(err)
	}
	lifeTex := ui.stringToNormalFontTexture("Lives: " + strconv.Itoa(level.Player.Lives), sdl.Color{255, 255, 255, 1})
	_, _, lifeW, lifeH, err := levTex.Query()
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
		ui.hud.horzTiles = int32(ui.WinWidth/2) / w
		ui.hud.vertTiles = 4
		ui.hud.BoundBox = &sdl.Rect{}
		ui.hud.BoundBox.W = w
		ui.hud.BoundBox.H = h
		ui.hud.totalWidth = ui.hud.BoundBox.W * ui.hud.horzTiles
		ui.hud.horzOffset = float64((ui.WinWidth - ui.hud.totalWidth) / 2)
		ui.hud.healthBar = &statusBar{}
		ui.hud.healthBar.BoundBox = &sdl.Rect{}
		ui.hud.shieldBar = &statusBar{}
		ui.hud.shieldBar.BoundBox = &sdl.Rect{}
		ui.hud.healthBar.maxTiles = 20
		ui.hud.healthBar.horzTiles = ui.hud.healthBar.maxTiles
		ui.hud.healthBar.vertTiles = 1
		ui.hud.healthBar.horzOffset = ui.hud.horzOffset + float64(ui.hud.totalWidth - 316)
		ui.hud.healthBar.vertOffset = float64(ui.WinHeight - ui.hud.BoundBox.H + 35)
		ui.hud.shieldBar.maxTiles = 20
		ui.hud.shieldBar.horzTiles = ui.hud.shieldBar.maxTiles
		ui.hud.shieldBar.vertTiles = 1
		ui.hud.shieldBar.horzOffset = ui.hud.horzOffset + float64(ui.hud.totalWidth - 316)
		ui.hud.shieldBar.vertOffset = float64(ui.WinHeight - ui.hud.BoundBox.H + 66)
		ui.hud.healthBar.BoundBox.W = 16
		ui.hud.shieldBar.BoundBox.W = 16
	}

	if len(ui.clickableElementMap) == 0 {
		ui.loadUiElements()
	}

	ui.hud.shieldBar.horzTiles = level.Player.ShieldHitpoints/5
	ui.hud.healthBar.horzTiles = level.Player.Hitpoints/5

	for i := 0; i < int(ui.hud.horzTiles); i++ {
		var tex *sdl.Texture
		ui.hud.BoundBox.X = int32(i) * ui.hud.BoundBox.W + int32(ui.hud.horzOffset)
		ui.hud.BoundBox.Y = ui.WinHeight - ui.hud.BoundBox.H
		if i == 0 {
			tex = ui.textureMap["metalPanel_blueCorner_noBorder"]
			ui.renderer.CopyEx(tex, nil, ui.hud.BoundBox, 0, nil, sdl.FLIP_HORIZONTAL)
		} else if i == int(ui.hud.horzTiles) - 1 {
			tex = ui.textureMap["metalPanel_blueCorner_noBorder"]
			ui.renderer.Copy(tex, nil, ui.hud.BoundBox)
		} else {
			tex = ui.textureMap["metalPanel_blue_noBorder"]
			ui.renderer.Copy(tex, nil, ui.hud.BoundBox)
		}
	}

	// Draw Health Bar Container and Badge
	for i := 0; i < int(ui.hud.healthBar.maxTiles); i++ {
		var tex *sdl.Texture
		ui.hud.healthBar.BoundBox.X = int32(i) * ui.hud.healthBar.BoundBox.W + int32(ui.hud.healthBar.horzOffset)
		ui.hud.healthBar.BoundBox.Y = int32(ui.hud.healthBar.vertOffset)
		if i == 0 {
			tex = ui.textureMap["barHorizontal_shadow_left"]
			ui.hud.healthBar.BoundBox.W = 6
			ui.hud.healthBar.BoundBox.H = 26
			ui.renderer.Copy(tex, nil, ui.hud.healthBar.BoundBox)
		} else if i == int(ui.hud.healthBar.maxTiles) - 1 {
			tex = ui.textureMap["barHorizontal_shadow_right"]
			ui.hud.healthBar.BoundBox.W = 6
			ui.hud.healthBar.BoundBox.H = 26
			ui.hud.healthBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.healthBar.BoundBox)
		} else {
			tex = ui.textureMap["barHorizontal_shadow_mid"]
			ui.hud.healthBar.BoundBox.W = 16
			ui.hud.healthBar.BoundBox.H = 26
			ui.hud.healthBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.healthBar.BoundBox)
		}
		tex = ui.textureMap["pill_green"]
		ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.hud.healthBar.horzOffset - 32), int32(ui.hud.healthBar.vertOffset), 22, 22})
	}


	//Draw Shield Bar Container and Badge
	for i := 0; i < int(ui.hud.shieldBar.maxTiles); i++ {
		var tex *sdl.Texture
		ui.hud.shieldBar.BoundBox.X = int32(i) * ui.hud.shieldBar.BoundBox.W + int32(ui.hud.shieldBar.horzOffset)
		ui.hud.shieldBar.BoundBox.Y = int32(ui.hud.shieldBar.vertOffset)
		if i == 0 {
			tex = ui.textureMap["barHorizontal_shadow_left"]
			ui.hud.shieldBar.BoundBox.W = 6
			ui.hud.shieldBar.BoundBox.H = 26
			ui.renderer.Copy(tex, nil, ui.hud.shieldBar.BoundBox)
		} else if i == int(ui.hud.shieldBar.maxTiles) - 1 {
			tex = ui.textureMap["barHorizontal_shadow_right"]
			ui.hud.shieldBar.BoundBox.W = 6
			ui.hud.shieldBar.BoundBox.H = 26
			ui.hud.shieldBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.shieldBar.BoundBox)
		} else {
			tex = ui.textureMap["barHorizontal_shadow_mid"]
			ui.hud.shieldBar.BoundBox.W = 16
			ui.hud.shieldBar.BoundBox.H = 26
			ui.hud.shieldBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.shieldBar.BoundBox)
		}
		tex = ui.textureMap["shield_gold"]
		ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.hud.shieldBar.horzOffset - 32), int32(ui.hud.shieldBar.vertOffset), 22, 22})
	}

	// Draw Health Bar
	for i := 0; i < int(ui.hud.healthBar.horzTiles); i++ {
		var tex *sdl.Texture
		ui.hud.healthBar.BoundBox.X = int32(i) * ui.hud.healthBar.BoundBox.W + int32(ui.hud.healthBar.horzOffset)
		ui.hud.healthBar.BoundBox.Y = int32(ui.hud.healthBar.vertOffset)
		if i == 0 {
			tex = ui.textureMap["barHorizontal_green_left"]
			ui.hud.healthBar.BoundBox.W = 6
			ui.hud.healthBar.BoundBox.H = 26
			ui.renderer.Copy(tex, nil, ui.hud.healthBar.BoundBox)
		} else if i == int(ui.hud.healthBar.maxTiles) - 1 {
			tex = ui.textureMap["barHorizontal_green_right"]
			ui.hud.healthBar.BoundBox.W = 6
			ui.hud.healthBar.BoundBox.H = 26
			ui.hud.healthBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.healthBar.BoundBox)
		} else {
			tex = ui.textureMap["barHorizontal_green_mid"]
			ui.hud.healthBar.BoundBox.W = 16
			ui.hud.healthBar.BoundBox.H = 26
			ui.hud.healthBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.healthBar.BoundBox)
		}
	}

	//Draw Shield Bar
	for i := 0; i < int(ui.hud.shieldBar.horzTiles); i++ {
		var tex *sdl.Texture
		ui.hud.shieldBar.BoundBox.X = int32(i) * ui.hud.shieldBar.BoundBox.W + int32(ui.hud.shieldBar.horzOffset)
		ui.hud.shieldBar.BoundBox.Y = int32(ui.hud.shieldBar.vertOffset)
		if i == 0 {
			tex = ui.textureMap["barHorizontal_yellow_left"]
			ui.hud.shieldBar.BoundBox.W = 6
			ui.hud.shieldBar.BoundBox.H = 26
			ui.renderer.Copy(tex, nil, ui.hud.shieldBar.BoundBox)
		} else if i == int(ui.hud.shieldBar.maxTiles) - 1 {
			tex = ui.textureMap["barHorizontal_yellow_right"]
			ui.hud.shieldBar.BoundBox.W = 6
			ui.hud.shieldBar.BoundBox.H = 26
			ui.hud.shieldBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.shieldBar.BoundBox)
		} else {
			tex = ui.textureMap["barHorizontal_yellow_mid"]
			ui.hud.shieldBar.BoundBox.W = 16
			ui.hud.shieldBar.BoundBox.H = 26
			ui.hud.shieldBar.BoundBox.X -= 10
			ui.renderer.Copy(tex, nil, ui.hud.shieldBar.BoundBox)
		}
	}

	// Draw Clickable Elements
	for _, element := range ui.clickableElementMap {
		if element.mouseOver && !element.clicked {
			element.texture.SetBlendMode(sdl.BLENDMODE_ADD)
			ui.renderer.Copy(element.texture, nil, element.BoundBox)
		} else if element.clicked {
			element.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
			element.texture.SetColorMod(220, 220, 220)
			ui.renderer.CopyEx(element.texture, nil, element.BoundBox, 0, nil, sdl.FLIP_VERTICAL)
		} else {
			element.texture.SetColorMod(255, 255, 255)
			element.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
			ui.renderer.Copy(element.texture, nil, element.BoundBox)
		}
		ui.renderer.Copy(element.textTexture, nil, element.textBoundBox)
	}

	// Copy HP and Point elements to the renderer
	ui.renderer.Copy(hpTex, nil, &sdl.Rect{0, 0, hpW, hpH})
	ui.renderer.Copy(pTex, nil, &sdl.Rect{hpW + 20, 0, pW, pH})
	ui.renderer.Copy(levTex, nil, &sdl.Rect{int32(ui.WinWidth - 20) - levW, 0, levW, levH})
	ui.renderer.Copy(lifeTex, nil, &sdl.Rect{int32(ui.WinWidth - 40) - levW - lifeW, 0, lifeW, lifeH})
}

func (ui *ui) DrawPlayer(level *game.Level, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	if level.Player.Texture == nil {
		tex := ui.textureMap[level.Player.TextureName]
		level.Player.Texture = tex
		_, _, w, h, err := tex.Query()
		if err != nil {
			fmt.Println("ERROR: Invalid Texture:", level.Player.TextureName)
			panic(err)
		}
		level.Player.BoundBox.W = w/2
		level.Player.BoundBox.H = h/2
		level.Player.X = float64(ui.WinWidth/2 - level.Player.BoundBox.W/2)
		level.Player.Y = float64(ui.WinHeight/2 - level.Player.BoundBox.H/2)
		level.Player.FireOffsetX = 0
		//player.Direction = game.FindDegreeRotation(int32(player.Y+player.H/2), int32(player.X+player.W/2), ui.currentMouseY, ui.currentMouseX) - 90
		// Arbitrary number 5 here to slightly move fire point forward of texture
		level.Player.FireOffsetY = float64(level.Player.BoundBox.H/2 + 5)
	}
	player := level.Player
	player.BoundBox.X = int32(player.X)
	player.BoundBox.Y = int32(player.Y)
	if !player.IsDestroyed {
		if level.Player.IsFiring {
			level.Player.FireRateTimer++
		}
		if player.Xvel > 0 {
			if !strings.Contains(player.TextureName, "left") {
				// Play the turn right animation for the player
				if int(player.AnimationCounter) <= player.TurnAnimationCounter && !player.TurnAnimationPlayed {
					player.TextureName = "turn_right" + strconv.Itoa(int(player.AnimationCounter))
					player.AnimationCounter += ui.AnimationSpeed * deltaTimeS
				} else {
					player.AnimationCounter = 0
					player.TurnAnimationPlayed = true
				}
			}
		} else {
			if !(player.TextureName == "player") && !strings.Contains(player.TextureName, "left") {
				// Player texture is somewhere in the turn right animation, reverse it
				n, err := strconv.Atoi(strings.Replace(player.TextureName, "turn_right", "", 1))
				if err != nil {
					panic(err)
				}
				if player.ReverseAnimationCounter == 0 {
					player.ReverseAnimationCounter = float64(n)
				}
				if int(player.ReverseAnimationCounter) >= 0 {
					player.TextureName = "turn_right" + strconv.Itoa(int(player.ReverseAnimationCounter))
					player.ReverseAnimationCounter -= 100 * deltaTimeS
				} else {
					player.TextureName = "player"
					player.ReverseAnimationCounter = 0
					player.AnimationCounter = 1
					player.TurnAnimationPlayed = false
				}
			}
		}
		if player.Xvel < 0 {
			if !strings.Contains(player.TextureName, "right") {
				// Play the turn right animation for the player
				if int(player.AnimationCounter) <= player.TurnAnimationCounter && !player.TurnAnimationPlayed {
					player.TextureName = "turn_left" + strconv.Itoa(int(player.AnimationCounter))
					player.AnimationCounter += ui.AnimationSpeed * deltaTimeS
				} else {
					player.AnimationCounter = 0
					player.TurnAnimationPlayed = true
				}
			}
		} else {
			if !(player.TextureName == "player") && !strings.Contains(player.TextureName, "right") {
				// Player texture is somewhere in the turn right animation, reverse it
				n, err := strconv.Atoi(strings.Replace(player.TextureName, "turn_left", "", 1))
				if err != nil {
					panic(err)
				}
				if player.ReverseAnimationCounter == 0 {
					player.ReverseAnimationCounter = float64(n)
				}
				if int(player.ReverseAnimationCounter) >= 0 {
					player.TextureName = "turn_left" + strconv.Itoa(int(player.ReverseAnimationCounter))
					player.ReverseAnimationCounter -= 100 * deltaTimeS
				} else {
					player.TextureName = "player"
					player.ReverseAnimationCounter = 0
					player.AnimationCounter = 1
					player.TurnAnimationPlayed = false
				}
			}
		}
		if player.TurnAnimationPlayed && player.TextureName == "player" {
			player.TurnAnimationPlayed = false
		}
		player.Texture = ui.textureMap[player.TextureName]
		tex := player.Texture

		// TODO: Rework shield so it looks correct with new ship texture
		// Draw Player Shield
		//shieldTex := ui.textureMap["shield"]
		//_, _, sw, sh, err := shieldTex.Query()
		//if err != nil {
		//	panic(err)
		//}
		//if player.ShieldHitpoints > 0 {
		//	if player.ShieldHitpoints > 25 && player.ShieldHitpoints < 75 {
		//		shieldTex.SetColorMod(255, 255, 0)
		//	} else if player.ShieldHitpoints <= 25 {
		//		shieldTex.SetColorMod(255, 0, 0)
		//	} else {
		//		shieldTex.SetColorMod(255, 255, 255)
		//	}
		//	// Subtract 10 from Y pos to move shield further forward from player
		//	ui.renderer.Copy(shieldTex, nil, &sdl.Rect{int32(player.X + player.W/2) - sw/2, int32(player.Y + player.H/2) - sh/2 - 10, sw, sh})
		//}
		ui.renderer.Copy(tex, nil, player.BoundBox)

		// TODO: Need to add draw logic here to account for player upgrades (better guns, better shield, etc)

		// Engine Fire Animation
		if player.EngineFireAnimationCounter > 5 {
			playerEngineFireTexture = ui.textureMap["fire0"+strconv.Itoa(ui.randNumGen.Intn(3)+1)]
			player.EngineFireAnimationCounter = 1
		} else {
			if playerEngineFireTexture == nil {
				playerEngineFireTexture = ui.textureMap["fire0"+strconv.Itoa(ui.randNumGen.Intn(3)+1)]
			}
			player.EngineFireAnimationCounter += ui.AnimationSpeed * deltaTimeS
		}
		_, _, w, h, err := playerEngineFireTexture.Query()
		if err != nil {
			panic(err)
		}
		if player.IsAccelerating {
			ui.renderer.Copy(playerEngineFireTexture, nil, &sdl.Rect{player.BoundBox.X + player.BoundBox.W/2 - w/2, player.BoundBox.Y + player.BoundBox.H + 5, w, h})
		}
	}
}

func (ui *ui) DrawEnemies(level *game.Level, deltaTime uint32) {
	// TODO: Have large meteors break up into smaller ones
	deltaTimeS := float64(deltaTime)/1000
	for _, enemy := range level.Enemies {
		if enemy.Texture == nil {
			enemy.Texture = ui.textureMap[enemy.TextureName]
			_, _, w, h, err := enemy.Texture.Query()
			if err != nil {
				panic(err)
			}
			enemy.BoundBox.W = w
			enemy.BoundBox.H = h
			enemy.FireOffsetX = 0
			// Arbitrary number 5 here to slightly move fire point forward of texture
			enemy.FireOffsetY = float64(h/2 + 5)
		}
		enemy.BoundBox.X = int32(enemy.X)
		enemy.BoundBox.Y = int32(enemy.Y)
		if !enemy.IsDestroyed {
			if enemy.ShouldSpin {
				ui.renderer.CopyEx(enemy.Texture, nil, enemy.BoundBox, enemy.SpinAngle, nil, sdl.FLIP_NONE)
				if enemy.SpinTimer >= 3 {
					enemy.SpinAngle += enemy.SpinSpeed
					if enemy.SpinAngle >= 360 {
						enemy.SpinAngle = 0
					}
					enemy.SpinTimer = 0
				} else {
					enemy.SpinTimer += ui.AnimationSpeed * deltaTimeS
				}
			} else {
				ui.renderer.Copy(enemy.Texture, nil, enemy.BoundBox)
			}
		}
	}
}

func (ui *ui) DrawExplosions(level *game.Level, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	index := 0
	for _, enemy := range level.Enemies {
		if enemy.IsDestroyed && !enemy.DestroyedAnimationPlayed {
			if !enemy.DestroyedSoundPlayed {
				explosionSound := "boom" + strconv.Itoa(int(ui.randNumGen.Intn(9)+1))
				enemyDestroyedSound := ui.soundFileMap[explosionSound]
				enemyDestroyedSound.Play(-1, 0)
				enemy.DestroyedSoundPlayed = true
			}
			seconds := int(enemy.DestroyedAnimationCounter)
			frameNumber := seconds % 64
			colNumber := frameNumber % 8
			rowNumber := frameNumber / 8
			if enemy.DestroyedAnimationTextureName == "" {
				enemy.DestroyedAnimationTextureName = "explosion" + strconv.Itoa(int(ui.randNumGen.Intn(4)+1))
			}
			tex := ui.textureMap[enemy.DestroyedAnimationTextureName]
			srcRect := &sdl.Rect{int32(colNumber * 256), int32(rowNumber * 256), 256, 256}
			ui.renderer.Copy(tex, srcRect, &sdl.Rect{enemy.BoundBox.X - enemy.BoundBox.W/2, enemy.BoundBox.Y - enemy.BoundBox.H/2, enemy.BoundBox.W * 2, enemy.BoundBox.H * 2})
			enemy.DestroyedAnimationCounter += ui.AnimationSpeed * deltaTimeS
			if enemy.DestroyedAnimationCounter >= 64 {
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

	if level.Player.IsDestroyed && !level.Player.DestroyedAnimationPlayed {
		if !level.Player.DestroyedSoundPlayed {
			explosionSound := "boom" + strconv.Itoa(int(ui.randNumGen.Intn(9)+1))
			destroyedSound := ui.soundFileMap[explosionSound]
			destroyedSound.Play(-1, 0)
			level.Player.DestroyedSoundPlayed = true
		}
		seconds := int(level.Player.DestroyedAnimationCounter)
		frameNumber := seconds % 64
		colNumber := frameNumber % 8
		rowNumber := frameNumber / 8
		if level.Player.DestroyedAnimationTextureName == "" {
			level.Player.DestroyedAnimationTextureName = "explosion" + strconv.Itoa(int(ui.randNumGen.Intn(4)+1))
		}
		tex := ui.textureMap[level.Player.DestroyedAnimationTextureName]
		srcRect := &sdl.Rect{int32(colNumber * 256), int32(rowNumber * 256), 256, 256}
		ui.renderer.Copy(tex, srcRect, &sdl.Rect{level.Player.BoundBox.X - level.Player.BoundBox.W/2, level.Player.BoundBox.Y - level.Player.BoundBox.H/2, level.Player.BoundBox.W * 2, level.Player.BoundBox.H * 2})
		level.Player.DestroyedAnimationCounter += ui.AnimationSpeed * deltaTimeS
		if level.Player.DestroyedAnimationCounter >= 64 {
			level.Player.DestroyedAnimationPlayed = true
		}
	}
}

func (ui *ui) DrawBullet(level *game.Level, deltaTime uint32) {
	deltaTimeS := float64(deltaTime)/1000
	index := 0
	for i, bullet := range level.Bullets {
		if bullet.Texture == nil {
			tex := ui.textureMap[bullet.TextureName]
			bullet.Texture = tex
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			bullet.BoundBox.W = w
			bullet.BoundBox.H = h
			bullet.X = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2 - bullet.BoundBox.W/2)
			bullet.Y = bullet.FiredBy.Y + float64(bullet.FiredBy.BoundBox.H/2 - bullet.BoundBox.H/2)
			//bullet.BoundBox.X = int32(bullet.X)
			//bullet.BoundBox.Y = int32(bullet.Y)
			bullet.Direction = bullet.FiredBy.Direction
		}
		tex := bullet.Texture
		bullet.BoundBox.X = int32(bullet.X)
		bullet.BoundBox.Y = int32(bullet.Y)
		// Fire Animation
		if bullet.FlashCounter < 5 && !bullet.FireAnimationPlayed {
			fireTex := ui.textureMap[explosionTexture]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			var posX, posY float64
			if bullet.FiredByEnemy {
				posX = bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2 - w/4)
				posY = bullet.FiredBy.Y + float64(bullet.FiredBy.BoundBox.H/2 - h/4) + bullet.FiredBy.FireOffsetY
			} else {
				posX = (bullet.FiredBy.X + float64(bullet.FiredBy.BoundBox.W/2)) - float64(w/4)
				posY = (bullet.FiredBy.Y + float64(bullet.FiredBy.BoundBox.H/2)) - float64(h/4) - bullet.FiredBy.FireOffsetY
			}
			ui.renderer.Copy(fireTex, nil, &sdl.Rect{int32(posX), int32(posY), w / 2, h / 2})
			bullet.FlashCounter += ui.AnimationSpeed * deltaTimeS
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
			if bullet.FiredByEnemy {
				ui.renderer.Copy(fireTex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y) + bullet.BoundBox.H, w / 2, h / 2})
			} else {
				ui.renderer.Copy(fireTex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), w / 2, h / 2})
			}
			bullet.ExplodeCounter += ui.AnimationSpeed * deltaTimeS
		} else {
			if bullet.ExplodeCounter >= 5 && bullet.IsColliding && !bullet.DestroyAnimationPlayed {
				bullet.DestroyAnimationPlayed = true
				bullet.ExplodeCounter = 0
			}
			ui.renderer.Copy(tex, nil, bullet.BoundBox)
		}
		// Keep bullets in the slice that aren't out of bounds (drop the bullets that go off screen so they aren't redrawn)
		if !ui.checkOutOfBounds(bullet.X, bullet.Y, bullet.BoundBox.W, bullet.BoundBox.H) && !bullet.DestroyAnimationPlayed {
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
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			enemy.Hitpoints = 0
			enemy.IsDestroyed = true
		}
	}
	level.Bullets = nil
	level.Player.IsFiring = false
	tex := ui.stringToLargeFontTexture("Level " + strconv.Itoa(level.LevelNumber) + " Complete", sdl.Color{255, 255, 255, 1})
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.WinWidth/2) - w/2, int32(ui.WinHeight/2) - h/2, w, h})
}

func (ui *ui) DrawGameOver(level *game.Level) {
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			enemy.Hitpoints = 0
			enemy.IsDestroyed = true
		}
	}
	level.Bullets = nil
	level.Player.IsFiring = false
	tex := ui.stringToLargeFontTexture("Game Over", sdl.Color{255, 255, 255, 1})
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	ui.renderer.Copy(tex, nil, &sdl.Rect{int32(ui.WinWidth/2) - w/2, int32(ui.WinHeight/2) - h/2, w, h})
}

// Remember to always draw from the ground up
func (ui *ui) Draw(level *game.Level, deltaTime uint32) {
	ui.renderer.Clear()
	ui.DrawBackground(level)
	ui.DrawSpeedLines(deltaTime)
	if level.Complete {
		ui.DrawLevelComplete(level)
	}
	if !ui.gameOver {
		ui.DrawBullet(level, deltaTime)
		ui.DrawPlayer(level, deltaTime)
		ui.DrawEnemies(level, deltaTime)
		ui.DrawExplosions(level, deltaTime)
	} else {
		ui.DrawGameOver(level)
	}
	ui.DrawUiElements(level)
	ui.DrawMenu()
	ui.DrawCursor()
	ui.renderer.Present()
}

func (ui *ui) Run() {
	var deltaTime uint32 = 0
	var frameStart uint32 = 0
	var frameEnd uint32 = 0
	var level *game.Level

	for {
		// Enforce at least a 1ms delay between frames
		if deltaTime < 1 {
			frameStart = sdl.GetTicks()
			sdl.Delay(1)
			frameEnd = sdl.GetTicks()
			deltaTime = frameEnd - frameStart
		} else {
			frameStart = sdl.GetTicks()
		}

		select {
		case newLevel := <-ui.levelChan:
			if level != nil && newLevel.LevelNumber > level.LevelNumber {
				ui.levelCompleteMessageTimer = 0
			}
			level = newLevel
			if level.Complete {
				if int(ui.levelCompleteMessageTimer) >= ui.levelCompleteMessageShowTime {
					ui.inputChan <- &game.Input{Type: game.LevelComplete}
				}
				ui.levelCompleteMessageTimer += ui.AnimationSpeed * (float64(deltaTime)/1000)
			}
			break
		default:
			if level.Complete {
				if int(ui.levelCompleteMessageTimer) >= ui.levelCompleteMessageShowTime {
					ui.inputChan <- &game.Input{Type: game.LevelComplete}
				}
				ui.levelCompleteMessageTimer += ui.AnimationSpeed * (float64(deltaTime)/1000)
			}
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				input := determineInputType(e)
				if !ui.paused && !ui.gameOver {
					ui.inputChan <- input
				}
				break
			case *sdl.MouseButtonEvent:
				input := ui.determineMouseButtonInput(e)
				if input.Type != game.None && !level.Complete {
					ui.inputChan <- input
				}
				break
			case *sdl.MouseMotionEvent:
				ui.checkMouseHover(e)
				ui.currentMouseX = e.X
				ui.currentMouseY = e.Y
				break
			default:
				ui.inputChan <- &game.Input{Type: game.None}
			}
		}

		if !ui.paused {
			ui.Update(level, deltaTime)
		}

		ui.Draw(level, deltaTime)

		frameEnd = sdl.GetTicks()
		deltaTime = frameEnd - frameStart
	}
}
