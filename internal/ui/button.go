// Copyright 2017 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

type Button struct {
	x             int
	y             int
	width         int
	height        int
	touchExpand   int
	visible       bool
	text          string
	disabled      bool
	Image         *ebiten.Image
	PressedImage  *ebiten.Image
	DisabledImage *ebiten.Image
	dropShadow    bool
	pressing      bool
	soundName     string
	showFrame     bool
	textColor     color.Color
	Lang          language.Tag

	onPressed func(button *Button)
}

func NewButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		width:      width,
		height:     height,
		visible:    true,
		soundName:  soundName,
		dropShadow: false,
		showFrame:  true,
		textColor:  color.White,
	}
}

func NewTextButton(x, y, width, height int, soundName string) *Button {
	return &Button{
		x:          x,
		y:          y,
		width:      width,
		height:     height,
		visible:    true,
		soundName:  soundName,
		dropShadow: false,
		showFrame:  false,
		textColor:  color.White,
	}
}

func NewImageButton(x, y int, image *ebiten.Image, pressedImage *ebiten.Image, soundName string) *Button {
	w, h := image.Size()
	return &Button{
		x:             x,
		y:             y,
		width:         w,
		height:        h,
		visible:       true,
		Image:         image,
		PressedImage:  pressedImage,
		DisabledImage: nil,
		soundName:     soundName,
		dropShadow:    true,
		showFrame:     true,
		textColor:     color.White,
	}
}

func (b *Button) SetX(x int) {
	b.x = x
}

func (b *Button) SetY(y int) {
	b.y = y
}

func (b *Button) SetWidth(width int) {
	b.width = width
}

func (b *Button) SetText(text string) {
	b.text = text
}

func (b *Button) Show() {
	b.visible = true
}

func (b *Button) Hide() {
	b.visible = false
}

func (b *Button) Enable() {
	b.disabled = false
}

func (b *Button) Disable() {
	b.disabled = true
}

func (b *Button) SetOnPressed(onPressed func(button *Button)) {
	b.onPressed = onPressed
}

func (b *Button) includesInput(offsetX, offsetY int) bool {
	x, y := input.Position()
	x = int(float64(x) / consts.TileScale)
	y = int(float64(y) / consts.TileScale)
	x -= offsetX
	y -= offsetY

	buttonWidth := b.width + b.touchExpand*2
	buttonHeight := b.height + b.touchExpand*2
	buttonX := b.x - b.touchExpand
	buttonY := b.y - b.touchExpand

	if buttonX <= x && x < buttonX+buttonWidth && buttonY <= y && y < buttonY+buttonHeight {
		return true
	}
	return false
}

func (b *Button) update(visible bool, offsetX, offsetY int) {
	if !visible {
		return
	}
	if !b.visible {
		return
	}
	if b.disabled {
		return
	}
	if !b.pressing {
		if !input.Triggered() {
			return
		}
	}
	if !input.Pressed() {
		b.pressing = false
		if b.onPressed != nil {
			b.onPressed(b)
		}
		if b.soundName != "" {
			audio.PlaySE(b.soundName, 1.0)
		}
		return
	}
	b.pressing = b.includesInput(offsetX, offsetY)
}

func (b *Button) Update() {
	b.update(true, 0, 0)
}

func (b *Button) UpdateAsChild(visible bool, offsetX, offsetY int) {
	b.update(visible, offsetX, offsetY)
}

func (b *Button) Draw(screen *ebiten.Image) {
	b.DrawAsChild(screen, 0, 0)
}

func (b *Button) DrawAsChild(screen *ebiten.Image, offsetX, offsetY int) {
	if !b.visible {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.x+offsetX), float64(b.y+offsetY))
	op.GeoM.Scale(consts.TileScale, consts.TileScale)

	opacity := uint8(255)
	if b.showFrame {
		if b.Image != nil {
			image := b.Image
			if b.disabled {
				if b.DisabledImage != nil {
					image = b.DisabledImage
				}
			} else {
				if b.pressing {
					if b.PressedImage == nil {
						op.ColorM.ChangeHSV(0, 0, 1)
						op.ColorM.Scale(0.5, 0.5, 0.5, 1)
					} else {
						image = b.PressedImage
					}
				}
			}
			screen.DrawImage(image, op)
		} else {
			img := assets.GetImage("system/common/9patch_frame_off.png")
			if b.pressing {
				img = assets.GetImage("system/common/9patch_frame_on.png")
			}

			if b.disabled {
				op.ColorM.ChangeHSV(0, 0, 1)
				op.ColorM.Scale(0.5, 0.5, 0.5, 1)
			}
			DrawNinePatches(screen, img, b.width, b.height, &op.GeoM, &op.ColorM)
		}
	} else {
		if b.pressing {
			opacity = uint8(127)
		}
	}

	_, th := font.MeasureSize(b.text)
	tx := (b.x + offsetX) * consts.TileScale
	tx += b.width * consts.TileScale / 2

	ty := (b.y + offsetY) * consts.TileScale
	ty += (b.height*consts.TileScale - th*consts.TextScale) / 2

	cr, cg, cb, ca := b.textColor.RGBA()
	r8 := uint8(cr >> 8)
	g8 := uint8(cg >> 8)
	b8 := uint8(cb >> 8)
	a8 := uint8(ca >> 8)
	var c color.Color = color.RGBA{r8, g8, b8, uint8(uint16(a8) * uint16(opacity) / 255)}
	if b.disabled {
		c = color.RGBA{r8, g8, b8, uint8(uint16(a8) * uint16(opacity) / (2 * 255))}
	}
	l := b.Lang
	if l == language.Und {
		l = lang.Get()
	}
	if b.dropShadow {
		font.DrawTextLang(screen, b.text, tx+consts.TextScale, ty+consts.TextScale, consts.TextScale, data.TextAlignCenter, color.Black, len([]rune(b.text)), l)
	}
	font.DrawTextLang(screen, b.text, tx, ty, consts.TextScale, data.TextAlignCenter, c, len([]rune(b.text)), l)
}
