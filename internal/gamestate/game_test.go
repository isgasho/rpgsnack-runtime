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

package gamestate_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/hajimehoshi/rpgsnack-runtime/internal/gamestate"
)

func TestRandomValue(t *testing.T) {
	value := []int{1, 3}
	g := &Game{}
	// TODO: We should mock math.random for consistent results
	g.SetRandomSource(rand.NewSource(time.Now().UnixNano()))
	got := g.RandomValue(value)
	if got <= 0 || got >= 4 {
		t.Errorf("RandomValue([1, 3]) out of range got: %v", got)
	}
}
