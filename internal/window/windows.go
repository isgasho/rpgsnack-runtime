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

package window

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

const (
	choiceBalloonHeight        = 20
	chosenBalloonWaitingFrames = 5
)

type Windows struct {
	nextBalloon               *balloon
	balloons                  []*balloon
	choiceBalloons            []*balloon
	chosenIndex               int
	choosing                  bool
	choosingInterpreterID     int
	chosenBalloonWaitingCount int
	hasChosenIndex            bool
}

func (w *Windows) ChosenIndex() int {
	return w.chosenIndex
}

func (w *Windows) HasChosenIndex() bool {
	return w.hasChosenIndex
}

func (w *Windows) ShowMessage(content string, eventID int, interpreterID int) {
	if w.nextBalloon != nil {
		panic("not reach")
	}
	// TODO: How to call newBalloonCenter?
	w.nextBalloon = newBalloonWithArrow(content, eventID, interpreterID)
}

func (w *Windows) ShowChoices(sceneManager *scene.Manager, choices []string, interpreterID int) {
	_, h := sceneManager.Size()
	ymin := h/scene.TileScale - len(choices)*choiceBalloonHeight
	w.choiceBalloons = nil
	for i, choice := range choices {
		x := sceneManager.MapOffsetX() / scene.TileScale
		y := i*choiceBalloonHeight + ymin
		width := scene.TileXNum * scene.TileSize
		balloon := newBalloon(x, y, width, choiceBalloonHeight, choice, interpreterID)
		w.choiceBalloons = append(w.choiceBalloons, balloon)
		balloon.open()
	}
	w.chosenIndex = 0
	w.choosing = true
	w.choosingInterpreterID = interpreterID
}

func (w *Windows) CloseAll() {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		b.close()
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		b.close()
	}
}

func (w *Windows) IsBusy(interpreterID int) bool {
	if w.isAnimating(interpreterID) {
		return true
	}
	if w.choosingInterpreterID == interpreterID {
		if w.choosing || w.chosenBalloonWaitingCount > 0 {
			return true
		}
	}
	if w.isOpened(interpreterID) {
		return true
	}
	if w.nextBalloon != nil {
		return true
	}
	return false
}

func (w *Windows) CanProceed(interpreterID int) bool {
	if !w.IsBusy(interpreterID) {
		return true
	}
	if !w.isOpened(interpreterID) {
		return false
	}
	if !input.Triggered() {
		return false
	}
	return true
}

func (w *Windows) isOpened(interpreterID int) bool {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isOpened() {
			return true
		}
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isOpened() {
			return true
		}
	}
	return false
}

func (w *Windows) isAnimating(interpreterID int) bool {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isAnimating() {
			return true
		}
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		if interpreterID > 0 && b.interpreterID != interpreterID {
			continue
		}
		if b.isAnimating() {
			return true
		}
	}
	return false
}

func (w *Windows) Update(sceneManager *scene.Manager) error {
	if !w.choosing {
		// 0 means to check all balloons.
		// TODO: Don't use magic numbers.
		if w.nextBalloon != nil && !w.isAnimating(0) && !w.isOpened(0) {
			w.balloons = []*balloon{w.nextBalloon}
			w.balloons[0].open()
			w.nextBalloon = nil
		}
	}
	if w.chosenBalloonWaitingCount > 0 {
		w.chosenBalloonWaitingCount--
		if w.chosenBalloonWaitingCount == 0 {
			w.choiceBalloons[w.chosenIndex].close()
			for _, b := range w.balloons {
				if b == nil {
					continue
				}
				b.close()
			}
			w.hasChosenIndex = false
		}
	} else if w.choosing && w.isOpened(0) && input.Triggered() {
		_, h := sceneManager.Size()
		ymax := h / scene.TileScale
		ymin := ymax - len(w.choiceBalloons)*choiceBalloonHeight
		_, y := input.Position()
		y /= scene.TileScale
		if y < ymin || ymax <= y {
			return nil
		}
		// Close regular balloons
		w.chosenIndex = (y - ymin) / choiceBalloonHeight
		for i, b := range w.choiceBalloons {
			if i == w.chosenIndex {
				continue
			}
			b.close()
		}
		w.chosenBalloonWaitingCount = chosenBalloonWaitingFrames
		w.choosing = false
		w.choosingInterpreterID = 0
		w.hasChosenIndex = true
	}
	for i, b := range w.balloons {
		if b == nil {
			continue
		}
		if err := b.update(); err != nil {
			return err
		}
		if b.isClosed() {
			w.balloons[i] = nil
		}
	}
	for i, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		if err := b.update(); err != nil {
			return err
		}
		if b.isClosed() {
			w.choiceBalloons[i] = nil
		}
	}
	return nil
}

func (w *Windows) Draw(screen *ebiten.Image, characters []*character.Character) error {
	for _, b := range w.balloons {
		if b == nil {
			continue
		}
		var c *character.Character
		for _, cc := range characters {
			if cc.EventID() == b.eventID {
				c = cc
				break
			}
		}
		if c == nil {
			return fmt.Errorf("windows: character (EventID=%d) not found", b.eventID)
		}
		if err := b.draw(screen, c); err != nil {
			return err
		}
	}
	for _, b := range w.choiceBalloons {
		if b == nil {
			continue
		}
		if err := b.draw(screen, nil); err != nil {
			return err
		}
	}
	return nil
}
