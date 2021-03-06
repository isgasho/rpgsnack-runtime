// Copyright 2019 Hajime Hoshi
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

package scene

type MinigameData struct {
	Score        int   `msgpack:"score"`
	LastActiveAt int64 `msgpack:"lastActiveAt"`
}

type Permanent struct {
	Minigames         []*MinigameData `msgpack:"minigame"`
	Variables         []int64         `msgpack:"variables"`
	BGMMute           int             `msgpack:"bgm_mute"`
	SEMute            int             `msgpack:"se_mute"`
	VibrationDisabled bool            `msgpack:"vibrationDisabled"`
}
