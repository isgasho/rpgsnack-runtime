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
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type TitleScene struct {
	newGameButton    *ui.Button
	resumeGameButton *ui.Button
	settingsButton   *ui.Button
	warningDialog    *ui.Dialog
	warningLabel     *ui.Label
	warningYesButton *ui.Button
	warningNoButton  *ui.Button
}

func NewTitleScene() *TitleScene {
	const warning = `You have a on-going game data.
Do you want to clear the progress
to start a new game?`
	t := &TitleScene{
		newGameButton:    ui.NewButton(0, 184, 120, 20, "New Game"),
		resumeGameButton: ui.NewButton(0, 208, 120, 20, "Resume Game"),
		settingsButton:   ui.NewImageButton(0, 0, assets.GetImage("icon_settings.png")),
		warningDialog:    ui.NewDialog(0, 4, 152, 232),
		warningYesButton: ui.NewButton(0, 180, 120, 20, "Yes"),
		warningNoButton:  ui.NewButton(0, 204, 120, 20, "No"),
	}
	t.warningDialog.AddChild(ui.NewLabel(8, 8, warning))
	t.warningDialog.AddChild(t.warningYesButton)
	t.warningDialog.AddChild(t.warningNoButton)
	return t
}

func (t *TitleScene) Update(sceneManager *scene.Manager) error {
	w, h := sceneManager.Size()
	t.newGameButton.X = (w/scene.TileScale - t.newGameButton.Width) / 2
	t.resumeGameButton.X = (w/scene.TileScale - t.resumeGameButton.Width) / 2
	t.settingsButton.X = w/scene.TileScale - 16
	t.settingsButton.Y = h/scene.TileScale - 16
	t.warningDialog.X = (w/scene.TileScale-160)/2 + 4
	t.warningYesButton.X = (t.warningDialog.Width - t.warningYesButton.Width) / 2
	t.warningNoButton.X = (t.warningDialog.Width - t.warningNoButton.Width) / 2
	t.warningDialog.Update()
	if t.warningYesButton.Pressed() {
		mapScene, err := NewMapScene()
		if err != nil {
			return err
		}
		sceneManager.GoTo(mapScene)
		return nil
	}
	if t.warningNoButton.Pressed() {
		t.warningDialog.Visible = false
		return nil
	}
	if t.warningDialog.Visible {
		return nil
	}
	t.resumeGameButton.Visible = data.Progress() != nil
	t.newGameButton.Update()
	t.resumeGameButton.Update()
	t.settingsButton.Update()
	if t.newGameButton.Pressed() {
		if data.Progress() != nil {
			t.warningDialog.Visible = true
		} else {
			mapScene, err := NewMapScene()
			if err != nil {
				return err
			}
			sceneManager.GoTo(mapScene)
		}
		return nil
	}
	if data.Progress() != nil && t.resumeGameButton.Pressed() {
		var game *gamestate.Game
		if err := json.Unmarshal(data.Progress(), &game); err != nil {
			return err
		}
		mapScene, err := NewMapSceneWithGame(game)
		if err != nil {
			return err
		}
		sceneManager.GoTo(mapScene)
		return nil
	}
	if t.settingsButton.Pressed() {
		sceneManager.GoTo(NewSettingsScene())
		return nil
	}
	return nil
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	timg := assets.GetImage("title.png")
	tw, _ := timg.Size()
	sw, _ := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, 0)
	screen.DrawImage(timg, op)
	t.newGameButton.Draw(screen)
	t.resumeGameButton.Draw(screen)
	t.settingsButton.Draw(screen)
	t.warningDialog.Draw(screen)
}
