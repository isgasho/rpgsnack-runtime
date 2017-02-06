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

package mobile

import (
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
)

type Requester scene.Requester

func FinishUnlockAhievement(id int, achievements string, err string) {
}

func FinishSaveProgress(id int, err string) {
}

func FinishPurchase(id int, productIDs string, err string) {
}

func FinishInterstitialAds(id int, err string) {
}

func FinishRewardedAds(id int, err string) {
}

func FinishOpenLink(id int, err string) {
}

func FinishShareImage(id int, err string) {
}