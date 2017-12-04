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
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/font"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/texts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/ui"
)

type MapScene struct {
	gameState          *gamestate.Game
	moveDstX           int
	moveDstY           int
	tilesImage         *ebiten.Image
	uiImage            *ebiten.Image
	triggeringFailed   bool
	initialState       bool
	cameraButton       *ui.Button
	cameraTaking       bool
	titleButton        *ui.Button
	screenShotImage    *ebiten.Image
	screenShotDialog   *ui.Dialog
	quitDialog         *ui.Dialog
	quitLabel          *ui.Label
	quitYesButton      *ui.Button
	quitNoButton       *ui.Button
	storeErrorDialog   *ui.Dialog
	storeErrorLabel    *ui.Label
	storeErrorOkButton *ui.Button
	removeAdsButton    *ui.Button
	removeAdsDialog    *ui.Dialog
	removeAdsLabel     *ui.Label
	removeAdsYesButton *ui.Button
	removeAdsNoButton  *ui.Button
	inventory          *ui.Inventory
	itemPreviewPopup   *ui.ItemPreviewPopup
	waitingRequestID   int
	isAdsRemoved       bool
	initialized        bool
	offsetX            int
	offsetY            int
}

func NewMapScene() *MapScene {
	m := &MapScene{
		gameState:    gamestate.NewGame(),
		initialState: true,
	}
	return m
}

func NewMapSceneWithGame(game *gamestate.Game) *MapScene {
	m := &MapScene{
		gameState: game,
	}
	return m
}

