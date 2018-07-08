package gui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/oxycleanman/towers/game"
)

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
