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

package texts

import (
	"golang.org/x/text/language"
)

type TextID int

const (
	TextIDNewGame TextID = iota
	TextIDResumeGame
	TextIDYes
	TextIDNo
	TextIDLanguage
	TextIDCredit
	TextIDRemoveAds
	TextIDReviewThisApp
	TextIDRestorePurchases
	TextIDMoreGames
	TextIDClose
)

func Text(lang language.Tag, id TextID) string {
	if lang == language.Und {
		lang = language.English
	}
	return texts[lang][id]
}

var texts = map[language.Tag]map[TextID]string{
	language.English: {
		TextIDNewGame:          "New Game",
		TextIDResumeGame:       "Resume Game",
		TextIDYes:              "Yes",
		TextIDNo:               "No",
		TextIDLanguage:         "Language",
		TextIDCredit:           "Credit",
		TextIDRemoveAds:        "Remove Ads",
		TextIDReviewThisApp:    "Review This App",
		TextIDRestorePurchases: "Restore Purchases",
		TextIDMoreGames:        "More Games",
		TextIDClose:            "Close",
	},
	language.Japanese: {
		TextIDNewGame:          "初めから",
		TextIDResumeGame:       "続きから",
		TextIDYes:              "はい",
		TextIDNo:               "いいえ",
		TextIDLanguage:         "言語",
		TextIDCredit:           "クレジット",
		TextIDRemoveAds:        "広告を非表示にする",
		TextIDReviewThisApp:    "このアプリをレビューする",
		TextIDRestorePurchases: "購入情報再取得",
		TextIDMoreGames:        "ほかのゲーム",
		TextIDClose:            "閉じる",
	},
}
