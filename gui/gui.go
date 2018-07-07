package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"image/png"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"math/rand"
	"github.com/veandco/go-sdl2/mix"
	"fmt"
)

type uiElement struct {
	game.Pos
	game.Size
	boundBox *sdl.Rect
	textBoundBox *sdl.Rect
	mouseOver bool
	clicked bool
	onClick func()
	texture, textTexture *sdl.Texture
}

type ui struct {
	WinWidth       int
	WinHeight      int
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
	uiElementMap   map[string]*uiElement
	mapMoveDelay   int
	mapMoveTimer   int
	muted bool
	randNumGen     *rand.Rand
}

const (
	playerLaserTexture  = "laserBlue01"
	enemyLaserTexture   = "laserGreen02"
	explosionTexture    = "laserBlue10"
	playerLaserSound    = "sfx_laser1"
	enemyLaserSound     = "sfx_laser2"
	impactSound         = "boom7"
	cursor              = "cursor"
)

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
	ui.textureMap = make(map[string]*sdl.Texture)
	ui.fontTextureMap = make(map[string]*sdl.Texture)
	ui.soundFileMap = make(map[string]*mix.Chunk)
	ui.uiElementMap = make(map[string]*uiElement)
	ui.playerInit = false
	ui.mapMoveTimer = 0
	ui.mapMoveDelay = 5
	ui.muted = false

	var err error
	ui.window, err = sdl.CreateWindow("Towers", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, int32(ui.WinWidth), int32(ui.WinHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.font, err = ttf.OpenFont("gui/assets/fonts/kenvector_future.ttf", 32)
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

func (ui *ui) loadTextures(dirName string) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			newFilepath := dirName + "/" + file.Name() + "/"
			ui.loadTextures(newFilepath)
		} else {
			filename := file.Name()[:len(file.Name())-4]
			filepath := dirName + "/" + file.Name()
			tex := imgFileToTexture(ui.renderer, filepath)
			ui.textureMap[filename] = tex
		}
	}
}

func (ui *ui) loadSounds(dirName string) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		filename := file.Name()[:len(file.Name())-4]
		filepath := dirName + "/" + file.Name()
		sound, err := mix.LoadWAV(filepath)
		if err != nil {
			panic(err)
		}
		ui.soundFileMap[filename] = sound
	}
}

func imgFileToTexture(renderer *sdl.Renderer, filename string) *sdl.Texture {
	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	tex.Update(nil, pixels, w*4)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return tex
}

func (ui *ui) loadUiElements() {
	{
		testButton := &uiElement{}
		testButton.texture = ui.textureMap["buttonBlue"]
		_, _, w, h, err := testButton.texture.Query()
		if err != nil {
			panic(err)
		}
		testButton.X = 20
		testButton.Y = ui.WinHeight - int(h) - 20
		testButton.W = int(w)
		testButton.H = int(h)
		testButton.boundBox = &sdl.Rect{int32(testButton.X), int32(testButton.Y), int32(testButton.W), int32(testButton.H)}
		testButton.textTexture = ui.stringToTexture("button 1", sdl.Color{0, 0, 0, 1})
		_, _, tw, th, err := testButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := (testButton.X + testButton.W/2) - int(tw/2)
		textY := (testButton.Y + testButton.H/2) - int(th/2)
		testButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), tw, th}
		ui.uiElementMap["button1"] = testButton
	}
	{
		testButton := ui.uiElementMap["button1"]
		muteButton := &uiElement{}
		muteButton.texture = ui.textureMap["buttonRed"]
		_, _, w, h, err := muteButton.texture.Query()
		if err != nil {
			panic(err)
		}
		muteButton.X = testButton.X + testButton.W + 20
		muteButton.Y = ui.WinHeight - int(h) - 20
		muteButton.W = int(w)
		muteButton.H = int(h)
		muteButton.boundBox = &sdl.Rect{int32(muteButton.X), int32(muteButton.Y), int32(muteButton.W), int32(muteButton.H)}
		muteButton.textTexture = ui.stringToTexture("mute", sdl.Color{0,0,0,1})
		_, _, tw, th, err := testButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := (muteButton.X + muteButton.W/2) - int(tw/2)
		textY := (muteButton.Y + muteButton.H/2) - int(th/2)
		muteButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), tw, th}
		muteButton.onClick = ui.mute
		ui.uiElementMap["muteButton"] = muteButton
	}
}

func (ui *ui) mute() {
	if mix.Volume(-1, -1) > 0 {
		mix.Volume(-1, 0)
		ui.muted = true
		ui.uiElementMap["muteButton"].textTexture = ui.stringToTexture("unmute", sdl.Color{0,0,0,1})
	} else {
		mix.Volume(-1, 128)
		ui.muted = false
		ui.uiElementMap["muteButton"].textTexture = ui.stringToTexture("mute", sdl.Color{0,0,0,1})
	}
}

func (ui *ui) DrawBackground(level *game.Level) {
	destRect := &sdl.Rect{0, 0, int32(ui.WinWidth), int32(ui.WinHeight)}
	ui.renderer.Copy(ui.textureMap["purple"], nil, destRect)
}

func (ui *ui) DrawCursor() {
	tex := ui.textureMap[cursor]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	ui.renderer.Copy(tex, nil, &sdl.Rect{ui.currentMouseX, ui.currentMouseY, w, h})
}

