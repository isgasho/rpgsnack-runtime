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

// +build js

package data

import (
	"path"

	"github.com/gopherjs/gopherjs/js"
)

func fetch(path string) <-chan []uint8 {
	// TODO: Use fetch API in the future.
	ch := make(chan []uint8)
	xhr := js.Global.Get("XMLHttpRequest").New()
	xhr.Set("responseType", "arraybuffer")
	xhr.Call("addEventListener", "load", func() {
		res := xhr.Get("response")
		ch <- js.Global.Get("Uint8Array").New(res).Interface().([]uint8)
		close(ch)
	})
	xhr.Call("open", "GET", path)
	xhr.Call("send")
	return ch
}

func fetchProgress() <-chan []uint8 {
	ch := make(chan []uint8)
	go func() {
		data := js.Global.Get("localStorage").Call("getItem", "progress")
		if data == nil {
			close(ch)
			return
		}
		ch <- js.Global.Get("Uint8Array").New(data).Interface().([]uint8)
		close(ch)
	}()
	return ch
}

func loadRawData(projectPath string) (*rawData, error) {
	return &rawData{
		Project:   <-fetch(path.Join(projectPath, "project.json")),
		Assets:    <-fetch(path.Join(projectPath, "assets.msgpack")),
		Progress:  <-fetchProgress(),
		Purchases: nil, // TODO: Implement this
		Language:  []uint8(`"en"`),
	}, nil
}
