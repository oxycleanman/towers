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
	inputChan chan *game.Input
	levelChan chan *game.Level
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
	tex := ui.textureMap["tank_huge"]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}

	player.W = int(w)
	player.H = int(h)
	player.Move()

	ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(player.X), int32(player.Y), w, h}, float64(player.Direction), &sdl.Point{w/2, h/2}, 0)
}

func (ui *ui) DrawBullet(level *game.Level) {
	tex := ui.textureMap["bulletRed1"]
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}

	bulletCenterX := w / 2
	bulletCenterY := h / 2

	for _, bullet := range level.Bullets {
		fmt.Println("Found bullets, drawing them")
		bullet.Update()
		ui.renderer.CopyEx(tex, nil, &sdl.Rect{int32(bullet.X), int32(bullet.Y), w, h}, float64(bullet.Direction + 180.0), &sdl.Point{bulletCenterX, bulletCenterY}, 0)
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

func determineMouseInput(event *sdl.MouseButtonEvent) *game.Input {
	input := &game.Input{}
	switch event.Type {
	case sdl.MOUSEBUTTONDOWN:
		fmt.Println("Mouse Key Pressed")
		input.Pressed = true
		switch event.Button {
		case sdl.BUTTON_LEFT:
			input.Type = game.FirePrimary
		case sdl.BUTTON_RIGHT:
			input.Type = game.FireSecondary
		}
	default:
		fmt.Println("Some other key event on mouse")
	}
	return input
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

	for {
		frameStart = time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				ui.inputChan <- determineInputType(e)
			case *sdl.MouseButtonEvent:
				ui.inputChan <- determineMouseInput(e)
			}
		}

		select {
		case newLevel := <-ui.levelChan:
			fmt.Println("Drawing New Level")
			ui.Draw(newLevel)
		default:
		}

		elapsedTime = time.Since(frameStart).Seconds()
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = time.Since(frameStart).Seconds()
		}
	}
}