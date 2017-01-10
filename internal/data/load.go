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

// +build !js

package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func Load() error {
	jsonPath := "data.json"
	if len(os.Args) >= 2 {
		jsonPath = os.Args[1]
	}
	dataJson, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var gameData *Game
	if err := json.Unmarshal(dataJson, &gameData); err != nil {
		if err, ok := err.(*json.SyntaxError); ok {
			begin := max(int(err.Offset)-20, 0)
			end := min(int(err.Offset)+40, len(dataJson))
			part := string(dataJson[begin:end])
			return fmt.Errorf("data: JSON syntax error: %s:\n%s", err.Error(), part)
		}
		return err
	}
	current = gameData
	return nil
}
