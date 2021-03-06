// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"reflect"
	"sort"
	"unicode/utf8"

	"github.com/goki/gi/units"
	"github.com/goki/ki"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/kit"
)

////////////////////////////////////////////////////////////////////////////////////////
// ComboBox for selecting items from a list

type ComboBox struct {
	ButtonBase
	Editable  bool          `xml:"editable" desc:"provide a text field for editing the value, or just a button for selecting items?  Set the editable property"`
	CurVal    interface{}   `json:"-" xml:"-" desc:"current selected value"`
	CurIndex  int           `json:"-" xml:"-" desc:"current index in list of possible items"`
	Items     []interface{} `json:"-" xml:"-" desc:"items available for selection"`
	ItemsMenu Menu          `json:"-" xml:"-" desc:"the menu of actions for selecting items -- automatically generated from Items"`
	ComboSig  ki.Signal     `json:"-" xml:"-" view:"-" desc:"signal for combo box, when a new value has been selected -- the signal type is the index of the selected item, and the data is the value"`
	MaxLength int           `desc:"maximum label length (in runes)"`
}

var KiT_ComboBox = kit.Types.AddType(&ComboBox{}, ComboBoxProps)

var ComboBoxProps = ki.Props{
	"border-width":     units.NewValue(1, units.Px),
	"border-radius":    units.NewValue(4, units.Px),
	"border-color":     &Prefs.Colors.Border,
	"border-style":     BorderSolid,
	"padding":          units.NewValue(4, units.Px),
	"margin":           units.NewValue(4, units.Px),
	"text-align":       AlignCenter,
	"background-color": &Prefs.Colors.Control,
	"color":            &Prefs.Colors.Font,
	"#icon": ki.Props{
		"width":   units.NewValue(1, units.Em),
		"height":  units.NewValue(1, units.Em),
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
		"fill":    &Prefs.Colors.Icon,
		"stroke":  &Prefs.Colors.Font,
	},
	"#label": ki.Props{
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
	},
	"#text": ki.Props{
		"margin":    units.NewValue(1, units.Px),
		"padding":   units.NewValue(1, units.Px),
		"max-width": -1,
		"width":     units.NewValue(12, units.Ch),
	},
	"#indicator": ki.Props{
		"width":          units.NewValue(1.5, units.Ex),
		"height":         units.NewValue(1.5, units.Ex),
		"margin":         units.NewValue(0, units.Px),
		"padding":        units.NewValue(0, units.Px),
		"vertical-align": AlignBottom,
		"fill":           &Prefs.Colors.Icon,
		"stroke":         &Prefs.Colors.Font,
	},
	"#ind-stretch": ki.Props{
		"width": units.NewValue(1, units.Em),
	},
	ButtonSelectors[ButtonActive]: ki.Props{
		"background-color": "linear-gradient(lighter-0, highlight-10)",
	},
	ButtonSelectors[ButtonInactive]: ki.Props{
		"border-color": "highlight-50",
		"color":        "highlight-50",
	},
	ButtonSelectors[ButtonHover]: ki.Props{
		"background-color": "linear-gradient(highlight-10, highlight-10)",
	},
	ButtonSelectors[ButtonFocus]: ki.Props{
		"border-width":     units.NewValue(2, units.Px),
		"background-color": "linear-gradient(samelight-50, highlight-10)",
	},
	ButtonSelectors[ButtonDown]: ki.Props{
		"color":            "highlight-90",
		"background-color": "linear-gradient(highlight-30, highlight-10)",
	},
	ButtonSelectors[ButtonSelected]: ki.Props{
		"background-color": "linear-gradient(pref(Select), highlight-10)",
		"color":            "highlight-90",
	},
}

// ButtonWidget interface

func (cb *ComboBox) ButtonAsBase() *ButtonBase {
	return &(cb.ButtonBase)
}

