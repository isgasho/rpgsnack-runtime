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

package gamestate

import (
	"fmt"
	"image/color"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/vmihailenco/msgpack"

	"github.com/hajimehoshi/rpgsnack-runtime/internal/audio"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/character"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/consts"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/data"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/easymsgpack"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/hints"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/input"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/items"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/lang"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/picture"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/scene"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/variables"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/weather"
	"github.com/hajimehoshi/rpgsnack-runtime/internal/window"
)

type Minigame struct {
	score        int
	lastActiveAt int64
	active       bool
	id           int
	reqScore     int
}

func (m *Minigame) Active() bool {
	if m == nil {
		return false
	}
	return m.active
}

func (m *Minigame) activate(lastActiveAt int64) {
	m.active = true
	m.lastActiveAt = lastActiveAt
}

func (m *Minigame) deactivate() {
	if m == nil {
		return
	}
	m.active = false
}

func (m *Minigame) ID() int {
	return m.id
}

func (m *Minigame) ReqScore() int {
	return m.reqScore
}

func (m *Minigame) Score() int {
	return m.score
}

func (m *Minigame) LastActiveAt() int64 {
	return m.lastActiveAt
}

func (m *Minigame) MarkLastActive() {
	m.lastActiveAt = time.Now().Unix()
}

func (m *Minigame) AddScore(score int) {
	m.score += score
}

func (m *Minigame) Success() bool {
	if m == nil {
		return false
	}
	return m.score >= m.reqScore
}

type Rand interface {
	Intn(n int) int
}

type Game struct {
	hints                *hints.Hints
	items                *items.Items
	variables            *variables.Variables
	screen               *Screen
	windows              *window.Windows
	pictures             *picture.Pictures
	currentMap           *Map
	lastInterpreterID    int
	autoSaveEnabled      bool
	playerControlEnabled bool
	inventoryVisible     bool
	weatherType          data.WeatherType
	cleared              bool

	lastPlayingBGMName   string
	lastPlayingBGMVolume float64

	backgrounds map[int]map[int]string
	foregrounds map[int]map[int]string
	playerSpeed data.Speed

	// Fields that are not dumped
	pressedPictureID             int
	releasedPictureID            int
	triggeredPictureID           int
	isTitle                      bool
	rand                         Rand
	waitingRequestIDs            map[int]struct{}
	prices                       map[string]string // TODO: We want to use https://godoc.org/golang.org/x/text/currency
	weather                      *weather.Weather
	onShakeStartGameButton       func()
	shouldShowCredits            bool
	shouldShowCreditsCloseButton bool
	minigame                     *Minigame
}

func generateDefaultRand() Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func NewGame() *Game {
	g := &Game{
		currentMap:           NewMap(),
		hints:                &hints.Hints{},
		items:                &items.Items{},
		variables:            &variables.Variables{},
		screen:               &Screen{},
		windows:              &window.Windows{},
		pictures:             &picture.Pictures{},
		rand:                 generateDefaultRand(),
		autoSaveEnabled:      true,
		playerControlEnabled: true,
		playerSpeed:          data.Speed5,
	}
	return g
}

func NewTitleGame(savedGame *Game, onShakeStartGameButton func()) *Game {
	g := &Game{
		currentMap:             NewTitleMap(),
		hints:                  &hints.Hints{},
		items:                  &items.Items{},
		variables:              &variables.Variables{},
		screen:                 &Screen{},
		windows:                &window.Windows{},
		pictures:               &picture.Pictures{},
		rand:                   generateDefaultRand(),
		playerControlEnabled:   true,
		playerSpeed:            data.Speed5,
		isTitle:                true,
		onShakeStartGameButton: onShakeStartGameButton,
	}

	if savedGame != nil {
		g.items = savedGame.items
		g.variables = savedGame.variables
	}

	return g
}

func (g *Game) EncodeMsgpack(enc *msgpack.Encoder) error {
	e := easymsgpack.NewEncoder(enc)
	e.BeginMap()

	e.EncodeString("hints")
	e.EncodeInterface(g.hints)

	e.EncodeString("items")
	e.EncodeInterface(g.items)

	e.EncodeString("variables")
	e.EncodeInterface(g.variables)

	e.EncodeString("screen")
	e.EncodeInterface(g.screen)

	e.EncodeString("windows")
	e.EncodeInterface(g.windows)

	e.EncodeString("pictures")
	e.EncodeInterface(g.pictures)

	e.EncodeString("currentMap")
	e.EncodeInterface(g.currentMap)

	e.EncodeString("lastInterpreterId")
	e.EncodeInt(g.lastInterpreterID)

	e.EncodeString("autoSaveEnabled")
	e.EncodeBool(g.autoSaveEnabled)

	e.EncodeString("playerControlEnabled")
	e.EncodeBool(g.playerControlEnabled)

	e.EncodeString("inventoryVisible")
	e.EncodeBool(g.inventoryVisible)

	e.EncodeString("weatherType")
	e.EncodeString(string(g.weatherType))

	e.EncodeString("cleared")
	e.EncodeBool(g.cleared)

	e.EncodeString("lastPlayingBGMName")
	e.EncodeString(audio.PlayingBGMName())

	e.EncodeString("lastPlayingBGMVolume")
	e.EncodeFloat64(audio.PlayingBGMVolume())

	e.EncodeString("playerSpeed")
	e.EncodeInt(int(g.playerSpeed))

	e.EncodeString("backgrounds")
	e.BeginMap()
	for id, m := range g.backgrounds {
		e.EncodeInt(id)
		e.BeginMap()
		for id, r := range m {
			e.EncodeInt(id)
			e.EncodeString(r)
		}
		e.EndMap()
	}
	e.EndMap()

	e.EncodeString("foregrounds")
	e.BeginMap()
	for id, m := range g.foregrounds {
		e.EncodeInt(id)
		e.BeginMap()
		for id, r := range m {
			e.EncodeInt(id)
			e.EncodeString(r)
		}
		e.EndMap()
	}
	e.EndMap()

	e.EndMap()
	return e.Flush()
}