func (ui *ui) DrawUiElements(level *game.Level) {
	//Draw Player Hitpoints, eventually the entire HUD
	p := level.Player
	tex := ui.stringToTexture(strconv.Itoa(p.Hitpoints)+" HP", sdl.Color{255, 255, 255, 1})
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
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

func (ui *ui) stringToTexture(s string, color sdl.Color) *sdl.Texture {
	if ui.fontTextureMap[s] != nil {
		return ui.fontTextureMap[s]
	}
	font, err := ui.font.RenderUTF8Blended(s, color)
	if err != nil {
		panic(err)
	}
	tex, err := ui.renderer.CreateTextureFromSurface(font)
	if err != nil {
		panic(err)
	}
	return tex
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
		level.Player.LeftBound = level.Player.X - level.Player.W/2
		level.Player.RightBound = level.Player.X + level.Player.W/2
		level.Player.TopBound = level.Player.Y - level.Player.H/2
		level.Player.BottomBound = level.Player.Y + level.Player.H/2
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
	player.Move(0, ui.WinHeight, 0, ui.WinWidth)
	ui.renderer.Copy(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), int32(player.W), int32(player.H)})

}

func (ui *ui) SpawnEnemies(level *game.Level) {
	if level.EnemySpawnTimer >= 100 && len(level.Enemies) < 1 {
		level.Enemies = append(level.Enemies, level.InitEnemy())
		level.EnemySpawnTimer = 0
	} else {
		level.EnemySpawnTimer++
	}
}

func (ui *ui) DrawEnemy(level *game.Level) {
	//player := level.Player
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
			enemy.LeftBound = enemy.X - enemy.W/2
			enemy.RightBound = enemy.X + enemy.W/2
			enemy.TopBound = enemy.Y - enemy.H/2
			enemy.BottomBound = enemy.Y + enemy.H/2
			enemy.FireOffsetX = 0
			// Arbitrary number 5 here to slightly move fire point forward of texture
			enemy.FireOffsetY = int(h/2) + 5
		}
		if !enemy.IsDestroyed {
			//enemy.Update(level)
			ui.CheckFiring(level, enemy)
			//enemy.Direction = game.FindDegreeRotation(int32(enemy.Y), int32(enemy.X), int32(player.Y), int32(player.X)) - 90
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
			//_, _, w, h, err := tex.Query()
			//if err != nil {
			//	panic(err)
			//}
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

func (ui *ui) CheckFiring(level *game.Level, entity game.Shooter) {
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
		entity.SetFireTimer(timer + 1)
	}
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
			bullet.LeftBound = bullet.X - bullet.W/2
			bullet.RightBound = bullet.X + bullet.W/2
			bullet.TopBound = bullet.Y - bullet.H/2
			bullet.BottomBound = bullet.Y + bullet.H/2
			bullet.Direction = bullet.FiredBy.Direction
			bullet.X = (bullet.FiredBy.X + bullet.FiredBy.W/2) - bullet.W/2
			bullet.Y = (bullet.FiredBy.Y + bullet.FiredBy.H/2) - bullet.H/2
		}
		tex := bullet.Texture
		bullet.Update()

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

func (ui *ui) checkBulletOutOfBounds(x, y int, w, h int32) bool {
	if x > ui.WinWidth+int(w) || x < int(0-w) || y > ui.WinHeight+int(h) || y < int(0-h) {
		return true
	}
	return false
}

func (ui *ui) checkCollisions(level *game.Level) {
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
							}
							bulletImpactSound := ui.soundFileMap[impactSound]
							bulletImpactSound.Volume(45)
							bulletImpactSound.Play(-1, 0)
						}
					}
				}
			} else {
				playerRect := &sdl.Rect{int32(level.Player.X), int32(level.Player.Y), int32(level.Player.W), int32(level.Player.H)}
				if playerRect.HasIntersection(bulletRect) {
					bullet.IsColliding = true
					level.Player.Hitpoints -= bullet.Damage
					if level.Player.Hitpoints <= 0 {
						level.Player.IsDestroyed = true
					}
				}
			}
		}
	}
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
					fmt.Println("Clicked a button")
					element.onClick()
					//sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_INFORMATION, "You Fucking Did it Bro", "Fuck Yeah", ui.window)
					input.Type = game.None
					input.Pressed = false
				} else {
					input.Type = game.FirePrimary
					input.Pressed = true
				}
			}
		case sdl.BUTTON_RIGHT:
			input.Type = game.FireSecondary
			input.Pressed = true
		}
	case sdl.MOUSEBUTTONUP:
		switch event.Button {
		case sdl.BUTTON_LEFT:
			for _, element := range ui.uiElementMap {
				element.clicked = false
			}
			input.Type = game.FirePrimary
			input.Pressed = false
		case sdl.BUTTON_RIGHT:
			input.Type = game.FireSecondary
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
	ui.checkCollisions(level)
	ui.DrawBackground(level)
	ui.DrawBullet(level)
	ui.DrawPlayer(level)
	ui.SpawnEnemies(level)
	ui.DrawEnemy(level)
	ui.CheckFiring(level, level.Player)
	//level.CheckBulletCollisions()
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
				ui.inputChan <- input
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
