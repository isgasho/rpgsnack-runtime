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

package character

import (
	"github.com/hajimehoshi/ebiten"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
)

type Event struct {
	character *Character
}

func NewEvent(id int, x, y int) (*Event, error) {
	c := &Character{
		id:      id,
		speed:   data.Speed3,
		x:       x,
		y:       y,
		visible: true,
	}
	e := &Event{
		character: c,
	}
	return e, nil
}

func (e *Event) ID() int {
	return e.character.id
}

func (e *Event) Size() (int, int) {
	return e.character.Size()
}

func (e *Event) Position() (int, int) {
	return e.character.Position()
}

func (e *Event) DrawPosition() (int, int) {
	return e.character.DrawPosition()
}

func (e *Event) Dir() data.Dir {
	return e.character.dir
}

func (e *Event) IsMoving() bool {
	return e.character.IsMoving()
}

func (e *Event) Move(dir data.Dir) {
	e.character.Move(dir)
}

func (e *Event) Turn(dir data.Dir) {
	e.character.Turn(dir)
}

func (e *Event) SetSpeed(speed data.Speed) {
	e.character.speed = speed
}

func (e *Event) SetVisibility(visible bool) {
	e.character.visible = visible
}

func (e *Event) SetDirFix(dirFix bool) {
	e.character.dirFix = dirFix
}

func (e *Event) SetStepping(stepping bool) {
	e.character.stepping = stepping
}

func (e *Event) SetWalking(walking bool) {
	e.character.walking = walking
}

func (e *Event) SetImage(imageName string, imageIndex int, frame int, dir data.Dir, useFrameAndDir bool) {
	e.character.imageName = imageName
	e.character.imageIndex = imageIndex
	if useFrameAndDir {
		e.character.dir = dir
		e.character.frame = frame
		e.character.prevFrame = frame
	}
}

func (e *Event) UpdateCharacterIfNeeded(page *data.Page) error {
	if page == nil {
		c := e.character
		c.imageName = ""
		c.imageIndex = 0
		c.dirFix = false
		c.dir = data.Dir(0)
		c.frame = 1
		c.stepping = false
		return nil
	}
	c := e.character
	c.imageName = page.Image
	c.imageIndex = page.ImageIndex
	c.dirFix = page.DirFix
	c.dir = page.Dir
	c.frame = page.Frame
	c.stepping = page.Stepping
	return nil
}

func (e *Event) Update() error {
	if err := e.character.Update(); err != nil {
		return err
	}
	return nil
}

func (e *Event) Draw(screen *ebiten.Image) error {
	if err := e.character.Draw(screen); err != nil {
		return err
	}
	return nil
}