func (g *Game) DecodeMsgpack(dec *msgpack.Decoder) error {
	d := easymsgpack.NewDecoder(dec)
	n := d.DecodeMapLen()
	for i := 0; i < n; i++ {
		k := d.DecodeString()
		switch k {
		case "hints":
			if !d.SkipCodeIfNil() {
				g.hints = &hints.Hints{}
				d.DecodeInterface(g.hints)
			}
		case "items":
			if !d.SkipCodeIfNil() {
				g.items = &items.Items{}
				d.DecodeInterface(g.items)
			}
		case "variables":
			if !d.SkipCodeIfNil() {
				g.variables = &variables.Variables{}
				d.DecodeInterface(g.variables)
			}
		case "screen":
			if !d.SkipCodeIfNil() {
				g.screen = &Screen{}
				d.DecodeInterface(g.screen)
			}
		case "windows":
			if !d.SkipCodeIfNil() {
				g.windows = &window.Windows{}
				d.DecodeInterface(g.windows)
			}
		case "pictures":
			if !d.SkipCodeIfNil() {
				g.pictures = &picture.Pictures{}
				d.DecodeInterface(g.pictures)
			}
		case "currentMap":
			if !d.SkipCodeIfNil() {
				g.currentMap = &Map{}
				d.DecodeInterface(g.currentMap)
			}
		case "lastInterpreterId":
			g.lastInterpreterID = d.DecodeInt()
		case "autoSaveEnabled":
			g.autoSaveEnabled = d.DecodeBool()
		case "playerControlEnabled":
			g.playerControlEnabled = d.DecodeBool()
		case "inventoryVisible":
			g.inventoryVisible = d.DecodeBool()
		case "weatherType":
			g.SetWeather(data.WeatherType(d.DecodeString()))
		case "cleared":
			g.cleared = d.DecodeBool()
		case "lastPlayingBGMName":
			g.lastPlayingBGMName = d.DecodeString()
		case "lastPlayingBGMVolume":
			g.lastPlayingBGMVolume = d.DecodeFloat64()
		case "playerSpeed":
			g.playerSpeed = data.Speed(d.DecodeInt())
			if g.playerSpeed == 0 {
				// The save data might be created before playerSpeed was introduced. Let's fallback.
				// (#843)
				g.playerSpeed = data.Speed5
			}
		case "backgrounds":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				g.backgrounds = map[int]map[int]string{}
				for i := 0; i < n; i++ {
					id := d.DecodeInt()
					g.backgrounds[id] = map[int]string{}
					n2 := d.DecodeMapLen()
					for j := 0; j < n2; j++ {
						id2 := d.DecodeInt()
						r := d.DecodeString()
						g.backgrounds[id][id2] = r
					}
				}
			}
		case "foregrounds":
			if !d.SkipCodeIfNil() {
				n := d.DecodeMapLen()
				g.foregrounds = map[int]map[int]string{}
				for i := 0; i < n; i++ {
					id := d.DecodeInt()
					g.foregrounds[id] = map[int]string{}
					n2 := d.DecodeMapLen()
					for j := 0; j < n2; j++ {
						id2 := d.DecodeInt()
						r := d.DecodeString()
						g.foregrounds[id][id2] = r
					}
				}
			}
		default:
			if err := d.Error(); err != nil {
				return err
			}
			return fmt.Errorf("gamestate: Game.DecodeMsgpack failed: unknown key: %s", k)
		}
	}
	g.rand = generateDefaultRand()
	if err := d.Error(); err != nil {
		return fmt.Errorf("gamestate: Game.DecodeMsgpack failed: %v", err)
	}
	return nil
}

func (g *Game) Items() *items.Items {
	return g.items
}

// TODO: Remove this
func (g *Game) Map() *Map {
	return g.currentMap
}

func (g *Game) MapPassableAt(through bool, x, y int, ignoreCharacters bool) bool {
	return g.currentMap.Passable(through, x, y, ignoreCharacters)
}

type messageSyntaxParser struct {
	game         *Game
	sceneManager *scene.Manager
}

