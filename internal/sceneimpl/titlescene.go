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
	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"
	"image/color"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type TitleScene struct {
	init             bool
	startGameButton  *ui.Button
	removeAdsButton  *ui.Button
	settingsButton   *ui.Button
	moregamesButton  *ui.Button
	quitDialog       *ui.Dialog
	quitLabel        *ui.Label
	quitYesButton    *ui.Button
	quitNoButton     *ui.Button
	waitingRequestID int
	initialized      bool
	animation        animation
	bgImage          *ebiten.Image
	err              error
}

const (
	footerHeight = 280
)

func NewTitleScene() *TitleScene {
	t := &TitleScene{}
	return t
}

func (t *TitleScene) initUI(sceneManager *scene.Manager) {
	w, h := sceneManager.Size()

	settingsIcon := assets.GetImage("system/common/icon_settings.png")
	moreGamesIcon := assets.GetImage("system/common/icon_moregames.png")

	by := 16
	if sceneManager.HasExtraBottomGrid() {
		by = 36
	}
	t.startGameButton = ui.NewTextButton((w/consts.TileScale-120)/2, h/consts.TileScale-by-32, 120, 20, "system/start")
	t.removeAdsButton = ui.NewTextButton((w/consts.TileScale-120)/2+20, h/consts.TileScale-by-8, 80, 20, "system/click")
	t.removeAdsButton.TextColor = color.RGBA{0xc8, 0xc8, 0xc8, 0xff}
	t.settingsButton = ui.NewImageButton(w/consts.TileScale-24, h/consts.TileScale-by, settingsIcon, settingsIcon, "system/click")
	t.settingsButton.TouchExpand = 10
	t.moregamesButton = ui.NewImageButton(12, h/consts.TileScale-by, moreGamesIcon, moreGamesIcon, "system/click")
	t.moregamesButton.TouchExpand = 10

	t.quitDialog = ui.NewDialog((w/consts.TileScale-160)/2+4, (h)/(2*consts.TileScale)-64, 152, 124)
	t.quitLabel = ui.NewLabel(16, 8)
	t.quitYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "system/click")
	t.quitNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "system/cancel")
	t.quitDialog.AddChild(t.quitLabel)
	t.quitDialog.AddChild(t.quitYesButton)
	t.quitDialog.AddChild(t.quitNoButton)

	t.quitYesButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.Requester().RequestTerminateGame()
	})
	t.quitNoButton.SetOnPressed(func(_ *ui.Button) {
		t.quitDialog.Hide()
	})
	t.startGameButton.SetOnPressed(func(_ *ui.Button) {
		if sceneManager.HasProgress() {
			var game *gamestate.Game
			if err := msgpack.Unmarshal(sceneManager.Progress(), &game); err != nil {
				t.err = err
				return
			}
			audio.StopBGM(0)
			sceneManager.GoToWithFading(NewMapSceneWithGame(game), 60)
		} else {
			audio.StopBGM(0)
			sceneManager.GoToWithFading(NewMapScene(), 60)
		}
	})
	t.removeAdsButton.SetOnPressed(func(_ *ui.Button) {
		if sceneManager.Game().IsShopAvailable(data.ShopTypeHome) {
			sceneManager.Requester().RequestShowShop(t.waitingRequestID, string(sceneManager.ShopProductsDataByShop(data.ShopTypeHome)))
		}
	})
	t.settingsButton.SetOnPressed(func(_ *ui.Button) {
		sceneManager.GoTo(NewSettingsScene())
	})
	t.moregamesButton.SetOnPressed(func(_ *ui.Button) {
		t.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestOpenLink(t.waitingRequestID, "more", "")
	})
}

