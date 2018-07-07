package gui

import (
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
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
	{
		pauseButton := &uiButton{}
		pauseButton.texture = ui.textureMap["buttonBlue"]
		_, _, w, h, err := pauseButton.texture.Query()
		if err != nil {
			panic(err)
		}
		pauseButton.X = 20
		pauseButton.Y = ui.WinHeight - int(h) - 20
		pauseButton.W = int(w)
		pauseButton.H = int(h)
		pauseButton.boundBox = &sdl.Rect{int32(pauseButton.X), int32(pauseButton.Y), int32(pauseButton.W), int32(pauseButton.H)}
		pauseButton.textTexture = ui.stringToTexture("pause", fontColor)
		_, _, tw, th, err := pauseButton.textTexture.Query()
		if err != nil {
			panic(err)
		}
		textX := (pauseButton.X + pauseButton.W/2) - int(tw/2)
		textY := (pauseButton.Y + pauseButton.H/2) - int(th/2)
		pauseButton.textBoundBox = &sdl.Rect{int32(textX), int32(textY), tw, th}
		pauseButton.onClick = ui.pause
		ui.uiElementMap["pauseButton"] = pauseButton
	}
	{
		testButton := ui.uiElementMap["pauseButton"]
		muteButton := &uiButton{}
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
		muteButton.textTexture = ui.stringToTexture("mute", fontColor)
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

	//Load speed lines
	lineTexture := ui.textureMap["speedLine"]
	_, _, w, h, err := lineTexture.Query()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {
		line := &uiElement{}
		line.texture = lineTexture
		line.W = int(w)
		line.H = int(h) * 2
		line.X = ui.randNumGen.Intn(ui.WinWidth)
		line.Y = -ui.randNumGen.Intn(ui.WinHeight)
		ui.uiSpeedLines = append(ui.uiSpeedLines, line)
	}
}