func (m *messageSyntaxParser) ParseMessageSyntax(content string) string {
	return m.game.parseMessageSyntax(m.sceneManager, content)
}

func (g *Game) Update(sceneManager *scene.Manager) error {
	g.items.SetDataItems(sceneManager.Game().Items)
	if g.lastPlayingBGMName != "" {
		audio.PlayBGM(g.lastPlayingBGMName, g.lastPlayingBGMVolume, 0)
		g.lastPlayingBGMName = ""
		g.lastPlayingBGMVolume = 0
	}
	for id := range g.waitingRequestIDs {
		if sceneManager.ReceiveResultIfExists(id) != nil {
			delete(g.waitingRequestIDs, id)
		}
	}
	g.weather.Update()
	g.screen.Update()
	playerY := 0
	if g.currentMap.player != nil {
		_, playerY = g.currentMap.player.DrawPosition()
	}
	g.windows.Update(playerY, &messageSyntaxParser{g, sceneManager}, sceneManager, g.createCharacterList())
	g.pictures.Update()

	if err := g.currentMap.Update(sceneManager, g); err != nil {
		return err
	}
	return nil
}

func (g *Game) Clear() {
	g.cleared = true
}

func (g *Game) SetBGM(bgm data.BGM) {
	if bgm.Name == "" {
		audio.StopBGM(0)
	} else {
		audio.PlayBGM(bgm.Name, float64(bgm.Volume)/100, 0)
	}
}

func (g *Game) ShowInventory(group int, wait bool, cancelable bool) {
	g.inventoryVisible = true
	g.items.SetActiveItemGroup(group)
	g.items.SetChoiceMode(wait, cancelable)
}

func (g *Game) HideInventory() {
	g.inventoryVisible = false
}

func (g *Game) InventoryVisible() bool {
	return g.inventoryVisible
}

func (g *Game) InitMinigame(id, reqScore, score int) {
	g.minigame = &Minigame{
		id:       id,
		reqScore: reqScore,
		score:    score,
	}
}

func (g *Game) ShowMinigame(lastActiveAt int64) {
	g.minigame.activate(lastActiveAt)
}

func (g *Game) HideMinigame() {
	g.minigame.deactivate()
}

func (g *Game) Minigame() *Minigame {
	return g.minigame
}

func (g *Game) SetAutoSaveEnabled(enabled bool) {
	g.autoSaveEnabled = enabled
}

func (g *Game) IsAutoSaveEnabled() bool {
	return g.autoSaveEnabled
}

func (g *Game) SetPlayerControlEnabled(enabled bool) {
	g.playerControlEnabled = enabled
}

func (g *Game) IsPlayerControlEnabled() bool {
	return g.playerControlEnabled
}

// RequestSave requests to save the progress to the platform.
func (g *Game) RequestSave(requestID int, sceneManager *scene.Manager) {
	if g.isTitle {
		return
	}

	id := requestID
	if id == 0 {
		id = sceneManager.GenerateRequestID()
		if g.waitingRequestIDs == nil {
			g.waitingRequestIDs = map[int]struct{}{}
		}
		g.waitingRequestIDs[id] = struct{}{}
	}

	m, err := msgpack.Marshal(g)
	if err != nil {
		panic(fmt.Sprintf("gamestate: msgpack encoding error: %v", err))
	}
	sceneManager.Requester().RequestSaveProgress(id, m)
	sceneManager.SetProgress(m)
}

func (g *Game) RequestSavePermanentVariable(requestID int, sceneManager *scene.Manager, permanentVariableID, variableID int) bool {
	v := int64(g.VariableValue(variableID))
	sceneManager.RequestSavePermanentVariable(requestID, permanentVariableID, v)
	return true
}

func (g *Game) RequestSavePermanentMinigame(requestID int, sceneManager *scene.Manager, id, score int, lastActiveAt int64) bool {
	sceneManager.RequestSavePermanentMinigame(requestID, id, score, lastActiveAt)
	return true
}

func (g *Game) RequestRewardedAds(requestID int, sceneManager *scene.Manager, forceAds bool) bool {
	sceneManager.RequestRewardedAds(requestID, forceAds)
	return true
}

var (
	reMessageCommand  = regexp.MustCompile(`\\([a-zA-Z])\[([^\\]+)\]`)
	reMessageItem     = regexp.MustCompile(`([^:]+):name`)
	reMessageTable    = regexp.MustCompile(`([^:]+):([^:]+):([^:]+)`)
	reMessageVariable = regexp.MustCompile(`v\[([0-9]+)\]`)
)

