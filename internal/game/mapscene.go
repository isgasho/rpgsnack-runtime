// Copyright 2016 Hajime Hoshi
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

package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/tsugunai/internal/assets"
	"github.com/hajimehoshi/tsugunai/internal/font"
)

type mapScene struct {
	tilesImage *ebiten.Image
}

func newMapScene() (*mapScene, error) {
	// TODO: The image should be loaded asyncly.
	tilesImage, err := assets.LoadImage("tiles.png", ebiten.FilterNearest)
	if err != nil {
		return nil, err
	}
	return &mapScene{
		tilesImage: tilesImage,
	}, nil
}

func (m *mapScene) Update(sceneManager *sceneManager) error {
	return nil
}

func (m *mapScene) Draw(screen *ebiten.Image) error {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileScale, tileScale)
	if err := screen.DrawImage(m.tilesImage, op); err != nil {
		return err
	}
	if err := font.DrawText(screen, "文字の大きさはこれくらい。", 0, 0, textScale, color.White); err != nil {
		return err
	}
	return nil
}
