package gui

import (
	"fmt"
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"image/png"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type GameTile struct {
	TextureName string
	IsTrack bool
	game.Pos
}

type ui struct {
	WinWidth                                     int
	WinHeight                                    int
	renderer                                     *sdl.Renderer
	window                                       *sdl.Window
	font                                         *ttf.Font
	textureMap                                   map[string]*sdl.Texture
	keyboardState                                []uint8
	inputChan                                    chan *game.Input
	levelChan                                    chan *game.Level
	currentMouseX                                int32
	currentMouseY                                int32
	playerInit                                   bool
	levelMap                                     [][]*GameTile
	tileMap                                      map[int]string
	testMap                                      [][]*GameTile
	topBound, bottomBound, leftBound, rightBound int
	fontTextureMap                               map[string]*sdl.Texture
	mapMoveDelay                                 int
	mapMoveTimer                                 int
}

func init() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	if err := ttf.Init(); err != nil {
		panic(err)
	}
}

func NewUi(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.WinHeight = 1080
	ui.WinWidth = 1920
	ui.topBound = int(float32(ui.WinHeight) * 0.25)
	ui.bottomBound = int(float32(ui.WinHeight) * 0.75)
	ui.leftBound = int(float32(ui.WinWidth) * 0.25)
	ui.rightBound = int(float32(ui.WinWidth) * 0.75)
	ui.textureMap = make(map[string]*sdl.Texture)
	ui.fontTextureMap = make(map[string]*sdl.Texture)
	ui.playerInit = false
	ui.mapMoveTimer = 0
	ui.mapMoveDelay = 5

	if ui.WinHeight%128 != 0 {
		ui.levelMap = make([][]*GameTile, (ui.WinHeight/128)+1)
	} else {
		ui.levelMap = make([][]*GameTile, ui.WinHeight/128)
	}
	for i := range ui.levelMap {
		ui.levelMap[i] = make([]*GameTile, ui.WinWidth/128)
	}
	ui.testMap = make([][]*GameTile, 300)
	for i := range ui.testMap {
		ui.testMap[i] = make([]*GameTile, 300)
	}
	for y := range ui.testMap {
		for x := range ui.testMap[y] {
			if y%2 == 0 {
				if x%2 == 0 {
					ui.testMap[y][x] = &GameTile{"tileGrass1", false, game.Pos{x, y}}
				} else {
					ui.testMap[y][x] = &GameTile{"tileSand1", false, game.Pos{x, y}}
				}
			} else {
				if x%2 == 0 {
					ui.testMap[y][x] = &GameTile{"tileSand1", false, game.Pos{x, y}}
				} else {
					ui.testMap[y][x] = &GameTile{"tileGrass1", false, game.Pos{x, y}}
				}
			}
		}
	}
	for y := range ui.levelMap {
		for x := range ui.levelMap[y] {
			ui.levelMap[y][x] = ui.testMap[y][x]
		}
	}
	ui.tileMap = make(map[int]string)
	var err error
	ui.window, err = sdl.CreateWindow("Towers", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, int32(ui.WinWidth), int32(ui.WinHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.font, err = ttf.OpenFont("gui/assets/fonts/sharpretro.ttf", 32)
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

	ui.loadTextures("gui/assets/images")
	return ui
}

func (ui *ui) loadTextures(dirName string) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		filename := file.Name()[:len(file.Name())-4]
		filepath := dirName + "/" + file.Name()
		tex := imgFileToTexture(ui.renderer, filepath)
		ui.textureMap[filename] = tex
	}
}

// TODO: Need to rework this to draw a meaningful map
func (ui *ui) DrawGround(level *game.Level) {
	// How do I know which part of the testMap to draw as the ground, and how do I change that as the player moves around?
	p := level.Player

	if ui.mapMoveTimer >= ui.mapMoveDelay && (p.AtTop || p.AtBottom || p.AtLeft || p.AtRight) {
		// Need to update level map
		ui.mapMoveTimer = 0
		for y := range ui.levelMap {
			for x := range ui.levelMap[y] {
				currentTile := ui.levelMap[y][x]
				newX, newY := currentTile.X, currentTile.Y
				if p.AtBottom {
					newY += 1
				}
				if p.AtRight {
					newX += 1
				}
				if p.AtLeft {
					newX -= 1
				}
				if p.AtTop {
					newY -= 1
				}
				if newY < 0 {
					newY = 0
				}
				if newY >= len(ui.testMap) {
					newY = len(ui.testMap) - 1
				}
				if newX < 0 {
					newX = 0
				}
				if newX >= len(ui.testMap[y]) {
					newX = len(ui.testMap[y]) - 1
				}
				if (newY < len(ui.testMap) && newY > 0) || (newX < len(ui.testMap[y]) && newX > 0) {
					fmt.Println("Getting tile at", newX, newY)
					nextTile := ui.testMap[newY][newX]
					//fmt.Println("Next Tile Texture", nextTile.TextureName)
					destRect := &sdl.Rect{int32(currentTile.X * 128), int32(currentTile.Y * 128), 128, 128}
					ui.renderer.Copy(ui.textureMap[nextTile.TextureName], nil, destRect)
					ui.levelMap[y][x].TextureName = nextTile.TextureName
				}
			}
		}
	} else {
		// Draw level map without updating it
		ui.mapMoveTimer++
		for y := range ui.levelMap {
			for x := range ui.levelMap[y] {
				currentTile := ui.levelMap[y][x]
				destRect := &sdl.Rect{int32(currentTile.X * 128), int32(currentTile.Y * 128), 128, 128}
				ui.renderer.Copy(ui.textureMap[currentTile.TextureName], nil, destRect)
			}
		}
	}

}