func (g *Game) parseMessageSyntax(sceneManager *scene.Manager, str string) string {
	return reMessageCommand.ReplaceAllStringFunc(str, func(part string) string {
		name := strings.ToLower(part[1:2])
		args := part[3 : len(part)-1]

		switch name {
		case "p":
			return g.price(args)
		case "i":
			if m1 := reMessageItem.FindStringSubmatch(args); m1 != nil {
				m2 := reMessageVariable.FindStringSubmatch(m1[1])
				itemID := 0
				if m2 != nil {
					varID, err := strconv.Atoi(m2[1])
					if err != nil {
						return fmt.Sprintf("gamestate: strconv.Atoi failed1: %v", m2[1])
					}
					itemID = int(g.VariableValue(varID))
				} else {
					var err error
					itemID, err = strconv.Atoi(m1[1])
					if err != nil {
						return fmt.Sprintf("gamestate: strconv.Atoi failed2: %v", m1[1])
					}
				}

				return g.GetItemValueString(sceneManager, itemID)
			}
		case "v":
			id, err := strconv.Atoi(args)
			if err != nil {
				return fmt.Sprintf("(error:%v)", part)
			}
			return fmt.Sprintf("%d", g.variables.VariableValue(id))
		case "t":
			if m1 := reMessageTable.FindStringSubmatch(args); m1 != nil {
				tableName := m1[1]
				m2 := reMessageVariable.FindStringSubmatch(m1[2])
				recordID := 0
				if m2 != nil {
					varID, err := strconv.Atoi(m2[1])
					if err != nil {
						return fmt.Sprintf("(error:subGroup:%v)", m2[1])
					}
					recordID = int(g.VariableValue(varID))
				} else {
					var err error
					recordID, err = strconv.Atoi(m1[2])
					if err != nil {
						return fmt.Sprintf("(error:group:%v)", m1[2])
					}
				}
				attrName := m1[3]

				return g.GetTableValueString(sceneManager, tableName, recordID, attrName)
			}
		}
		return str
	})
}

const (
	specialConditionEventExistsAtPlayer = "event_exists_at_player"
)

func (g *Game) MeetsCondition(cond *data.Condition, eventID int) (bool, error) {
	if cond == nil {
		return true, nil
	}
	switch cond.Type {
	case data.ConditionTypeSwitch:
		id := cond.ID
		v := g.variables.SwitchValue(id)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeSelfSwitch:
		m, r := g.currentMap.mapID, g.currentMap.roomID
		v := g.variables.SelfSwitchValue(m, r, eventID, cond.ID)
		rhs := cond.Value.(bool)
		return v == rhs, nil
	case data.ConditionTypeVariable:
		id := cond.ID
		v := g.variables.VariableValue(id)
		var rhs int64
		// TODO: This is redundant: can we refactor them?
		switch value := cond.Value.(type) {
		case float32:
			rhs = int64(value)
		case float64:
			rhs = int64(value)
		case int:
			rhs = int64(value)
		case int8:
			rhs = int64(value)
		case int16:
			rhs = int64(value)
		case int32:
			rhs = int64(value)
		case int64:
			rhs = value
		case uint8:
			rhs = int64(value)
		case uint16:
			rhs = int64(value)
		case uint32:
			rhs = int64(value)
		case uint64:
			rhs = int64(value)
		}
		switch cond.ValueType {
		case data.ConditionValueTypeConstant:
		case data.ConditionValueTypeVariable:
			rhs = g.variables.VariableValue(int(rhs))
		default:
			return false, fmt.Errorf("gamestate: invalid value type: %v eventID %d", cond, eventID)
		}
		switch cond.Comp {
		case data.ConditionCompEqualTo:
			return v == rhs, nil
		case data.ConditionCompNotEqualTo:
			return v != rhs, nil
		case data.ConditionCompGreaterThanOrEqualTo:
			return v >= rhs, nil
		case data.ConditionCompGreaterThan:
			return v > rhs, nil
		case data.ConditionCompLessThanOrEqualTo:
			return v <= rhs, nil
		case data.ConditionCompLessThan:
			return v < rhs, nil
		default:
			return false, fmt.Errorf("gamestate: invalid comp: %s eventID %d", cond.Comp, eventID)
		}
	case data.ConditionTypeItem:
		id := cond.ID
		itemValue := data.ConditionItemValue(cond.Value.(string))

		switch itemValue {
		case data.ConditionItemOwn:
			if id == 0 {
				return g.items.ItemNum() > 0, nil
			} else {
				return g.items.Includes(id), nil
			}
		case data.ConditionItemNotOwn:
			if id == 0 {
				return g.items.ItemNum() == 0, nil
			} else {
				return !g.items.Includes(id), nil
			}
		case data.ConditionItemActive:
			if id == 0 {
				return g.items.ActiveItem() > 0, nil
			} else {
				return id == g.items.ActiveItem(), nil
			}

		default:
			return false, fmt.Errorf("gamestate: invalid item value: %s eventID %d", itemValue, eventID)
		}
	case data.ConditionTypeSpecial:
		switch cond.Value.(string) {
		case specialConditionEventExistsAtPlayer:
			e := g.currentMap.executableEventAt(g.currentMap.player.Position())
			return e != nil, nil
		default:
			return false, fmt.Errorf("gamestate: ConditionTypeSpecial: invalid value: %v eventID %d", cond, eventID)
		}
	default:
		return false, fmt.Errorf("gamestate: invalid condition: %v eventID %d", cond, eventID)
	}
	return false, nil
}

func (g *Game) GenerateInterpreterID() int {
	g.lastInterpreterID++
	return g.lastInterpreterID
}

