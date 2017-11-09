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

package data

type CombineType string

const (
	CombineTypeUse     CombineType = "use"
	CombineTypeCombine CombineType = "combine"
)

type Game struct {
	Maps         []*Map         `json:"maps"`
	Texts        *Texts         `json:"texts"`
	TileSets     []*TileSet     `json:"tileSets"`
	Achievements []*Achievement `json:"achievements"`
	Hints        []*Hint        `json:"hints"`
	IAPProducts  []*IAPProduct  `json:"iapProducts"`
	Items        []*Item        `json:"items"`
	Combines     []*Combine     `json:"combines"`
	CommonEvents []*CommonEvent `json:"commonEvents"`
	System       *System        `json:"system"`
}

type BGM struct {
	Name   string `json:"name"`
	Volume int    `json:"volume"`
}

type Achievement struct {
	ID    int    `json:"id"`
	Name  UUID   `json:"name"`
	Desc  UUID   `json:"desc"`
	Image string `json:"image"`
}

type Hint struct {
	ID       int        `json:"id"`
	Commands []*Command `json:"commands"`
}

type IAPProduct struct {
	ID  int    `json:"id"`
	Key string `json:"key"`
}

type Item struct {
	ID       int        `json:"id"`
	Name     UUID       `json:"name"`
	Icon     string     `json:"icon"`
	Desc     UUID       `json:"desc"`
	Commands []*Command `json:"commands"`
}

type Combine struct {
	ID       int         `json:"id"`
	Item1    int         `json:"item1"`
	Item2    int         `json:"item2"`
	Type     CombineType `json:"type"`
	Commands []*Command  `json:"commands"`
}

func (g *Game) CreateCombine(itemID1, itemID2 int) *Combine {
	for _, combine := range g.Combines {
		if (combine.Item1 == itemID1 && combine.Item2 == itemID2) || (combine.Type == CombineTypeCombine && combine.Item1 == itemID2 && combine.Item2 == itemID1) {
			return combine
		}
	}
	return nil
}
