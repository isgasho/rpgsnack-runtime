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

package sceneimpl

import (
	"encoding/json"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type TitleScene struct {
	newGameButton        *ui.Button
	resumeGameButton     *ui.Button
	waitingLoadRequestID int
}

func NewTitleScene() *TitleScene {
	return &TitleScene{
		newGameButton:    ui.NewButton(0, 184, 120, 20, "New Game"),
		resumeGameButton: ui.NewButton(0, 208, 120, 20, "Resume Game"),
	}
}

func (t *TitleScene) Update(sceneManager *scene.Manager) error {
	if t.waitingLoadRequestID != 0 {
		if sceneManager.HasFinishedRequestID(t.waitingLoadRequestID) {
			sceneManager.FinishRequestID(t.waitingLoadRequestID)
			t.waitingLoadRequestID = 0
			data := sceneManager.LastLoadedData()
			var game *gamestate.Game
			if err := json.Unmarshal(data, &game); err != nil {
				return err
			}
			mapScene, err := NewMapSceneWithGame(game)
			if err != nil {
				return err
			}
			sceneManager.GoTo(mapScene)
		}
		return nil
	}
	w, _ := sceneManager.Size()
	t.newGameButton.X = (w/scene.TileScale - t.newGameButton.Width) / 2
	t.resumeGameButton.X = (w/scene.TileScale - t.resumeGameButton.Width) / 2
	if err := t.newGameButton.Update(); err != nil {
		return err
	}
	if err := t.resumeGameButton.Update(); err != nil {
		return err
	}
	if t.newGameButton.Pressed() {
		mapScene, err := NewMapScene()
		if err != nil {
			return err
		}
		sceneManager.GoTo(mapScene)
		return nil
	}
	if t.resumeGameButton.Pressed() {
		id := sceneManager.GenerateRequestID()
		t.waitingLoadRequestID = id
		sceneManager.Requester().RequestLoadProgress(id)
		return nil
	}
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) error {
	timg := assets.GetImage("title.png")
	tw, _ := timg.Size()
	sw, _ := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, 0)
	if err := screen.DrawImage(timg, op); err != nil {
		return err
	}
	if err := t.newGameButton.Draw(screen); err != nil {
		return err
	}
	if err := t.resumeGameButton.Draw(screen); err != nil {
		return err
	}
	return nil
}