func (m *MapScene) initUI(sceneManager *scene.Manager) {
	const (
		inventoryHeight = 33 * consts.TileScale
		uiWidth         = consts.TileXNum * consts.TileSize
	)

	screenW, screenH := sceneManager.Size()
	m.offsetX = (screenW - consts.TileXNum*consts.TileSize*consts.TileScale) / 2
	m.offsetY = screenH - consts.TileYNum*consts.TileSize*consts.TileScale - inventoryHeight
	// offset y should be multiplies of TileScale for pixel-pefect rendering
	m.offsetY -= m.offsetY % consts.TileScale

	// TODO: Rename tilesImage to screenImage, and create another tilesImage that doesn't consider offsets
	m.tilesImage, _ = ebiten.NewImage(consts.TileXNum*consts.TileSize, screenH/consts.TileScale, ebiten.FilterNearest)
	m.uiImage, _ = ebiten.NewImage(uiWidth*consts.TileScale, screenH, ebiten.FilterNearest)

	screenShotImage, _ := ebiten.NewImage(480, 720, ebiten.FilterLinear)
	camera, _ := ebiten.NewImage(12, 12, ebiten.FilterNearest)
	camera.Fill(color.RGBA{0xff, 0, 0, 0xff})
	cameraImagePart := ui.NewImagePart(camera)
	m.cameraButton = ui.NewImageButton(0, 0, cameraImagePart, cameraImagePart, "click")
	m.screenShotImage = screenShotImage
	m.screenShotDialog = ui.NewDialog((uiWidth-160)/2+4, 4, 152, 232)
	m.screenShotDialog.AddChild(ui.NewImageView(8, 8, 1.0/consts.TileScale/2, ui.NewImagePart(m.screenShotImage)))
	m.titleButton = ui.NewButton(4, 2, 40, 12, "click")

	// TODO: Implement the camera functionality later
	m.cameraButton.Visible = false

	m.quitDialog = ui.NewDialog((uiWidth-160)/2+4, 64, 152, 124)
	m.quitLabel = ui.NewLabel(16, 8)
	m.quitYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "click")
	m.quitNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "cancel")

	m.quitDialog.AddChild(m.quitLabel)
	m.quitDialog.AddChild(m.quitYesButton)
	m.quitDialog.AddChild(m.quitNoButton)

	m.storeErrorDialog = ui.NewDialog((uiWidth-160)/2+4, 64, 152, 124)
	m.storeErrorLabel = ui.NewLabel(16, 8)
	m.storeErrorOkButton = ui.NewButton((152-120)/2, 96, 120, 20, "click")
	m.storeErrorDialog.AddChild(m.storeErrorLabel)
	m.storeErrorDialog.AddChild(m.storeErrorOkButton)

	m.removeAdsButton = ui.NewButton(104, 8, 52, 12, "click")
	m.removeAdsDialog = ui.NewDialog((uiWidth-160)/2+4, 64, 152, 124)
	m.removeAdsLabel = ui.NewLabel(16, 8)
	m.removeAdsYesButton = ui.NewButton((152-120)/2, 72, 120, 20, "click")
	m.removeAdsNoButton = ui.NewButton((152-120)/2, 96, 120, 20, "cancel")
	m.removeAdsDialog.AddChild(m.removeAdsLabel)
	m.removeAdsDialog.AddChild(m.removeAdsYesButton)
	m.removeAdsDialog.AddChild(m.removeAdsNoButton)

	m.inventory = ui.NewInventory(0, (screenH-inventoryHeight)/consts.TileScale)
	m.itemPreviewPopup = ui.NewItemPreviewPopup((uiWidth-160)/2+16, m.offsetY/consts.TileScale)
	m.quitDialog.AddChild(m.quitLabel)

	m.removeAdsButton.Visible = false // TODO: Clock of Atonement does not need this feature, so turn it off for now

	m.quitYesButton.SetOnPressed(func(_ *ui.Button) {
		if m.gameState.IsAutoSaveEnabled() {
			m.gameState.RequestSave(sceneManager)
		}
		audio.Stop()
		sceneManager.GoToWithFading(NewTitleScene(), 30)
	})
	m.quitNoButton.SetOnPressed(func(_ *ui.Button) {
		m.quitDialog.Hide()
	})
	m.titleButton.SetOnPressed(func(_ *ui.Button) {
		m.quitDialog.Show()
	})
	m.storeErrorOkButton.SetOnPressed(func(_ *ui.Button) {
		m.storeErrorDialog.Hide()
	})
	m.removeAdsButton.SetOnPressed(func(_ *ui.Button) {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestGetIAPPrices(m.waitingRequestID)
	})
	m.removeAdsYesButton.SetOnPressed(func(_ *ui.Button) {
		m.waitingRequestID = sceneManager.GenerateRequestID()
		sceneManager.Requester().RequestPurchase(m.waitingRequestID, "ads_removal")
	})
	m.removeAdsNoButton.SetOnPressed(func(_ *ui.Button) {
		m.removeAdsDialog.Hide()
	})
	m.cameraButton.SetOnPressed(func(_ *ui.Button) {
		m.cameraTaking = true
		m.screenShotDialog.Show()
	})

	m.inventory.SetOnSlotPressed(func(_ *ui.Inventory, index int) {
		if index < m.gameState.Items().ItemNum() {
			activeItemID := m.gameState.Items().ActiveItem()
			itemID := m.gameState.Items().ItemIDAt(index)
			switch m.inventory.Mode() {
			case ui.DefaultMode:
				if itemID == m.gameState.Items().ActiveItem() {
					m.gameState.Items().Deactivate()
				} else {
					m.gameState.Items().Activate(itemID)
				}
			case ui.PreviewMode:
				var combineItem *data.Item
				for _, item := range sceneManager.Game().Items {
					if m.inventory.CombineItemID() == item.ID {
						combineItem = item
						break
					}
				}

				m.itemPreviewPopup.SetCombineItem(combineItem, sceneManager.Game().CreateCombine(activeItemID, m.inventory.CombineItemID()))
			default:
				panic("not reached")
			}
		}
	})
	m.inventory.SetOnActiveItemPressed(func(_ *ui.Inventory) {
		m.gameState.Items().SetEventItem(m.gameState.Items().ActiveItem())
		m.updateItemPreviewPopupVisibility(sceneManager)
	})
	m.itemPreviewPopup.SetOnClosePressed(func(_ *ui.ItemPreviewPopup) {
		m.gameState.Items().SetEventItem(0)
		m.updateItemPreviewPopupVisibility(sceneManager)
	})
	m.inventory.SetOnBackPressed(func(_ *ui.Inventory) {
		m.gameState.Items().SetEventItem(0)
		m.updateItemPreviewPopupVisibility(sceneManager)
	})
	// TODO: ItemPreviewPopup is not standarized as the other Popups
	m.itemPreviewPopup.SetOnActionPressed(func(_ *ui.ItemPreviewPopup) {
		if m.gameState.ExecutingItemCommands() {
			return
		}
		if m.gameState.Map().IsBlockingEventExecuting() {
			return
		}
		activeItemID := m.gameState.Items().ActiveItem()
		if m.inventory.CombineItemID() != 0 {
			combine := sceneManager.Game().CreateCombine(activeItemID, m.inventory.CombineItemID())
			m.gameState.StartCombineCommands(combine)
		} else {
			m.gameState.StartItemCommands(activeItemID)
		}
	})
}

