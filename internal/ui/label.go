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

	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Label struct {
	X    int
	Y    int
	text string
}

func NewLabel(x, y int, text string) *Label {
	return &Label{
		X:    x,
		Y:    y,
		text: text,
	}
}

func (l *Label) Update(offsetX, offsetY int) error {
	return nil
}

func (l *Label) Draw(screen *ebiten.Image) error {
	tx := l.X * scene.TileScale
	ty := l.Y * scene.TileScale
	if err := font.DrawText(screen, l.text, tx, ty, scene.TextScale, color.White); err != nil {
		return err
	}
	return nil
}