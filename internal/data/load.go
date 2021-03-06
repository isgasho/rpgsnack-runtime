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

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/vmihailenco/msgpack"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
)

var assetDirs = []string{
	filepath.Join("audio", "bgm"),
	filepath.Join("audio", "se"),
	filepath.Join("audio", "se", "system"),
	filepath.Join("images", "backgrounds"),
	filepath.Join("images", "characters"),
	filepath.Join("images", "fonts"),
	filepath.Join("images", "foregrounds"),
	filepath.Join("images", "icons"),
	filepath.Join("images", "items", "preview"),
	filepath.Join("images", "pictures"),
	filepath.Join("images", "system", "common"),
	filepath.Join("images", "system", "game"),
	filepath.Join("images", "system", "footer"),
	filepath.Join("images", "system", "itempreview"),
	filepath.Join("images", "system", "splash"),
	filepath.Join("images", "system", "minigame"),
	filepath.Join("images", "tilesets", "backgrounds"),
	filepath.Join("images", "tilesets", "autotiles"),
	filepath.Join("images", "tilesets", "objects"),
	filepath.Join("images", "titles"),
}

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

func unmarshalJSON(data []uint8, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		switch err := err.(type) {
		case *json.UnmarshalTypeError:
			begin := max(int(err.Offset)-20, 0)
			end := min(int(err.Offset)+40, len(data))
			part := string(data[begin:end])
			return fmt.Errorf("data JSON type error: %s:\n%s", err.Error(), part)
		case *json.SyntaxError:
			begin := max(int(err.Offset)-20, 0)
			end := min(int(err.Offset)+40, len(data))
			part := string(data[begin:end])
			return fmt.Errorf("data: JSON syntax error: %s:\n%s", err.Error(), part)
		}
		return err
	}
	return nil
}

type rawData struct {
	Project     []byte
	ProjectJSON []byte
	Assets      [][]byte
	Progress    []byte
	Permanent   []byte
	Purchases   []byte
	Language    []byte
}

type Project struct {
	Data *Game `msgpack:"data"`
}

type LoadedData struct {
	Game           *Game
	Assets         map[string][]byte
	AssetsMetadata map[string]*AssetMetadata
	Progress       []byte
	Permanent      []byte
	Purchases      []string
	Language       language.Tag
}

type LoadProgress struct {
	Progress   float64
	LoadedData *LoadedData
	Error      error
}

type assetsReader struct {
	rawAssets [][]byte
}

func (a *assetsReader) Read(buf []byte) (int, error) {
	if len(a.rawAssets) == 0 {
		return 0, io.EOF
	}
	n := copy(buf, a.rawAssets[0])
	a.rawAssets[0] = a.rawAssets[0][n:]
	if len(a.rawAssets[0]) == 0 {
		a.rawAssets = a.rawAssets[1:]
	}
	return n, nil
}

func Load(projectionLocation string, progress chan<- LoadProgress) {
	defer close(progress)

	rawDataProgress := make(chan float64, 16)
	rawDataProgressDone := make(chan struct{})
	go func() {
		for v := range rawDataProgress {
			progress <- LoadProgress{
				Progress: v / 2,
			}
		}
		close(rawDataProgressDone)
	}()
	data, err := loadRawData(projectionLocation, rawDataProgress)
	if err != nil {
		progress <- LoadProgress{
			Error: fmt.Errorf("data: loadRawData failed: %s", err.Error()),
		}
		return
	}
	<-rawDataProgressDone

	gameDataCh := make(chan *Game)
	assetsCh := make(chan map[string][]byte)
	assetsMetadataCh := make(chan map[string]*AssetMetadata)
	purchasesCh := make(chan []string)
	errCh := make(chan error)

	go func() {
		var project *Project
		if data.Project != nil {
			if err := msgpack.Unmarshal(data.Project, &project); err != nil {
				errCh <- fmt.Errorf("data: parsing project data failed (Msgpack): %s", err.Error())
				return
			}
		}
		gameDataCh <- project.Data
	}()
	go func() {
		assets, assetsMetadata, err := parseAssets(&assetsReader{data.Assets})
		if err != nil {
			errCh <- err
			return
		}
		assetsCh <- assets
		assetsMetadataCh <- assetsMetadata
	}()
	go func() {
		var purchases []string
		if data.Purchases != nil {
			if err := unmarshalJSON(data.Purchases, &purchases); err != nil {
				errCh <- fmt.Errorf("data: parsing purchases data failed: %s", err.Error())
				return
			}
		} else {
			purchases = []string{}
		}
		purchasesCh <- purchases
	}()

	loadedData := &LoadedData{
		Progress:  data.Progress,
		Permanent: data.Permanent,
	}

	count := 0
	for count < 6 {
		select {
		case loadedData.Game = <-gameDataCh:
			count += 3
		case loadedData.Assets = <-assetsCh:
			count++
		case loadedData.AssetsMetadata = <-assetsMetadataCh:
			count++
		case loadedData.Purchases = <-purchasesCh:
			count++
		case err := <-errCh:
			progress <- LoadProgress{
				Error: err,
			}
			return
		}
		progress <- LoadProgress{
			Progress: 0.5 + float64(count)/((6+1)*2),
		}
	}

	var tag language.Tag
	if data.Language != nil {
		var langID string
		if err := unmarshalJSON(data.Language, &langID); err != nil {
			progress <- LoadProgress{
				Error: fmt.Errorf("data: parsing language data failed: %s", err.Error()),
			}
			return
		}
		tag, err = language.Parse(langID)
		if err != nil {
			progress <- LoadProgress{
				Error: err,
			}
			return
		}
	} else {
		tag = language.Tag(loadedData.Game.System.DefaultLanguage)
	}
	loadedData.Language = lang.Normalize(tag)
	progress <- LoadProgress{
		Progress:   1,
		LoadedData: loadedData,
	}
}

func parseAssets(rawAssets io.Reader) (map[string][]byte, map[string]*AssetMetadata, error) {
	var m map[string][]byte
	assets := map[string][]byte{}
	assetsMetadata := map[string]*AssetMetadata{}

	if err := msgpack.NewDecoder(rawAssets).Decode(&m); err != nil {
		return nil, nil, fmt.Errorf("data: msgpack.Unmarshal error: %s", err.Error())
	}

	for k, v := range m {
		if filepath.Ext(k) == ".json" {
			var assetMetadata *AssetMetadata
			if err := unmarshalJSON(v, &assetMetadata); err != nil {
				return nil, nil, fmt.Errorf("data: parsing asset metadata %s failed: %s", k, err.Error())
			}
			assetsMetadata[k] = assetMetadata
		} else {
			assets[k] = v
		}
	}

	return assets, assetsMetadata, nil
}