func (m *MapScene) updateItemPreviewPopupVisibility(sceneManager *scene.Manager) {
	itemID := m.gameState.Items().EventItem()
	if itemID > 0 {
		m.inventory.SetActiveItemID(itemID)
		m.inventory.SetMode(ui.PreviewMode)
		var eventItem *data.Item
		for _, item := range sceneManager.Game().Items {
			if item.ID == itemID {
				eventItem = item
				break
			}
		}

		m.itemPreviewPopup.SetActiveItem(eventItem)
		m.itemPreviewPopup.Show()
	} else {
		m.gameState.Items().SetEventItem(0)
		m.itemPreviewPopup.SetActiveItem(nil)
		m.itemPreviewPopup.Hide()
		m.inventory.SetMode(ui.DefaultMode)
	}
}

func (m *MapScene) updatePurchasesState(sceneManager *scene.Manager) {
	m.isAdsRemoved = sceneManager.IsPurchased("ads_removal")
}

func (m *MapScene) runEventIfNeeded(sceneManager *scene.Manager) {
	if m.itemPreviewPopup.Visible() {
		m.triggeringFailed = false
		return
	}
	if m.gameState.Map().IsBlockingEventExecuting() {
		m.triggeringFailed = false
		return
	}
	if !input.Triggered() {
		return
	}
	x, y := input.Position()
	x -= m.offsetX
	y -= m.offsetY
	if x < 0 || y < 0 {
		return
	}
	tx := x / consts.TileSize / consts.TileScale
	ty := y / consts.TileSize / consts.TileScale
	if tx < 0 || consts.TileXNum <= tx || ty < 0 || consts.TileYNum <= ty {
		return
	}
	m.moveDstX = tx
	m.moveDstY = ty
	if m.gameState.Map().TryRunDirectEvent(m.gameState, tx, ty) {
		m.triggeringFailed = false
		return
	}
	if !m.gameState.Map().TryMovePlayerByUserInput(sceneManager, m.gameState, tx, ty) {
		m.triggeringFailed = true
		return
	}
	m.triggeringFailed = false
}

func (m *MapScene) receiveRequest(sceneManager *scene.Manager) bool {
	if m.waitingRequestID == 0 {
		return true
	}

	r := sceneManager.ReceiveResultIfExists(m.waitingRequestID)
	if r == nil {
		return false
	}
	m.waitingRequestID = 0
	switch r.Type {
	case scene.RequestTypeIAPPrices:
		if !r.Succeeded {
			m.storeErrorDialog.Show()
			break
		}
		priceText := "???"
		var prices map[string]string
		if err := json.Unmarshal(r.Data, &prices); err != nil {
			panic(err)
		}
		text := texts.Text(sceneManager.Language(), texts.TextIDRemoveAdsDesc)
		if _, ok := prices["ads_removal"]; ok {
			priceText = prices["ads_removal"]
		}
		m.removeAdsLabel.Text = fmt.Sprintf(text, priceText)
		m.removeAdsDialog.Show()
	case scene.RequestTypePurchase:
		// Note: Ideally we should show a notification toast to notify users about the result
		// For now, the notifications are handled on the native platform side
		if r.Succeeded {
			m.updatePurchasesState(sceneManager)
		}
		m.removeAdsDialog.Hide()
	}
	return false
}

