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

package picture

import (
	"fmt"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/assets"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/interpolation"
)

type Pictures struct {
	pictures []*picture
	screen   *ebiten.Image
}

func (p *Pictures) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("pictures")
	e.BeginArray()
	for _, pic := range p.pictures {
		e.EncodeInterface(pic)
	}
	e.EndArray()

	e.EndMap()
	return e.Flush()
}

func (p *Pictures) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)

	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "pictures":
			if !d.SkipCodeIfNil() {
				n := d.DecodeArrayLen()
				p.pictures = make([]*picture, n)
				for i := 0; i < n; i++ {
					if !d.SkipCodeIfNil() {
						p.pictures[i] = &picture{}
						d.DecodeInterface(p.pictures[i])
					}
				}
			}
		}
	}

	if err := d.Error(); err != nil {
		return fmt.Errorf("pictures: Pictures.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (p *Pictures) MoveTo(id int, x, y int, count int) {
	p.pictures[id].moveTo(x, y, count)
}

func (p *Pictures) Scale(id int, scaleX, scaleY float64, count int) {
	p.pictures[id].scale(scaleX, scaleY, count)
}

func (p *Pictures) Update() {
	for _, pic := range p.pictures {
		if pic == nil {
			continue
		}
		pic.update()
	}
}

func (p *Pictures) Draw(screen *ebiten.Image) {
	if p.screen == nil {
		p.screen, _ = ebiten.NewImage(consts.TileSize*consts.TileXNum, consts.TileSize*consts.TileYNum, ebiten.FilterNearest)
	}
	p.screen.Clear()
	for _, pic := range p.pictures {
		if pic == nil {
			continue
		}
		pic.draw(p.screen)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(consts.TileScale, consts.TileScale)
	op.GeoM.Translate(0, consts.GameMarginTop)
	screen.DrawImage(p.screen, op)
}

func (p *Pictures) Add(id int, name string, x, y int, scaleX, scaleY, angle, opacity float64, origin data.ShowPictureOrigin, blendType data.ShowPictureBlendType) {
	if len(p.pictures) < id+1 {
		p.pictures = append(p.pictures, make([]*picture, id+1-len(p.pictures))...)
	}
	p.pictures[id] = &picture{
		imageName: name,
		image:     assets.GetImage("pictures/" + name + ".png"),
		x:         interpolation.New(float64(x)),
		y:         interpolation.New(float64(y)),
		scaleX:    interpolation.New(scaleX),
		scaleY:    interpolation.New(scaleX),
		angle:     interpolation.New(angle),
		opacity:   interpolation.New(opacity),
		origin:    origin,
		blendType: blendType,
	}
}

func (p *Pictures) Remove(id int) {
	p.pictures[id] = nil
}

type picture struct {
	imageName string
	image     *ebiten.Image
	x         *interpolation.I
	y         *interpolation.I
	scaleX    *interpolation.I
	scaleY    *interpolation.I
	angle     *interpolation.I
	opacity   *interpolation.I
	origin    data.ShowPictureOrigin
	blendType data.ShowPictureBlendType
}

func (p *picture) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("imageName")
	e.EncodeString(p.imageName)

	e.EncodeString("x")
	e.EncodeInterface(p.x)

	e.EncodeString("y")
	e.EncodeInterface(p.y)

	e.EncodeString("scaleX")
	e.EncodeInterface(p.scaleX)

	e.EncodeString("scaleY")
	e.EncodeInterface(p.scaleY)

	e.EncodeString("angle")
	e.EncodeInterface(p.angle)

	e.EncodeString("opacity")
	e.EncodeInterface(p.opacity)

	e.EncodeString("origin")
	e.EncodeString(string(p.origin))

	e.EncodeString("blendType")
	e.EncodeString(string(p.blendType))

	e.EndMap()
	return e.Flush()
}

func (p *picture) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)

	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		switch k := d.DecodeString(); k {
		case "imageName":
			p.imageName = d.DecodeString()
			p.image = assets.GetImage("pictures/" + p.imageName + ".png")
		case "x":
			p.x = &interpolation.I{}
			d.DecodeInterface(p.x)
		case "y":
			p.y = &interpolation.I{}
			d.DecodeInterface(p.y)
		case "scaleX":
			p.scaleX = &interpolation.I{}
			d.DecodeInterface(p.scaleX)
		case "scaleY":
			p.scaleY = &interpolation.I{}
			d.DecodeInterface(p.scaleY)
		case "angle":
			p.angle = &interpolation.I{}
			d.DecodeInterface(p.angle)
		case "opacity":
			p.opacity = &interpolation.I{}
			d.DecodeInterface(p.opacity)
		case "origin":
			p.origin = data.ShowPictureOrigin(d.DecodeString())
		case "blendType":
			p.blendType = data.ShowPictureBlendType(d.DecodeString())
		}
	}

	if err := d.Error(); err != nil {
		return fmt.Errorf("pictures: picture.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (p *picture) moveTo(x, y int, count int) {
	p.x.Set(float64(x), count)
	p.y.Set(float64(y), count)
}

func (p *picture) scale(scaleX, scaleY float64, count int) {
	p.scaleX.Set(scaleX, count)
	p.scaleY.Set(scaleY, count)
}

func (p *picture) update() {
	p.x.Update()
	p.y.Update()
	p.scaleX.Update()
	p.scaleY.Update()
	p.angle.Update()
	p.opacity.Update()
}

func (p *picture) draw(screen *ebiten.Image) {
	sx, sy := p.image.Size()

	op := &ebiten.DrawImageOptions{}
	if p.origin == data.ShowPictureOriginCenter {
		op.GeoM.Translate(-float64(sx)/2, -float64(sy)/2)
	}
	op.GeoM.Scale(p.scaleX.Current(), p.scaleY.Current())
	op.GeoM.Rotate(p.angle.Current())
	op.GeoM.Translate(p.x.Current(), p.y.Current())

	if p.opacity.Current() < 1 {
		op.ColorM.Scale(1, 1, 1, p.opacity.Current())
	}
	switch p.blendType {
	case data.ShowPictureBlendTypeNormal:
		// Use default
	case data.ShowPictureBlendTypeAdd:
		op.CompositeMode = ebiten.CompositeModeLighter
	}

	screen.DrawImage(p.image, op)
}