func (g *Game) SetRandomForTesting(r Rand) {
	g.rand = r
}

func (g *Game) RandomValue(min, max int) int {
	return min + g.rand.Intn(max-min)
}

func (g *Game) DrawWeather(screen *ebiten.Image) {
	g.weather.Draw(screen)
}

func (g *Game) ApplyTintColor(c *ebiten.ColorM) {
	g.screen.ApplyTintColor(c)
}

func (g *Game) ZeroTint() bool {
	return g.screen.ZeroTint()
}

func (g *Game) ApplyShake(geo *ebiten.GeoM) {
	g.screen.ApplyShake(geo)
}

func (g *Game) DrawScreen(screenImage *ebiten.Image) {
	g.screen.Draw(screenImage)
}

func (g *Game) DrawWindows(screen *ebiten.Image, offsetX, offsetY, windowOffsetY int) {
	g.windows.Draw(screen, g.createCharacterList(), offsetX, offsetY, windowOffsetY)
}

func (g *Game) createCharacterList() []*character.Character {
	cs := []*character.Character{}
	cs = append(cs, g.currentMap.player)
	cs = append(cs, g.currentMap.events...)
	return cs
}

func (g *Game) DrawPictures(screen *ebiten.Image, offsetX, offsetY int, priority data.PicturePriorityType) {
	g.pictures.Draw(screen, offsetX, offsetY, priority)
}

func (g *Game) Character(mapID, roomID, eventID int) *character.Character {
	if eventID == character.PlayerEventID {
		return g.currentMap.player
	}
	if g.currentMap.mapID != mapID {
		return nil
	}
	if g.currentMap.roomID != roomID {
		return nil
	}
	for _, e := range g.currentMap.events {
		if eventID == e.EventID() {
			return e
		}
	}
	return nil
}

func (g *Game) price(productID string) string {
	if _, ok := g.prices[productID]; ok {
		return g.prices[productID]
	}
	return ""
}

func (g *Game) SetPrices(p map[string]string) {
	g.prices = p
}

func (g *Game) CanWindowProceed(interpreterID int) bool {
	return g.windows.CanProceed(interpreterID)
}

func (g *Game) IsWindowBusy() bool {
	return g.windows.IsBusy(0)
}

func (g *Game) IsWindowAnimating(interpreterID int) bool {
	return g.windows.IsAnimating(interpreterID)
}

func (g *Game) CloseAllWindows() {
	g.windows.CloseAll()
}

func (g *Game) HasChosenWindowIndex() bool {
	return g.windows.HasChosenIndex()
}

func (g *Game) ChosenWindowIndex() int {
	return g.windows.ChosenIndex()
}

func (g *Game) ShowBalloon(sceneManager *scene.Manager, interpreterID, mapID, roomID, eventID int, contentID data.UUID, balloonType data.BalloonType, messageStyle *data.MessageStyle) bool {
	ch := g.Character(mapID, roomID, eventID)
	if ch == nil {
		return false
	}

	g.windows.ShowBalloon(contentID, &messageSyntaxParser{g, sceneManager}, sceneManager.Game(), balloonType, eventID, interpreterID, messageStyle)
	return true
}

func (g *Game) ShowMessage(sceneManager *scene.Manager, interpreterID, eventID int, contentID data.UUID, background data.MessageBackground, positionType data.MessagePositionType, textAlign data.TextAlign, messageStyle *data.MessageStyle) {
	g.windows.ShowMessage(contentID, &messageSyntaxParser{g, sceneManager}, sceneManager.Game(), eventID, background, positionType, textAlign, interpreterID, messageStyle)
}

func (g *Game) ShowChoices(sceneManager *scene.Manager, interpreterID int, eventID int, choiceIDs []data.UUID, conditions []*data.ChoiceCondition) {
	choices := []*window.Choice{}
	for i, id := range choiceIDs {
		choice := &window.Choice{ID: id, Checked: false}

		var err error
		m := true
		if i < len(conditions) {
			m, err = g.MeetsCondition(conditions[i].Visible, eventID)
			if err != nil {
				panic(err)
			}
		}

		if m {
			if i < len(conditions) && conditions[i].Checked != nil {
				m, err := g.MeetsCondition(conditions[i].Checked, eventID)
				if err != nil {
					panic(err)
				}
				choice.Checked = m
			}
			choices = append(choices, choice)
		}
	}
	g.windows.ShowChoices(&messageSyntaxParser{g, sceneManager}, sceneManager.Game(), choices, interpreterID)
}

func (g *Game) RealChoiceIndex(sceneManager *scene.Manager, index int, eventID int, conditions []*data.ChoiceCondition) int {
	if len(conditions) == 0 {
		return index
	}
	j := 0
	for i, condition := range conditions {
		m, err := g.MeetsCondition(condition.Visible, eventID)
		if err != nil {
			panic(err)
		}
		if !m {
			continue
		}
		if j == index {
			return i
		}
		j++
	}
	return -1
}

func (g *Game) SetSwitchValue(id int, value bool) {
	g.variables.SetSwitchValue(id, value)
}