func (m *MapScene) isUIBusy() bool {
	if m.quitDialog.Visible() {
		return true
	}
	if m.storeErrorDialog.Visible() {
		return true
	}
	if m.removeAdsDialog.Visible() {
		return true
	}
	if m.screenShotDialog.Visible() {
		return true
	}
	return false
}

func (m *MapScene) updateUI(sceneManager *scene.Manager) {
	l := sceneManager.Language()
	m.quitLabel.Text = texts.Text(l, texts.TextIDBackToTitle)
	m.quitYesButton.Text = texts.Text(l, texts.TextIDYes)
	m.quitNoButton.Text = texts.Text(l, texts.TextIDNo)
	m.storeErrorLabel.Text = texts.Text(l, texts.TextIDStoreError)
	m.storeErrorOkButton.Text = texts.Text(l, texts.TextIDOK)
	m.removeAdsYesButton.Text = texts.Text(l, texts.TextIDYes)
	m.removeAdsNoButton.Text = texts.Text(l, texts.TextIDNo)
	m.titleButton.Text = texts.Text(l, texts.TextIDTitle)
	m.removeAdsButton.Text = texts.Text(l, texts.TextIDRemoveAds)

	// Call SetOffset as temporary hack for UI input
	input.SetOffset(m.offsetX, 0)
	defer input.SetOffset(0, 0)

	m.quitDialog.Update()
	m.screenShotDialog.Update()
	m.storeErrorDialog.Update()
	m.removeAdsDialog.Update()

	m.cameraButton.Update()
	m.titleButton.Disabled = m.gameState.Map().IsBlockingEventExecuting()
	m.titleButton.Update()

	m.removeAdsButton.Disabled = m.gameState.Map().IsBlockingEventExecuting()
	m.removeAdsButton.Update()

	if m.gameState.InventoryVisible() {
		m.inventory.Show()
	} else {
		m.inventory.Hide()
	}
	m.inventory.SetDisabled(m.gameState.Map().IsBlockingEventExecuting())
	m.inventory.SetItems(m.gameState.Items().Items(sceneManager.Game().Items))
	m.inventory.SetActiveItemID(m.gameState.Items().ActiveItem())
	m.inventory.Update()
	m.itemPreviewPopup.Update(l)

	// Event handling
	m.updateItemPreviewPopupVisibility(sceneManager)
}

func (m *MapScene) Update(sceneManager *scene.Manager) error {
	if !m.initialized {
		m.initUI(sceneManager)
		m.initialized = true
	}

	m.updatePurchasesState(sceneManager)

	if ok := m.receiveRequest(sceneManager); !ok {
		return nil
	}

	if input.BackButtonPressed() {
		m.handleBackButton()
	}

	m.updateUI(sceneManager)
	if m.isUIBusy() {
		return nil
	}

	if m.initialState && m.gameState.IsAutoSaveEnabled() {
		m.gameState.RequestSave(sceneManager)
	}
	m.initialState = false
	if err := m.gameState.Update(sceneManager); err != nil {
		if err == gamestate.GoToTitle {
			m.goToTitle(sceneManager)
			return nil
		}
		return err
	}
	m.runEventIfNeeded(sceneManager)
	return nil
}

func (m *MapScene) goToTitle(sceneManager *scene.Manager) {
	audio.Stop()
	sceneManager.GoToWithFading(NewTitleScene(), 60)
}

func (m *MapScene) handleBackButton() {
	if m.storeErrorDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		m.storeErrorDialog.Hide()
		return
	}

	if m.quitDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		m.quitDialog.Hide()
		return
	}
	if m.quitDialog.Visible() {
		audio.PlaySE("cancel", 1.0)
		m.quitDialog.Hide()
		return
	}

	audio.PlaySE("click", 1.0)
	m.quitDialog.Show()
}

