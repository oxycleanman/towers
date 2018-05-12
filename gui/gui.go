package gui

import (
	"github.com/oxycleanman/towers/game"
	"github.com/veandco/go-sdl2/sdl"
	"image/png"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type ui struct {
	WinWidth       int
	WinHeight      int
	renderer       *sdl.Renderer
	window         *sdl.Window
	textureMap     map[string]*sdl.Texture
	keyboardState  []uint8
	inputChan      chan *game.Input
	levelChan      chan *game.Level
	leftButtonDown bool
	bulletTimer    int
	currentMouseX  int32
	currentMouseY  int32
	playerInit     bool
}

func init() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
}

func NewUi(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.WinHeight = 768
	ui.WinWidth = 1280
	ui.textureMap = make(map[string]*sdl.Texture)
	ui.playerInit = false

	window, err := sdl.CreateWindow("Towers", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 1280, 720, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window

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

func (ui *ui) initBullet(level *game.Level) *game.Bullet {
	bullet := &game.Bullet{}
	bullet.TextureName = "bulletDark1"
	tex := ui.textureMap[bullet.TextureName]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	bullet.Speed = 10.0
	bullet.Texture = tex
	bullet.FlashCounter = 0
	bullet.FireAnimationPlayed = false
	bullet.DestroyAnimationPlayed = false
	bullet.Damage = 50
	bullet.W = int(w)
	bullet.H = int(h)
	bullet.Direction = level.Player.Direction
	bullet.X = (level.Player.X + level.Player.FireOffsetX) - bullet.W/2
	bullet.Y = (level.Player.Y + level.Player.FireOffsetY) - bullet.H/2
	return bullet
}

func (ui *ui) initPlayer(level *game.Level) {
	player := &game.Player{}
	player.TextureName = "tank_huge"
	tex := ui.textureMap[player.TextureName]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	player.IsDestroyed = false
	player.Hitpoints = 100
	player.Speed = 1.0
	player.W = int(w)
	player.H = int(h)
	player.X = ui.WinWidth/2 - player.W/2
	player.Y = ui.WinHeight/2 - player.H/2
	player.FireOffsetX = player.W / 2
	player.FireOffsetY = player.H / 2
	player.Texture = tex
	level.Player = player
	ui.playerInit = true
}

func (ui *ui) initEnemy(level *game.Level) *game.Enemy {
	enemy := &game.Enemy{}
	enemy.TextureName = "tank_dark"
	tex := ui.textureMap[enemy.TextureName]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	enemy.IsDestroyed = false
	enemy.Hitpoints = 50
	enemy.Speed = 1.0
	enemy.W = int(w)
	enemy.H = int(h)
	enemy.X = 300 - enemy.W/2
	enemy.Y = 300 - enemy.H/2
	enemy.FireOffsetX = enemy.W / 2
	enemy.FireOffsetY = enemy.H / 2
	enemy.Texture = tex
	return enemy
}

func (ui *ui) DrawGround() {
	w, h := ui.window.GetSize()
	numTilesX := w / 128
	numTilesY := h / 128

	for y := int32(0); y <= numTilesY; y++ {
		for x := int32(0); x < numTilesX; x++ {
			destRect := &sdl.Rect{x * 128, y * 128, 128, 128}
			ui.renderer.Copy(ui.textureMap["tileGrass1"], nil, destRect)
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

func (ui *ui) DrawPlayer(level *game.Level) {
	player := level.Player
	tex := player.Texture
	player.Direction = game.FindDegreeRotation(int32(player.Y+player.H/2), int32(player.X+player.W/2), ui.currentMouseY, ui.currentMouseX) - 90
	player.Move()
	ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), int32(player.W), int32(player.H)}, float64(player.Direction), nil, 0)

}

func (ui *ui) DrawEnemy(level *game.Level) {
	player := level.Player
	for _, enemy := range level.Enemies {
		if !enemy.IsDestroyed {
			enemy.Direction = game.FindDegreeRotation(int32(enemy.Y+enemy.H/2), int32(enemy.X+enemy.W/2), int32(player.Y+player.H/2), int32(player.X+player.W/2)) - 90
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

func (ui *ui) DrawBullet(level *game.Level) {
	if ui.leftButtonDown == true && ui.bulletTimer == 10 {
		bullet := ui.initBullet(level)
		level.Bullets = append(level.Bullets, bullet)
		ui.bulletTimer = 0
	} else if ui.leftButtonDown == true {
		ui.bulletTimer++
	}

	index := 0
	for i, bullet := range level.Bullets {
		tex := bullet.Texture
		bullet.Update()
		if bullet.FlashCounter < 5 && !bullet.FireAnimationPlayed {
			fireTex := ui.textureMap["explosionSmoke2"]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			posX := level.Player.X + level.Player.FireOffsetX - int(w/4)
			posY := level.Player.Y + level.Player.FireOffsetY - int(h/4)
			ui.renderer.CopyEx(fireTex, nil, &sdl.Rect{int32(posX), int32(posY), w / 2, h / 2}, float64(bullet.Direction), nil, sdl.FLIP_NONE)
			bullet.FlashCounter++
		}
		if bullet.FlashCounter >= 5 && !bullet.FireAnimationPlayed {
			bullet.FlashCounter = 0
			bullet.FireAnimationPlayed = true
		}
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

func (ui *ui) checkCollisions(level *game.Level) {
	for _, bullet := range level.Bullets {
		for _, enemy := range level.Enemies {
			if game.CheckCollision(enemy, bullet) && !bullet.IsColliding && !enemy.IsDestroyed {
				bullet.IsColliding = true
				enemy.Hitpoints -= bullet.Damage
				if enemy.Hitpoints <= 0 {
					enemy.IsDestroyed = true
				}
			}
		}
	}
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

func (ui *ui) determineMouseButtonInput(event *sdl.MouseButtonEvent) {
	switch event.Type {
	case sdl.MOUSEBUTTONDOWN:
		switch event.Button {
		case sdl.BUTTON_LEFT:
			ui.leftButtonDown = true
		case sdl.BUTTON_RIGHT:
			//	input.Type = game.FireSecondary
		}
	case sdl.MOUSEBUTTONUP:
		switch event.Button {
		case sdl.BUTTON_LEFT:
			ui.leftButtonDown = false
		case sdl.BUTTON_RIGHT:
			//input.Type = game.FireSecondary
		}
	default:
	}
}

// Remember to always draw from the ground up
func (ui *ui) Draw(level *game.Level) {
	ui.renderer.Clear()
	ui.DrawGround()
	ui.checkCollisions(level)
	ui.DrawPlayer(level)
	ui.DrawEnemy(level)
	ui.DrawBullet(level)
	ui.DrawExplosions(level)
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
			if !ui.playerInit {
				ui.initPlayer(level)
			}
			ui.Draw(level)
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
				if input.Type == game.Pause {
					level.Enemies = append(level.Enemies, ui.initEnemy(level))
				}
			case *sdl.MouseButtonEvent:
				ui.determineMouseButtonInput(e)
			case *sdl.MouseMotionEvent:
				ui.currentMouseX = e.X
				ui.currentMouseY = e.Y
			}
		}

		elapsedTime = time.Since(frameStart).Seconds() * 1000
		if elapsedTime < targetFrameTime {
			sdl.Delay(uint32(targetFrameTime - elapsedTime))
			elapsedTime = time.Since(frameStart).Seconds()
		}
	}
}