func (g *Game) SetSwitchRefValue(id int, value bool) {
	g.variables.SetSwitchValue(int(g.VariableValue(id)), value)
}

func (g *Game) SetSelfSwitchValue(eventID int, id int, value bool) {
	m, r := g.currentMap.mapID, g.currentMap.roomID
	g.variables.SetSelfSwitchValue(m, r, eventID, id, value)
}

func (g *Game) SetVariableValue(id int, value int64) {
	g.variables.SetVariableValue(id, value)
}

func (g *Game) VariableValue(id int) int64 {
	return g.variables.VariableValue(id)
}

func (g *Game) SwitchValue(id int) int64 {
	if g.variables.SwitchValue(id) {
		return 1
	} else {
		return 0
	}
}

func (g *Game) StartCombineCommands(combine *data.Combine) {
	g.currentMap.StartCombineCommands(g, combine)
}

func (g *Game) StartItemCommands(itemID int) {
	g.currentMap.StartItemCommands(g, itemID)
}

func (g *Game) SetPlayerDir(dir data.Dir) {
	g.currentMap.player.SetDir(dir)
}

func (g *Game) SetWeather(weatherType data.WeatherType) {
	if g.weatherType == weatherType {
		return
	}
	g.weatherType = weatherType
	if weatherType == data.WeatherTypeNone {
		g.weather = nil
		return
	}
	g.weather = weather.New(weatherType)
}

func (g *Game) TransferPlayerImmediately(roomID, x, y int, interpreter *Interpreter) {
	g.currentMap.transferPlayerImmediately(g, roomID, x, y, interpreter)
}

func (g *Game) ExecutableEventAtPlayer() *character.Character {
	p := g.currentMap.player
	return g.currentMap.executableEventAt(p.Position())
}

func (g *Game) CurrentEvents() []*data.Event {
	return g.currentMap.CurrentRoom().Events
}

func (g *Game) SetFadeColor(clr color.Color) {
	g.screen.setFadeColor(clr)
}

func (g *Game) IsScreenFadedOut() bool {
	return g.screen.isFadedOut()
}

func (g *Game) IsScreenFading() bool {
	return g.screen.isFading()
}

func (g *Game) FadeIn(time int) {
	g.screen.fadeIn(time)
}

func (g *Game) FadeOut(time int) {
	g.screen.fadeOut(time)
}

func (g *Game) StartShaking(power, speed, count int, dir data.ShakeDirection) {
	g.screen.startShaking(power, speed, count, dir)
}

func (g *Game) StopShaking() {
	g.screen.stopShaking()
}

func (g *Game) IsShaking() bool {
	return g.screen.isShaking()
}

func (g *Game) StartTint(red, green, blue, gray float64, time int) {
	g.screen.startTint(red, green, blue, gray, time)
}

func (g *Game) IsChangingTint() bool {
	return g.screen.isChangingTint()
}

func (g *Game) RefreshEvents() error {
	return g.currentMap.refreshEvents(g)
}

func (g *Game) InterfaceToTableValue(sceneManager *scene.Manager, v interface{}) interface{} {
	a := v.(*data.TableValueArgs)
	id := a.ID
	if a.Type == data.ValueTypeVariable {
		id = int(g.VariableValue(id))
	}
	return sceneManager.Game().GetTableValue(a.Name, id, a.Attr)
}

func (g *Game) GetItemValueString(sceneManager *scene.Manager, itemID int) string {
	i := g.items.Item(itemID)
	if i == nil {
		panic(fmt.Sprintf("gamestate: GetItemValueString: invalid itemID %d", itemID))
	}
	return sceneManager.Game().Texts.Get(lang.Get(), i.Name)
}

func (g *Game) GetTableValueString(sceneManager *scene.Manager, tableName string, recordID int, attrName string) string {
	t := sceneManager.Game().GetTableValueType(tableName, attrName)
	v := sceneManager.Game().GetTableValue(tableName, recordID, attrName)
	r := ""
	switch t {
	case data.TableValueTypeUUID:
		key, err := data.UUIDFromString(v.(string))
		if err != nil {
			panic(fmt.Sprintf("GetTableValueString: invalid UUID %v", v))
		}
		r = sceneManager.Game().Texts.Get(lang.Get(), key)

	case data.TableValueTypeInt:
		i, ok := data.InterfaceToInt(v)
		if !ok {
			panic(fmt.Sprintf("GetTableValueString: v isn't an integer %s:%d:%s", tableName, recordID, attrName))
		}
		r = fmt.Sprintf("%d", i)

	case data.TableValueTypeString:
		r = v.(string)

	default:
		panic(fmt.Sprintf("GetTableValueString: invalid valueType %s", t))
	}

	return r
}

