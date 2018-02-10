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
	"path/filepath"

	"github.com/vmihailenco/msgpack"
	"golang.org/x/text/language"
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
	Project   []byte
	Assets    []byte
	Progress  []byte
	Purchases []byte
	Language  []byte
}

type Project struct {
	Data *Game `json:data`
}

type LoadedData struct {
	Game           *Game
	Assets         map[string][]byte
	AssetsMetadata map[string]*AssetMetadata
	Progress       []byte
	Purchases      []string
	Language       language.Tag
}

func Load(projectPath string) (*LoadedData, error) {
	data, err := loadRawData(projectPath)
	if err != nil {
		return nil, fmt.Errorf("data: loadRawData failed: %s", err.Error())
	}
	var project *Project
	if err := unmarshalJSON(data.Project, &project); err != nil {
		return nil, fmt.Errorf("data: parsing project data failed: %s", err.Error())
	}
	gameData := project.Data
	assets, assetsMetadata, err := parseAssets(data.Assets)
	if err := msgpack.Unmarshal(data.Assets, &assets); err != nil {
		return nil, fmt.Errorf("data: msgpack.Unmarshal error: %s", err.Error())
	}
	var purchases []string
	if data.Purchases != nil {
		if err := unmarshalJSON(data.Purchases, &purchases); err != nil {
			return nil, fmt.Errorf("data: parsing purchases data failed: %s", err.Error())
		}
	} else {
		purchases = []string{}
	}
	var tag language.Tag
	if data.Language != nil {
		var langId string
		if err := unmarshalJSON(data.Language, &langId); err != nil {
			return nil, fmt.Errorf("data: parsing language data failed: %s", err.Error())
		}
		tag, err = language.Parse(langId)
		if err != nil {
			return nil, err
		}
	} else {
		tag = gameData.System.DefaultLanguage
	}

	base, _ := tag.Base()
	newTag, _ := language.Compose(base)
	if newTag == language.Chinese {
		// If the language is Chinese use zh-Hans or zh-Hant.
		s, _ := tag.Script()
		if s.String() != "Hans" && s.String() != "Hant" {
			// If the language is just "zh" or other Chinese, use Hans (simplified).
			// There is no strong reason why Hans is preferred.
			s = language.MustParseScript("Hans")
		}
		newTag, err = language.Compose(base, s)
		if err != nil {
			return nil, err
		}
	}
	tag = newTag

	// Don't set the language here.
	// Determining a language requires checking the game text data.

	return &LoadedData{
		Game:           gameData,
		Assets:         assets,
		AssetsMetadata: assetsMetadata,
		Purchases:      purchases,
		Progress:       data.Progress,
		Language:       tag,
	}, nil
}

func parseAssets(rawAssets []byte) (map[string][]byte, map[string]*AssetMetadata, error) {
	var m map[string][]byte
	assets := map[string][]byte{}
	assetsMetadata := map[string]*AssetMetadata{}

	if err := msgpack.Unmarshal(rawAssets, &m); err != nil {
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