func (cb *ComboBox) ButtonRelease() {
	if cb.IsInactive() {
		return
	}
	wasPressed := (cb.State == ButtonDown)
	cb.MakeItemsMenu()
	if len(cb.ItemsMenu) == 0 {
		return
	}
	updt := cb.UpdateStart()
	cb.SetButtonState(ButtonActive)
	cb.ButtonSig.Emit(cb.This(), int64(ButtonReleased), nil)
	if wasPressed {
		cb.ButtonSig.Emit(cb.This(), int64(ButtonClicked), nil)
	}
	cb.UpdateEnd(updt)
	pos := cb.WinBBox.Max
	if pos.X == 0 && pos.Y == 0 { // offscreen
		pos = cb.ObjBBox.Max
	}
	indic, ok := cb.Parts.ChildByName("indicator", 3)
	if ok {
		pos = KiToNode2DBase(indic).WinBBox.Min
		if pos.X == 0 && pos.Y == 0 {
			pos = KiToNode2DBase(indic).ObjBBox.Min
		}
	} else {
		pos.Y -= 10
		pos.X -= 10
	}
	PopupMenu(cb.ItemsMenu, pos.X, pos.Y, cb.Viewport, cb.Text)
}

// ConfigPartsIconText returns a standard config for creating parts, of icon
// and text left-to right in a row -- always makes text
func (cb *ComboBox) ConfigPartsIconText(config *kit.TypeAndNameList, icnm string) (icIdx, txIdx int) {
	// todo: add some styles for button layout
	icIdx = -1
	txIdx = -1
	if IconName(icnm).IsValid() {
		icIdx = len(*config)
		config.Add(KiT_Icon, "icon")
		config.Add(KiT_Space, "space")
	}
	txIdx = len(*config)
	config.Add(KiT_TextField, "text")
	return
}

// ConfigPartsSetText sets part style props, using given props if not set in
// object props
func (cb *ComboBox) ConfigPartsSetText(txt string, txIdx, icIdx, indIdx int) {
	if txIdx >= 0 {
		tx := cb.Parts.KnownChild(txIdx).(*TextField)
		tx.SetText(txt)
		if _, ok := tx.Prop("__comboInit"); !ok {
			cb.StylePart(Node2D(tx))
			if icIdx >= 0 {
				cb.StylePart(cb.Parts.KnownChild(txIdx - 1).(Node2D)) // also get the space
			}
			tx.SetProp("__comboInit", true)
			if cb.MaxLength > 0 {
				tx.SetMinPrefWidth(units.NewValue(float32(cb.MaxLength), units.Ch))
			}
			if indIdx > 0 {
				ispc := cb.Parts.KnownChild(indIdx - 1).(Node2D)
				ispc.SetProp("max-width", 0)
			}
		}
	}
}

// ConfigPartsAddIndicatorSpace adds indicator with a space instead of a stretch
// for editable combobox, where textfield then takes up the rest of the space
func (bb *ButtonBase) ConfigPartsAddIndicatorSpace(config *kit.TypeAndNameList, defOn bool) int {
	needInd := (bb.HasMenu() || defOn) && bb.Indicator != "none"
	if !needInd {
		return -1
	}
	indIdx := -1
	config.Add(KiT_Space, "ind-stretch")
	indIdx = len(*config)
	config.Add(KiT_Icon, "indicator")
	return indIdx
}

func (cb *ComboBox) ConfigPartsIfNeeded() {
	if cb.Editable {
		_, ok := cb.Parts.ChildByName("text", 2)
		if !cb.PartsNeedUpdateIconLabel(string(cb.Icon), "") && ok {
			return
		}
	} else {
		if !cb.PartsNeedUpdateIconLabel(string(cb.Icon), cb.Text) {
			return
		}
	}
	cb.This().(ButtonWidget).ConfigParts()
}