func (g *Game) calcVariableRhs(sceneManager *scene.Manager, lhs int64, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}, mapID, roomID, eventID int) (int64, error) {
	var rhs int64
	switch valueType {
	case data.SetVariableValueTypeConstant:
		switch value.(type) {
		case int:
			rhs = int64(value.(int))
		case int64:
			rhs = value.(int64)
		}
	case data.SetVariableValueTypeVariable:
		rhs = g.VariableValue(value.(int))
	case data.SetVariableValueTypeVariableRef:
		rhs = g.VariableValue(int(g.VariableValue(value.(int))))
	case data.SetVariableValueTypeSwitch:
		rhs = g.SwitchValue(value.(int))
	case data.SetVariableValueTypeSwitchRef:
		rhs = g.SwitchValue(int(g.VariableValue(value.(int))))
	case data.SetVariableValueTypeRandom:
		v := value.(*data.SetVariableValueRandom)
		rhs = int64(g.RandomValue(v.Begin, v.End+1))
	case data.SetVariableValueTypeCharacter:
		args := value.(*data.SetVariableCharacterArgs)
		id := args.EventID
		if id == 0 {
			id = eventID
		}
		ch := g.Character(mapID, roomID, id)
		if ch == nil {
			// TODO: return error?
			return 0, nil
		}
		dir := ch.Dir()
		switch args.Type {
		case data.SetVariableCharacterTypeDirection:
			switch dir {
			case data.DirUp:
				rhs = 0
			case data.DirRight:
				rhs = 1
			case data.DirDown:
				rhs = 2
			case data.DirLeft:
				rhs = 3
			default:
				panic(fmt.Sprintf("gamestate: invalid dir: %d at data.SetVariableValueTypeCharacter", dir))
			}
		case data.SetVariableCharacterTypeRoomX:
			x, _ := ch.Position()
			rhs = int64(x)
		case data.SetVariableCharacterTypeRoomY:
			_, y := ch.Position()
			rhs = int64(y)
		case data.SetVariableCharacterTypeScreenX:
			x, _ := ch.DrawPosition()
			rhs = int64(x)
		case data.SetVariableCharacterTypeScreenY:
			_, y := ch.DrawPosition()
			rhs = int64(y)
		case data.SetVariableCharacterTypeIsPressed:
			x, y := ch.Position()
			pressX, pressY := g.currentMap.GetPressedPosition()
			if x == pressX && y == pressY {
				rhs = 1
			}
		default:
			return 0, fmt.Errorf("gamestate: not implemented yet (set_variable)(character): type %s", args.Type)
		}

	case data.SetVariableValueTypeItemGroup:
		args := value.(*data.SetVariableItemGroupArgs)
		group := args.Group
		switch args.Type {
		case data.SetVariableItemGroupTypeOwned:
			rhs = int64(g.items.ItemCount(group, true))
		case data.SetVariableItemGroupTypeTotal:
			rhs = int64(g.items.ItemCount(group, false))
		default:
			return 0, fmt.Errorf("gamestate: not implemented yet (set_variable)(item_group): type %s", args.Type)
		}
	case data.SetVariableValueTypeIAPProduct:
		rhs = 0
		id := value.(int)
		rhs = 0
		if sceneManager.IsUnlocked(id) {
			rhs = 1
		}
	case data.SetVariableValueTypeSystem:
		systemVariableType := value.(data.SystemVariableType)
		switch systemVariableType {
		case data.SystemVariableHintCount:
			rhs = int64(g.hints.ActiveHintCount())
		case data.SystemVariableInterstitialAdsLoaded:
			if sceneManager.InterstitialAdsLoaded() {
				rhs = 1
			}
		case data.SystemVariableRewardedAdsLoaded:
			if sceneManager.RewardedAdsLoaded() {
				rhs = 1
			}
		case data.SystemVariableRoomID:
			rhs = int64(roomID)
		case data.SystemVariableCurrentTime:
			rhs = time.Now().Unix()
		case data.SystemVariableActiveItemID:
			rhs = int64(g.items.ActiveItem())
		case data.SystemVariableEventItemID:
			rhs = int64(g.items.EventItem())
		case data.SystemVariableTriggeredPictureID:
			rhs = int64(g.triggeredPictureID)
		case data.SystemVariablePressedPictureID:
			rhs = int64(g.pressedPictureID)
		case data.SystemVariableReleasedPictureID:
			rhs = int64(g.releasedPictureID)
		case data.SystemVariableSponsorTier:
			rhs = int64(sceneManager.SponsorTier())
		default:
			return 0, fmt.Errorf("gamestate: not implemented yet (set_variable): systemVariableType %s", systemVariableType)
		}
	case data.SetVariableValueTypeTable:
		v := g.InterfaceToTableValue(sceneManager, value)
		i, ok := data.InterfaceToInt(v)
		if !ok {
			return 0, fmt.Errorf("gamestate: table value isn't an integer v", v)
		}

		rhs = int64(i)
	}
	switch op {
	case data.SetVariableOpAssign:
	case data.SetVariableOpAdd:
		rhs = lhs + rhs
	case data.SetVariableOpSub:
		rhs = lhs - rhs
	case data.SetVariableOpMul:
		rhs = lhs * rhs
	case data.SetVariableOpDiv:
		rhs = lhs / rhs
	case data.SetVariableOpMod:
		rhs = lhs % rhs
	default:
		return 0, fmt.Errorf("gamestate: not implemented yet (set_variable): SetVariableOp %s", op)
	}
	return rhs, nil
}