func (t *TitleScene) Update(sceneManager *scene.Manager) error {
	t.animation.Update()

	if t.err != nil {
		return t.err
	}
	if !t.initialized {
		t.initUI(sceneManager)
		t.initialized = true
	}
	if t.waitingRequestID != 0 {
		r := sceneManager.ReceiveResultIfExists(t.waitingRequestID)
		if r != nil {
			t.waitingRequestID = 0
		}
		return nil
	}

	if !t.init {
		var titleBGM = sceneManager.Game().System.TitleBGM
		if titleBGM.Name == "" {
			audio.StopBGM(0)
		} else {
			audio.PlayBGM(titleBGM.Name, float64(titleBGM.Volume)/100, 0)
		}
		t.init = true
	}

	if input.BackButtonPressed() {
		t.handleBackButton()
	}

	if sceneManager.HasProgress() {
		t.startGameButton.Text = texts.Text(lang.Get(), texts.TextIDResumeGame)
	} else {
		t.startGameButton.Text = texts.Text(lang.Get(), texts.TextIDNewGame)
	}
	if sceneManager.Game().System.TitleTextColor == "black" {
		t.startGameButton.TextColor = color.Black
	} else {
		t.startGameButton.TextColor = color.White
	}

	t.removeAdsButton.Text = texts.Text(lang.Get(), texts.TextIDRemoveAds)
	t.quitLabel.Text = texts.Text(lang.Get(), texts.TextIDQuitGame)
	t.quitYesButton.Text = texts.Text(lang.Get(), texts.TextIDYes)
	t.quitNoButton.Text = texts.Text(lang.Get(), texts.TextIDNo)

	t.quitDialog.Update()
	if !t.quitDialog.Visible() {
		t.startGameButton.Update()
		t.removeAdsButton.Update()
		t.settingsButton.Update()
		t.moregamesButton.Update()
	}

	t.removeAdsButton.Visible = sceneManager.IsAdsRemovable() && !sceneManager.IsAdsRemoved()

	return nil
}

func (t *TitleScene) handleBackButton() {
	if t.quitDialog.Visible() {
		audio.PlaySE("system/cancel", 1.0)
		t.quitDialog.Hide()
		return
	}

	audio.PlaySE("system/click", 1.0)
	t.quitDialog.Show()
}

func (t *TitleScene) DrawFooter(screen *ebiten.Image) {
	fimg := assets.GetImage("system/common/title_footer.png")
	_, sh := screen.Size()

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(sh-footerHeight))
	screen.DrawImage(fimg, op)
}

func (t *TitleScene) DrawBackgroundAnimation(screen *ebiten.Image) {
	_, sh := screen.Size()
	if assets.ImageExists("titles/bg.png") {
		const (
			frameWidth  = 160
			frameHeight = 280
		)

		if t.bgImage == nil {
			t.bgImage, _ = ebiten.NewImage(frameWidth, frameHeight, ebiten.FilterDefault)
		}
		t.bgImage.Clear()
		// We would like to focus on the 1/3 point of the title image
		// This allows it to show the title "nicely" on any device
		ty := ((sh-footerHeight)/consts.TileScale - frameHeight) / 3
		t.animation.Draw(t.bgImage, assets.GetImage("titles/bg.png"), frameWidth, 0, ty)

		op := &ebiten.DrawImageOptions{}
		sw, _ := screen.Size()
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(float64(sw-frameWidth*consts.TileScale)/2, 0)
		screen.DrawImage(t.bgImage, op)
	}
}

func (t *TitleScene) DrawTitle(screen *ebiten.Image) {
	// TODO: titles/title is used in games before 'title as map' was introduced.
	// Remove this usage in the future.
	if !assets.Exists("titles/title") {
		return
	}
	timg := assets.GetLocalizeImage("titles/title")
	tw, th := timg.Size()
	sw, sh := screen.Size()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate((float64(sw)-float64(tw))/2, float64(sh-footerHeight-th)/2)
	screen.DrawImage(timg, op)
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	if !t.initialized {
		return
	}

	t.DrawBackgroundAnimation(screen)
	t.DrawFooter(screen)
	t.DrawTitle(screen)

	// TODO: hide buttons to avoid visual conflicts between the dialog and the buttons
	if !t.quitDialog.Visible() {
		t.startGameButton.Draw(screen)
		t.removeAdsButton.Draw(screen)
		t.settingsButton.Draw(screen)
		t.moregamesButton.Draw(screen)
	}
	t.quitDialog.Draw(screen)
}

func (t *TitleScene) Resize() {
	t.initialized = false
}