func (cb *ComboBox) ConfigParts() {
	if eb, ok := cb.Prop("editable"); ok {
		cb.Editable, _ = kit.ToBool(eb)
	}
	config := kit.TypeAndNameList{}
	var icIdx, lbIdx, txIdx, indIdx int
	if cb.Editable {
		lbIdx = -1
		icIdx, txIdx = cb.ConfigPartsIconText(&config, string(cb.Icon))
		cb.SetProp("no-focus", true)
		indIdx = cb.ConfigPartsAddIndicatorSpace(&config, true) // use space instead of stretch
	} else {
		txIdx = -1
		icIdx, lbIdx = cb.ConfigPartsIconLabel(&config, string(cb.Icon), cb.Text)
		indIdx = cb.ConfigPartsAddIndicator(&config, true) // default on
	}
	mods, updt := cb.Parts.ConfigChildren(config, false) // not unique names
	cb.ConfigPartsSetIconLabel(string(cb.Icon), cb.Text, icIdx, lbIdx)
	cb.ConfigPartsIndicator(indIdx)
	if txIdx >= 0 {
		cb.ConfigPartsSetText(cb.Text, txIdx, icIdx, indIdx)
	}
	if cb.MaxLength > 0 && lbIdx >= 0 {
		lbl := cb.Parts.KnownChild(lbIdx).(*Label)
		lbl.SetMinPrefWidth(units.NewValue(float32(cb.MaxLength), units.Ch))
	}
	if mods {
		cb.UpdateEnd(updt)
	}
}

// TextField returns the text field of an editable combobox, and false if not made
func (cb *ComboBox) TextField() (*TextField, bool) {
	tff, ok := cb.Parts.ChildByName("text", 2)
	if !ok {
		return nil, ok
	}
	return tff.(*TextField), ok
}

// MakeItems makes sure the Items list is made, and if not, or reset is true,
// creates one with the given capacity
func (cb *ComboBox) MakeItems(reset bool, capacity int) {
	if cb.Items == nil || reset {
		cb.Items = make([]interface{}, 0, capacity)
	}
}

// SortItems sorts the items according to their labels
func (cb *ComboBox) SortItems(ascending bool) {
	sort.Slice(cb.Items, func(i, j int) bool {
		if ascending {
			return ToLabel(cb.Items[i]) < ToLabel(cb.Items[j])
		} else {
			return ToLabel(cb.Items[i]) > ToLabel(cb.Items[j])
		}
	})
}

// SetToMaxLength gets the maximum label length so that the width of the
// button label is automatically set according to the max length of all items
// in the list -- if maxLen > 0 then it is used as an upper do-not-exceed
// length
func (cb *ComboBox) SetToMaxLength(maxLen int) {
	ml := 0
	for _, it := range cb.Items {
		ml = ints.MaxInt(ml, utf8.RuneCountInString(ToLabel(it)))
	}
	if maxLen > 0 {
		ml = ints.MinInt(ml, maxLen)
	}
	cb.MaxLength = ml
}