func (m *MapScene) Draw(screen *ebiten.Image) {
	if !m.initialized {
		return
	}
	m.tilesImage.Fill(color.Black)

	// TODO: This accesses *data.Game, but is it OK?
	room := m.gameState.Map().CurrentRoom()

	if room.Background.Name != "" {
		m.gameState.Map().DrawFullscreenImage(m.tilesImage, assets.GetImage("backgrounds/"+room.Background.Name+".png"), 0, m.offsetY/consts.TileScale)
	}
	op := &ebiten.DrawImageOptions{}
	for k := 0; k < 3; k++ {
		layer := 0
		if k >= 1 {
			layer = 1
		}
		if tileSet := m.gameState.Map().TileSet(layer); tileSet != nil {
			tileSetImg := assets.GetImage("tilesets/" + tileSet.Name + ".png")
			for j := 0; j < consts.TileYNum; j++ {
				for i := 0; i < consts.TileXNum; i++ {
					tile := room.Tiles[layer][j*consts.TileXNum+i]
					if layer == 1 {
						p := tileSet.PassageTypes[tile]
						if k == 1 && p == data.PassageTypeOver {
							continue
						}
						if k == 2 && p != data.PassageTypeOver {
							continue
						}
					}
					sx := tile % consts.PaletteWidth * consts.TileSize
					sy := tile / consts.PaletteWidth * consts.TileSize
					r := image.Rect(sx, sy, sx+consts.TileSize, sy+consts.TileSize)
					op.SourceRect = &r
					dx := i * consts.TileSize
					dy := j*consts.TileSize + m.offsetY/consts.TileScale
					op.GeoM.Reset()
					op.GeoM.Translate(float64(dx), float64(dy))
					m.tilesImage.DrawImage(tileSetImg, op)
				}
			}
		}
		var p data.Priority
		switch k {
		case 0:
			p = data.PriorityBottom
		case 1:
			p = data.PriorityMiddle
		case 2:
			p = data.PriorityTop
		default:
			panic("not reached")
		}
		m.gameState.Map().DrawCharacters(m.tilesImage, p, 0, m.offsetY/consts.TileScale)
	}
	if room.Foreground.Name != "" {
		m.gameState.Map().DrawFullscreenImage(m.tilesImage, assets.GetImage("foregrounds/"+room.Foreground.Name+".png"), 0, m.offsetY/consts.TileScale)
	}

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(float64(m.offsetX), 0)
	m.gameState.DrawScreen(screen, m.tilesImage, op)
	m.gameState.DrawPictures(screen, m.offsetX, m.offsetY)

	if m.gameState.IsPlayerControlEnabled() && (m.gameState.Map().IsPlayerMovingByUserInput() || m.triggeringFailed) {
		x, y := m.moveDstX, m.moveDstY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*consts.TileSize), float64(y*consts.TileSize))
		op.GeoM.Scale(consts.TileScale, consts.TileScale)
		op.GeoM.Translate(float64(m.offsetX), float64(m.offsetY))
		screen.DrawImage(assets.GetImage("system/marker.png"), op)
	}

	m.uiImage.Clear()

	m.itemPreviewPopup.Draw(m.uiImage)
	m.inventory.Draw(m.uiImage)

	m.cameraButton.Draw(m.uiImage)
	m.titleButton.Draw(m.uiImage)
	m.removeAdsButton.Draw(m.uiImage)

	m.screenShotDialog.Draw(m.uiImage)
	m.quitDialog.Draw(m.uiImage)
	m.storeErrorDialog.Draw(m.uiImage)
	m.removeAdsDialog.Draw(m.uiImage)

	m.gameState.DrawWindows(m.uiImage, 0, m.offsetY)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(m.offsetX), 0)
	screen.DrawImage(m.uiImage, op)

	if m.cameraTaking {
		m.cameraTaking = false
		m.screenShotImage.Clear()
		op := &ebiten.DrawImageOptions{}
		sw, _ := screen.Size()
		w, _ := m.screenShotImage.Size()
		op.GeoM.Translate((float64(w)-float64(sw))/2, 0)
		m.screenShotImage.DrawImage(m.uiImage, nil)
	}

	msg := fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS())
	font.DrawText(screen, msg, 160+m.offsetX, 8, consts.TextScale, data.TextAlignLeft, color.White)
}
