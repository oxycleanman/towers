package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"strconv"
	"time"
	"math/rand"
	"github.com/veandco/go-sdl2/mix"
)

type uiElement struct {
	game.Pos
	game.Size
	texture *sdl.Texture
}

type uiButton struct {
	uiElement
	boundBox *sdl.Rect
	textBoundBox *sdl.Rect
	mouseOver bool
	clicked bool
	onClick func()
	textTexture *sdl.Texture
}

type cursor struct {
	uiElement
}

type hud struct {
	uiElement
	insideTexture *sdl.Texture
	horzTiles, vertTiles int
}

type ui struct {
	WinWidth       int
	WinHeight      int
	horzTiles, vertTiles int
	backgroundTexture *sdl.Texture
	cursor *cursor
	hud *hud
	renderer       *sdl.Renderer
	window         *sdl.Window
	font           *ttf.Font
	textureMap     map[string]*sdl.Texture
	keyboardState  []uint8
	inputChan      chan *game.Input
	levelChan      chan *game.Level
	currentMouseX  int32
	currentMouseY  int32
	playerInit     bool
	fontTextureMap map[string]*sdl.Texture
	soundFileMap   map[string]*mix.Chunk
	uiElementMap   map[string]*uiButton
	mapMoveDelay   int
	mapMoveTimer   int
	muted bool
	paused bool
	randNumGen     *rand.Rand
}

const (
	playerLaserTexture  = "laserBlue01"
	enemyLaserTexture   = "laserGreen02"
	explosionTexture    = "laserBlue10"
	playerLaserSound    = "sfx_laser1"
	enemyLaserSound     = "sfx_laser2"
	impactSound         = "boom7"
	cursorTexture = "cursor_pointer3D"
	hudTexture = "metalPanel"
	innerHudTexture = "metalPanel_plate"
	backgroundTexture = "purple"
	fontSize = 24
)

var fontColor = sdl.Color{0,0,0,1}

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
	ui.playerInit = false
	ui.mapMoveTimer = 0
	ui.mapMoveDelay = 5
	ui.muted = false

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

func (ui *ui) DrawBackground(level *game.Level) {
	if ui.backgroundTexture == nil {
		ui.backgroundTexture = ui.textureMap[backgroundTexture]
	}
	if ui.horzTiles == 0 && ui.vertTiles == 0 {
		_, _, w, h, err := ui.backgroundTexture.Query()
		if err != nil {
			panic(err)
		}
		ui.horzTiles = ui.WinWidth/int(w)
		ui.vertTiles = ui.WinHeight/int(h)
	}
	destRect := &sdl.Rect{0, 0, int32(ui.WinWidth), int32(ui.WinHeight)}
	ui.renderer.Copy(ui.backgroundTexture, nil, destRect)
}

func (ui *ui) DrawSpeedLines(level *game.Level) {
	// TODO: Maybe the number of lines changes as the player moves towards the top of the screen/accelerates?

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
	//Draw Player Hitpoints, eventually the entire HUD
	p := level.Player
	tex := ui.stringToTexture(strconv.Itoa(p.Hitpoints)+" HP", sdl.Color{255, 255, 255, 1})
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}

	//Initialize and Draw HUD
	if ui.hud == nil {
		ui.hud = &hud{}
		ui.hud.texture = ui.textureMap[hudTexture]
		ui.hud.insideTexture = ui.textureMap[innerHudTexture]
		_, _, w, h, err := ui.hud.texture.Query()
		if err != nil {
			panic(err)
		}
		ui.hud.horzTiles = ui.WinWidth/int(w)
		ui.hud.vertTiles = 4
		ui.hud.W = int(w)
		ui.hud.H = int(h)
	}
	for i := 0; i < ui.hud.horzTiles; i++ {
		var tex *sdl.Texture
		if i == 0 {
			tex = ui.textureMap["metalPanel_greenCorner_noBorder"]
			ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(i * ui.hud.W), int32(ui.WinHeight - ui.hud.H), int32(ui.hud.W), int32(ui.hud.H)}, 0, nil, sdl.FLIP_HORIZONTAL)
		} else if i == ui.hud.horzTiles - 1 {
			tex = ui.textureMap["metalPanel_greenCorner_noBorder"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.W), int32(ui.WinHeight - ui.hud.H), int32(ui.hud.W), int32(ui.hud.H)})
		} else {
			tex = ui.textureMap["metalPanel_green_noBorder"]
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(i * ui.hud.W), int32(ui.WinHeight - ui.hud.H), int32(ui.hud.W), int32(ui.hud.H)})
		}
	}

	//Draw Other UI elements
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

	//Copy all elements to the renderer
	ui.renderer.Copy(tex, nil, &sdl.Rect{0, 0, w, h})
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
		// Arbitrary number 5 here to slightly move fire point forward of texture
		level.Player.FireOffsetY = int(h/2) + 5
	}
	player := level.Player
	if player.IsDestroyed {
		// TODO: Lose Scenario... Some kind of modal? Loss of life?
	}
	tex := player.Texture
	//player.Direction = game.FindDegreeRotation(int32(player.Y+player.H/2), int32(player.X+player.W/2), ui.currentMouseY, ui.currentMouseX) - 90
	ui.renderer.Copy(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), int32(player.W), int32(player.H)})

}