// ItemsFromTypes sets the Items list from a list of types -- see e.g.,
// AllImplementersOf or AllEmbedsOf in kit.TypeRegistry -- if setFirst then
// set current item to the first item in the list, sort sorts the list in
// ascending order, and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit
func (cb *ComboBox) ItemsFromTypes(tl []reflect.Type, setFirst, sort bool, maxLen int) {
	sz := len(tl)
	if sz == 0 {
		return
	}
	cb.Items = make([]interface{}, sz)
	for i, typ := range tl {
		cb.Items[i] = typ
	}
	if sort {
		cb.SortItems(true)
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

// ItemsFromStringList sets the Items list from a list of string values -- if
// setFirst then set current item to the first item in the list, and maxLen if
// > 0 auto-sets the width of the button to the contents, with the given upper
// limit
func (cb *ComboBox) ItemsFromStringList(el []string, setFirst bool, maxLen int) {
	sz := len(el)
	if sz == 0 {
		return
	}
	cb.Items = make([]interface{}, sz)
	for i, str := range el {
		cb.Items[i] = str
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

// ItemsFromEnumList sets the Items list from a list of enum values (see
// kit.EnumRegistry) -- if setFirst then set current item to the first item in
// the list, and maxLen if > 0 auto-sets the width of the button to the
// contents, with the given upper limit
func (cb *ComboBox) ItemsFromEnumList(el []kit.EnumValue, setFirst bool, maxLen int) {
	sz := len(el)
	if sz == 0 {
		return
	}
	cb.Items = make([]interface{}, sz)
	for i, enum := range el {
		cb.Items[i] = enum
	}
	if maxLen > 0 {
		cb.SetToMaxLength(maxLen)
	}
	if setFirst {
		cb.SetCurIndex(0)
	}
}

// ItemsFromEnum sets the Items list from an enum type, which must be
// registered on kit.EnumRegistry -- if setFirst then set current item to the
// first item in the list, and maxLen if > 0 auto-sets the width of the button
// to the contents, with the given upper limit -- see kit.EnumRegistry, and
// maxLen if > 0 auto-sets the width of the button to the contents, with the
// given upper limit
func (cb *ComboBox) ItemsFromEnum(enumtyp reflect.Type, setFirst bool, maxLen int) {
	cb.ItemsFromEnumList(kit.Enums.TypeValues(enumtyp, true), setFirst, maxLen)
}

// FindItem finds an item on list of items and returns its index
func (cb *ComboBox) FindItem(it interface{}) int {
	if cb.Items == nil {
		return -1
	}
	for i, v := range cb.Items {
		if v == it {
			return i
		}
	}
	return -1
}

// SetCurVal sets the current value (CurVal) and the corresponding CurIndex
// for that item on the current Items list (adds to items list if not found)
// -- returns that index -- and sets the text to the string value of that
// value (using standard Stringer string conversion)
func (cb *ComboBox) SetCurVal(it interface{}) int {
	cb.CurVal = it
	cb.CurIndex = cb.FindItem(it)
	if cb.CurIndex < 0 { // add to list if not found..
		cb.CurIndex = len(cb.Items)
		cb.Items = append(cb.Items, it)
	}
	cb.SetText(ToLabel(it))
	return cb.CurIndex
}

// SetCurIndex sets the current index (CurIndex) and the corresponding CurVal
// for that item on the current Items list (-1 if not found) -- returns value
// -- and sets the text to the string value of that value (using standard
// Stringer string conversion)
func (cb *ComboBox) SetCurIndex(idx int) interface{} {
	cb.CurIndex = idx
	if idx < 0 || idx >= len(cb.Items) {
		cb.CurVal = nil
		cb.SetText(fmt.Sprintf("idx %v > len", idx))
	} else {
		cb.CurVal = cb.Items[idx]
		cb.SetText(ToLabel(cb.CurVal))
	}
	return cb.CurVal
}

// SelectItem selects a given item and emits the index as the ComboSig signal
// and the selected item as the data
func (cb *ComboBox) SelectItem(idx int) {
	if cb.This() == nil {
		return
	}
	updt := cb.UpdateStart()
	cb.SetCurIndex(idx)
	cb.ComboSig.Emit(cb.This(), int64(cb.CurIndex), cb.CurVal)
	cb.UpdateEnd(updt)
}

// MakeItemsMenu makes menu of all the items
func (cb *ComboBox) MakeItemsMenu() {
	nitm := len(cb.Items)
	if cb.ItemsMenu == nil {
		cb.ItemsMenu = make(Menu, 0, nitm)
	}
	sz := len(cb.ItemsMenu)
	if nitm < sz {
		cb.ItemsMenu = cb.ItemsMenu[0:nitm]
	}
	for i, it := range cb.Items {
		var ac *Action
		if sz > i {
			ac = cb.ItemsMenu[i].(*Action)
		} else {
			ac = &Action{}
			ac.Init(ac)
			cb.ItemsMenu = append(cb.ItemsMenu, ac.This().(Node2D))
		}
		txt := ToLabel(it)
		nm := fmt.Sprintf("Item_%v", i)
		ac.SetName(nm)
		ac.Text = txt
		ac.Data = i // index is the data
		ac.SetSelectedState(i == cb.CurIndex)
		ac.SetAsMenu()
		ac.ActionSig.ConnectOnly(cb.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
			idx := data.(int)
			cb := recv.(*ComboBox)
			cb.SelectItem(idx)
		})
	}
}