func (ui *ui) DrawCursor() {
	tex := ui.textureMap["cross-02"]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	ui.renderer.Copy(tex, nil, &sdl.Rect{ui.currentMouseX - w/8, ui.currentMouseY - h/8, w / 4, h / 4})
}

func (ui *ui) DrawUiElements(level *game.Level) {
	p := level.Player
	tex := ui.stringToTexture(strconv.Itoa(p.Hitpoints)+" HP", sdl.Color{255, 255, 255, 1})
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}

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
		level.Player.FireOffsetX = 0
		level.Player.FireOffsetY = 0
	}
	player := level.Player
	tex := player.Texture
	player.Direction = game.FindDegreeRotation(int32(player.Y+player.H/2), int32(player.X+player.W/2), ui.currentMouseY, ui.currentMouseX) - 90
	player.Move(ui.topBound, ui.bottomBound, ui.leftBound, ui.rightBound)
	ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), int32(player.W), int32(player.H)}, float64(player.Direction), nil, 0)

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
	player := level.Player
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
			enemy.FireOffsetX = int(w / 2)
			enemy.FireOffsetY = 0
		}
		if !enemy.IsDestroyed {
			//enemy.Update(level)
			ui.CheckFiring(level, enemy)
			enemy.Direction = game.FindDegreeRotation(int32(enemy.Y), int32(enemy.X), int32(player.Y), int32(player.X)) - 90
			ui.renderer.CopyEx(enemy.Texture, nil, &sdl.Rect{int32(enemy.X), int32(enemy.Y), int32(enemy.W), int32(enemy.H)}, float64(enemy.Direction), nil, sdl.FLIP_NONE)
		}
	}
}

func (ui *ui) DrawExplosions(level *game.Level) {
	index := 0
	for _, enemy := range level.Enemies {
		if enemy.IsDestroyed && !enemy.DestroyedAnimationPlayed {
			seconds := sdl.GetTicks() * 1000
			imageNumber := seconds % 9
			imageName := "explosion0" + strconv.Itoa(int(imageNumber))
			tex := ui.textureMap[imageName]
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			ui.renderer.Copy(tex, nil, &sdl.Rect{int32(enemy.X) - w/16, int32(enemy.Y) - h/16, w / 4, h / 4})
			enemy.DestroyedAnimationCounter++
			if enemy.DestroyedAnimationCounter == 24 {
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
		if isPlayer {
			texName = "bulletBlue1"
		} else {
			texName = "bulletRed1"
		}
		bullet := level.InitBullet(texName)
		bullet.FiredByEnemy = !isPlayer
		bullet.FiredBy = entity.GetSelf()
		bullet.Damage = bullet.FiredBy.Strength
		level.Bullets = append(level.Bullets, bullet)
		entity.SetFireTimer(0)
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
			bullet.Direction = bullet.FiredBy.Direction
			bullet.X = (bullet.FiredBy.X + bullet.FiredBy.W/2) - bullet.W/2 + bullet.FiredBy.FireOffsetX
			bullet.Y = (bullet.FiredBy.Y + bullet.FiredBy.H/2) - bullet.H/2 + bullet.FiredBy.FireOffsetY
		}
		tex := bullet.Texture
		bullet.Update()

		// Fire Animation
		if bullet.FlashCounter < 5 && !bullet.FireAnimationPlayed {
			fireTex := ui.textureMap["explosionSmoke2"]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			posX := (bullet.FiredBy.X + bullet.FiredBy.W/2) - int(w/4) + bullet.FiredBy.FireOffsetX
			posY := (bullet.FiredBy.Y + bullet.FiredBy.H/2) - int(h/4) + bullet.FiredBy.FireOffsetY
			ui.renderer.CopyEx(fireTex, nil, &sdl.Rect{int32(posX), int32(posY), w / 2, h / 2}, float64(bullet.Direction), nil, sdl.FLIP_NONE)
			bullet.FlashCounter++
		}
		if bullet.FlashCounter >= 5 && !bullet.FireAnimationPlayed {
			bullet.FlashCounter = 0
			bullet.FireAnimationPlayed = true
		}

		// Collision Animation && Normal Travel
		if bullet.ExplodeCounter < 5 && bullet.IsColliding && !bullet.DestroyAnimationPlayed {
			fireTex := ui.textureMap["explosion2"]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			bullet.Xvel = 0
			bullet.Yvel = 0
			ui.renderer.CopyEx(fireTex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), w / 2, h / 2}, float64(bullet.Direction), nil, sdl.FLIP_NONE)
			bullet.ExplodeCounter++
		} else {
			if bullet.ExplodeCounter >= 5 && bullet.IsColliding && !bullet.DestroyAnimationPlayed {
				bullet.DestroyAnimationPlayed = true
				bullet.ExplodeCounter = 0
			}
			//point := &sdl.Point{int32(bullet.FiredBy.X), int32(bullet.FiredBy.Y)}
			ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), int32(bullet.W), int32(bullet.H)}, float64(bullet.Direction+180.0), nil, sdl.FLIP_NONE)
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
			input.Type = game.FirePrimary
			input.Pressed = true
		case sdl.BUTTON_RIGHT:
			input.Type = game.FireSecondary
			input.Pressed = true
		}
	case sdl.MOUSEBUTTONUP:
		switch event.Button {
		case sdl.BUTTON_LEFT:
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

// Remember to always draw from the ground up
func (ui *ui) Draw(level *game.Level) {
	ui.renderer.Clear()
	ui.DrawGround(level)
	ui.DrawPlayer(level)
	ui.SpawnEnemies(level)
	ui.DrawEnemy(level)
	ui.CheckFiring(level, level.Player)
	ui.DrawBullet(level)
	level.CheckBulletCollisions()
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
