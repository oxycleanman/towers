package gui

import (
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"image/png"
	"io/ioutil"
	"time"
	"github.com/oxycleanman/towers/game"
	"fmt"
)

type ui struct {
	WinWidth  int
	WinHeight int
	renderer  *sdl.Renderer
	window    *sdl.Window
	textureMap map[string]*sdl.Texture
	keyboardState []uint8
	inputChan chan *game.Input
	levelChan chan *game.Level
	leftButtonDown bool
	bulletTimer int
}

func init() {
	fmt.Println("Init GUI")
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

	window, err := sdl.CreateWindow("Towers", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 1280, 720, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.window = window

	renderer, err := sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	ui.renderer = renderer
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.loadTextures("gui/assets/images")

	//_, _, w, h, err := ui.textureMap["tileGrass1"].Query()
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println("Created new GUI")
	return ui
}

func (ui *ui) loadTextures(dirName string) {
	fmt.Println("Loading Textures into Texture Map")
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		filename := file.Name()[:len(file.Name()) - 4]
		filepath := dirName + "/" + file.Name()
		tex := imgFileToTexture(ui.renderer, filepath)
		ui.textureMap[filename] = tex
	}
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

func (ui *ui) DrawPlayer(level *game.Level) {
	player := level.Player
	tex := ui.textureMap[player.TextureName]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}

	player.W = int(w)
	player.H = int(h)
	player.FireOffsetX = player.W/2
	player.FireOffsetY = player.H/2

	player.Move()
	ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), w, h}, float64(player.Direction), nil, 0)
}

func (ui *ui) DrawBullet(level *game.Level) {
	if ui.leftButtonDown == true && ui.bulletTimer == 20 {
		bullet := game.NewBullet("bulletDark1")
		tex := ui.textureMap[bullet.TextureName]
		_, _, w, h, err := tex.Query()
		if err != nil {
			panic(err)
		}
		bullet.IsNew = true
		bullet.W = int(w)
		bullet.H = int(h)
		bullet.Direction = level.Player.Direction
		bullet.X = (level.Player.X + level.Player.FireOffsetX) - bullet.W/2
		bullet.Y = (level.Player.Y + level.Player.FireOffsetY) - bullet.H/2
		switch bullet.Direction {
		case game.DUp:
			bullet.Yvel = -10
		case game.DDown:
			bullet.Yvel = 10
		case game.DLeft:
			bullet.Xvel = -10
		case game.DRight:
			bullet.Xvel = 10
		}
		level.Bullets = append(level.Bullets, bullet)
		ui.bulletTimer = 0
	} else if ui.leftButtonDown == true {
		ui.bulletTimer++
	}

	index := 0
	for i, bullet := range level.Bullets {
		tex := ui.textureMap[bullet.TextureName]
		_, _, w, h, err := tex.Query()
		if err != nil {
			panic(err)
		}
		bullet.Update()
		if bullet.IsNew {
			fireTex := ui.textureMap["explosionSmoke2"]
			_, _, w, h, err := fireTex.Query()
			if err != nil {
				panic(err)
			}
			posX := level.Player.X + level.Player.FireOffsetX - int(w/4)
			posY := level.Player.Y + level.Player.FireOffsetY - int(h/4)
			ui.renderer.Copy(fireTex, nil, &sdl.Rect{int32(posX), int32(posY), w/2, h/2})
		}
		ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), w, h}, float64(bullet.Direction + 180.0), nil, sdl.FLIP_NONE)

		// Keep bullets in the slice that aren't out of bounds (drop the bullets that go off screen so they aren't redrawn)
		if !ui.checkBulletOutOfBounds(bullet.X, bullet.Y, w, h) {
			if index != i {
				bullet.IsNew = false
				level.Bullets[index] = bullet
			}
			index++
		}
	}
	level.Bullets = level.Bullets[:index]
}

func (ui *ui) checkBulletOutOfBounds(x, y int, w,h int32) bool {
	if x > ui.WinWidth + int(w) || x < int(0 - w) || y > ui.WinHeight + int(h) || y < int(0 - h) {
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

func (ui *ui) determineMouseInput(event *sdl.MouseButtonEvent) {
	switch event.Type {
	case sdl.MOUSEBUTTONDOWN:
		switch event.Button {
		case sdl.BUTTON_LEFT:
			ui.leftButtonDown = true
		//case sdl.BUTTON_RIGHT:
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

func (ui *ui) UpdateEntitiesAndDraw(level *game.Level) {

}

func (ui *ui) Draw(level *game.Level) {
	ui.renderer.Clear()
	ui.DrawGround()
	ui.DrawPlayer(level)
	ui.DrawBullet(level)
	ui.renderer.Present()
}

func (ui *ui) Run() {
	var frameStart time.Time
	var elapsedTime float64
	var level *game.Level

	for {
		frameStart = time.Now()

		select {
		case newLevel := <-ui.levelChan:
			level = newLevel
			ui.Draw(level)
		default:
			ui.Draw(level)
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				ui.inputChan <- determineInputType(e)
			case *sdl.MouseButtonEvent:
				ui.determineMouseInput(e)
			}
		}

		elapsedTime = time.Since(frameStart).Seconds()
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = time.Since(frameStart).Seconds()
		}
	}
}