func (ui *ui) DrawEnemy(level *game.Level) {
	for _, enemy := range level.Enemies {
		if enemy.Texture == nil {
			tex := ui.textureMap[enemy.TextureName]
			enemy.Texture = tex
			_, _, w, h, err := tex.Query()
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
			ui.renderer.Copy(enemy.Texture, nil, &sdl.Rect{int32(enemy.X), int32(enemy.Y), int32(enemy.W), int32(enemy.H)})
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
		if !ui.checkBulletOutOfBounds(bullet.X, bullet.Y, int32(bullet.W), int32(bullet.H)) && !bullet.DestroyAnimationPlayed {
			if index != i {
				level.Bullets[index] = bullet
			}
			index++
		}
	}
	level.Bullets = level.Bullets[:index]
}

func determineInputType(event *sdl.KeyboardEvent) *game.Input {
	input := &game.Input{}
	switch event.Type {
	case sdl.KEYDOWN:
		input.Pressed = true
		switch event.Keysym.Scancode {
		case sdl.SCANCODE_W:
			input.Type = game.Up
		case sdl.SCANCODE_S:
			input.Type = game.Down
		case sdl.SCANCODE_A:
			input.Type = game.Left
		case sdl.SCANCODE_D:
			input.Type = game.Right
		case sdl.SCANCODE_TAB:
			input.Type = game.Pause
		}
	case sdl.KEYUP:
		input.Pressed = false
		switch event.Keysym.Scancode {
		case sdl.SCANCODE_W:
			input.Type = game.Up
		case sdl.SCANCODE_S:
			input.Type = game.Down
		case sdl.SCANCODE_A:
			input.Type = game.Left
		case sdl.SCANCODE_D:
			input.Type = game.Right
		case sdl.SCANCODE_TAB:
			input.Type = game.Pause
		}
	}
	return input
}

func (ui *ui) determineMouseButtonInput(event *sdl.MouseButtonEvent) *game.Input {
	input := &game.Input{}
	switch event.Type {
	case sdl.MOUSEBUTTONDOWN:
		switch event.Button {
		case sdl.BUTTON_LEFT:
			for _, element := range ui.uiElementMap {
				if element.boundBox.HasIntersection(&sdl.Rect{event.X, event.Y, 1, 1}) {
					element.clicked = true
					element.onClick()
					input.Type = game.None
					input.Pressed = true
				} else {
					input.Type = game.FirePrimary
					input.Pressed = true
				}
			}
			break
		case sdl.BUTTON_RIGHT:
			input.Type = game.FireSecondary
			input.Pressed = true
			break
		default:
			input.Type = game.None
			input.Pressed = false
		}
	case sdl.MOUSEBUTTONUP:
		switch event.Button {
		case sdl.BUTTON_LEFT:
			for _, element := range ui.uiElementMap {
				element.clicked = false
			}
			input.Type = game.FirePrimary
			input.Pressed = false
			break
		case sdl.BUTTON_RIGHT:
			input.Type = game.FireSecondary
			input.Pressed = false
			break
		default:
			input.Type = game.None
			input.Pressed = false
		}
	default:
	}
	return input
}

func (ui *ui) checkMouseHover(event *sdl.MouseMotionEvent) {
	for _, element := range ui.uiElementMap {
		if element.boundBox.HasIntersection(&sdl.Rect{event.X, event.Y, 1, 1}) {
			element.mouseOver = true
		} else if element.mouseOver {
			element.mouseOver = false
		}
	}
}

// Remember to always draw from the ground up
func (ui *ui) Draw(level *game.Level) {
	ui.renderer.Clear()
	if !ui.paused {
		ui.Update(level)
	}
	ui.DrawBackground(level)
	ui.DrawBullet(level)
	ui.DrawPlayer(level)
	ui.DrawEnemy(level)
	ui.DrawExplosions(level)
	ui.DrawUiElements(level)
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
			level = newLevel
			ui.Draw(level)
			break
		default:
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
				ui.inputChan <- ui.determineMouseButtonInput(e)
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
	}
}
