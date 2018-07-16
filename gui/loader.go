package gui

import (
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
	"strings"
)

func (ui *ui) loadTextures(dirName string) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			newFilepath := dirName + "/" + file.Name() + "/"
			ui.loadTextures(newFilepath)
		} else if !strings.HasPrefix(file.Name(), ".") {
			filename := file.Name()[:len(file.Name())-4]
			if strings.Contains(filename, "meteor") {
				ui.meteorTextureNames = append(ui.meteorTextureNames, filename)
			}
			if strings.Contains(filename, "enemy") || strings.Contains(filename, "ufo") {
				ui.enemyTextureNames = append(ui.enemyTextureNames, filename)
			}
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
		if file.IsDir() {
			newFilepath := dirName + "/" + file.Name() + "/"
			ui.loadSounds(newFilepath)
		} else {
			filename := file.Name()[:len(file.Name())-4]
			filepath := dirName + "/" + file.Name()
			sound, err := mix.LoadWAV(filepath)
			if err != nil {
				panic(err)
			}
			ui.soundFileMap[filename] = sound
		}
	}
}

func (ui *ui) loadUiElements() {
	// Load Buttons
	{
		pauseButton := &uiButton{}
		pauseButton.texture = ui.textureMap["buttonBlue"]
		_, _, w, h, err := pauseButton.texture.Query()
		if err != nil {
			panic(err)
		}
		pauseButton.X = 20
		pauseButton.Y = float64(ui.WinHeight - h - 20)
		pauseButton.BoundBox = &sdl.Rect{int32(pauseButton.X), int32(pauseButton.Y), w, h}
		pauseButton.textTexture = ui.stringToNormalFontTexture("pause", fontColor)
		_, _, tw, th, err := pauseButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := pauseButton.BoundBox.X + pauseButton.BoundBox.W/2 - tw/2
		textY := pauseButton.BoundBox.Y + pauseButton.BoundBox.H/2 - th/2
		pauseButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), tw, th}
		pauseButton.onClick = ui.pause
		ui.clickableElementMap["pauseButton"] = pauseButton
	}
	{
		menuButton := &uiButton{}
		menuButton.texture = ui.textureMap["buttonBlue"]
		_, _, w, h, err := menuButton.texture.Query()
		if err != nil {
			panic(err)
		}
		menuButton.X = ui.hud.horzOffset + 20
		menuButton.Y = float64(int32(ui.WinHeight) - ui.hud.BoundBox.H/2 - 10)
		menuButton.BoundBox = &sdl.Rect{int32(menuButton.X), int32(menuButton.Y), w, h}
		menuButton.textTexture = ui.stringToNormalFontTexture("menu", fontColor)
		_, _, tw, th, err := menuButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := menuButton.BoundBox.X + menuButton.BoundBox.W/2 - tw/2
		textY := menuButton.BoundBox.Y + menuButton.BoundBox.H/2 - th/2
		menuButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), tw, th}
		menuButton.onClick = ui.openCloseMenu
		ui.clickableElementMap["menuButton"] = menuButton
	}
	{
		pauseButton := ui.clickableElementMap["pauseButton"]
		muteButton := &uiButton{}
		muteButton.texture = ui.textureMap["buttonRed"]
		_, _, w, h, err := muteButton.texture.Query()
		if err != nil {
			panic(err)
		}
		muteButton.X = pauseButton.X + float64(pauseButton.BoundBox.W + 20)
		muteButton.Y = float64(ui.WinHeight - h - 20)
		muteButton.BoundBox = &sdl.Rect{int32(muteButton.X), int32(muteButton.Y), w, h}
		muteButton.textTexture = ui.stringToNormalFontTexture("mute", fontColor)
		_, _, tw, th, err := pauseButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := muteButton.BoundBox.X + muteButton.BoundBox.W/2 - tw/2
		textY := muteButton.BoundBox.Y + muteButton.BoundBox.H/2 - th/2
		muteButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), tw, th}
		muteButton.onClick = ui.mute
		ui.clickableElementMap["muteButton"] = muteButton
	}

	// Load speed lines
	lineTexture := ui.textureMap["speedLine"]
	_, _, w, h, err := lineTexture.Query()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {
		line := &uiElement{}
		line.texture = lineTexture
		line.X = float64(ui.randNumGen.Intn(int(ui.WinWidth)))
		line.Y = -float64(ui.randNumGen.Intn(int(ui.WinHeight)))
		line.BoundBox = &sdl.Rect{}
		line.BoundBox.W = w
		line.BoundBox.H = h
		line.BoundBox.X = int32(line.X)
		line.BoundBox.Y = int32(line.Y)
		ui.uiSpeedLines = append(ui.uiSpeedLines, line)
	}
}