func (g *Game) SetVariable(sceneManager *scene.Manager, variableID int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}, mapID, roomID, eventID int) error {
	lhs := g.VariableValue(variableID)
	rhs, err := g.calcVariableRhs(sceneManager, lhs, op, valueType, value, mapID, roomID, eventID)
	if err != nil {
		return err
	}
	g.variables.SetVariableValue(variableID, rhs)
	return nil
}

func (g *Game) SetVariableRef(sceneManager *scene.Manager, variableID int, op data.SetVariableOp, valueType data.SetVariableValueType, value interface{}, mapID, roomID, eventID int) error {
	lhs := g.VariableValue(int(g.VariableValue(variableID)))
	rhs, err := g.calcVariableRhs(sceneManager, lhs, op, valueType, value, mapID, roomID, eventID)
	if err != nil {
		return err
	}

	g.variables.SetVariableValue(int(g.VariableValue(variableID)), rhs)
	return nil
}

func (g *Game) PauseHint(id int) {
	g.hints.Pause(id)
}

func (g *Game) ActivateHint(id int) {
	g.hints.Activate(id)
}

func (g *Game) CompleteHint(id int) {
	g.hints.Complete(id)
}

func (g *Game) AddItem(id int) {
	g.items.Add(id)
}

func (g *Game) RemoveItem(id int) {
	g.items.Remove(id)
}

func (g *Game) SetEventItem(id int) {
	g.items.SetEventItem(id)
}

func (g *Game) InsertItemBefore(targetItemID int, insertItemID int) {
	g.items.InsertBefore(targetItemID, insertItemID)
}

func (g *Game) SetBackground(mapID, roomID int, image string) {
	if g.backgrounds == nil {
		g.backgrounds = map[int]map[int]string{}
	}
	if _, ok := g.backgrounds[mapID]; !ok {
		g.backgrounds[mapID] = map[int]string{}
	}
	g.backgrounds[mapID][roomID] = image
}

func (g *Game) SetForeground(mapID, roomID int, image string) {
	if g.foregrounds == nil {
		g.foregrounds = map[int]map[int]string{}
	}
	if _, ok := g.foregrounds[mapID]; !ok {
		g.foregrounds[mapID] = map[int]string{}
	}
	g.foregrounds[mapID][roomID] = image
}

func (g *Game) Background(mapID, roomID int) (string, bool) {
	if g.backgrounds != nil {
		if r, ok := g.backgrounds[mapID]; ok {
			if img, ok := r[roomID]; ok {
				return img, true
			}
		}
	}
	return "", false
}

func (g *Game) Foreground(mapID, roomID int) (string, bool) {
	if g.foregrounds != nil {
		if r, ok := g.foregrounds[mapID]; ok {
			if img, ok := r[roomID]; ok {
				return img, true
			}
		}
	}
	return "", false
}

func (g *Game) PlayerSpeed() data.Speed {
	return g.playerSpeed
}

func (g *Game) SetPlayerSpeed(value data.Speed) {
	if value == 0 {
		panic("gamestate: value must not be 0 at SetPlayerSpeed")
	}
	g.playerSpeed = value
}

func (g *Game) ShakeStartGameButton() {
	if g.onShakeStartGameButton != nil {
		g.onShakeStartGameButton()
	}
}

func (g *Game) ShouldShowCredits() bool {
	return g.shouldShowCredits
}

func (g *Game) ShouldShowCreditsCloseButton() bool {
	return g.shouldShowCreditsCloseButton
}

func (g *Game) ShowCredits(shouldShowCreditsCloseButton bool) {
	g.shouldShowCredits = true
	g.shouldShowCreditsCloseButton = shouldShowCreditsCloseButton
}

func (g *Game) ShowedCredits() {
	g.shouldShowCredits = false
}

func (g *Game) touchingPictureID(x, y int) int {
	return g.pictures.TouchingPictureID(x, y)
}

// UpdatePictureTouch updates the picture touch states
// and reports whether one of the pictures is touched.
func (g *Game) UpdatePictureTouch(offsetY int) bool {
	g.resetPictureIDs()
	x, y := input.Position()
	if y < consts.HeaderHeight {
		return false
	}
	if !g.Map().IsBlockingEventExecuting() && !g.Map().IsPlayerMovingByUserInput() {
		sx := x / consts.TileScale
		sy := (y - offsetY) / consts.TileScale
		if g.updatePictureIDs(sx, sy) {
			return true
		}
	}
	return false
}

func (g *Game) resetPictureIDs() {
	g.triggeredPictureID = 0
	g.pressedPictureID = 0
	g.releasedPictureID = 0
}

func (g *Game) updatePictureIDs(x, y int) bool {
	if input.Triggered() {
		g.triggeredPictureID = g.touchingPictureID(x, y)
	}
	if input.Pressed() {
		g.pressedPictureID = g.touchingPictureID(x, y)
	}
	if input.Released() {
		g.releasedPictureID = g.touchingPictureID(x, y)
	}
	return g.triggeredPictureID > 0 || g.releasedPictureID > 0 || g.releasedPictureID > 0
}
