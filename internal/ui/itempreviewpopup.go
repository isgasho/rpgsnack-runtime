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
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type ItemPreviewPopup struct {
	X             int
	Y             int
	Width         int
	Height        int
	item          *data.Item
	closeButton   *Button
	previewButton *Button
	Visible       bool
	widgets       []Widget
	fadeImage     *ebiten.Image
}

func NewItemPreviewPopup(x, y, width, height int) *ItemPreviewPopup {
	previewButton := NewButton(20, 40, 120, 120, "ok")
	closeButton := NewButton(30, 170, 100, 20, "cancel")
	closeButton.Text = "Close"

	fadeImage, err := ebiten.NewImage(16, 16, ebiten.FilterNearest)
	if err != nil {
		panic(err)
	}

	return &ItemPreviewPopup{
		X:             x,
		Y:             y,
		Width:         width,
		Height:        height,
		closeButton:   closeButton,
		previewButton: previewButton,
		fadeImage:     fadeImage,
	}
}

func (i *ItemPreviewPopup) Update() {
	i.previewButton.Update()
	i.closeButton.Update()

	if i.previewButton.Pressed() {
		i.Visible = false
	}

	if i.closeButton.Pressed() {
		i.Visible = false
	}
}

func (i *ItemPreviewPopup) SetItem(item *data.Item) {
	i.item = item
	if i.item == nil || i.item.Preview == "" {
		i.previewButton.Visible = false
		return
	}
	i.previewButton.Visible = true
	i.previewButton.Image = assets.GetImage("items/preview/" + i.item.Preview + ".png")
}

func (i *ItemPreviewPopup) Draw(screen *ebiten.Image) {
	if !i.Visible {
		return
	}

	w, h := i.fadeImage.Size()
	sw, _ := screen.Size()
	sh := consts.TileYNum*consts.TileSize*consts.TileScale + consts.GameMarginTop
	sx := float64(sw) / float64(w)
	sy := float64(sh) / float64(h)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sx, sy)
	op.ColorM.Translate(0, 0, 0, 1)
	op.ColorM.Scale(1, 1, 1, 0.5)
	screen.DrawImage(i.fadeImage, op)

	i.previewButton.Draw(screen)
	i.closeButton.Draw(screen)
}