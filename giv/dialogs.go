// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package giv

import (
	"image"

	"github.com/goki/gi/gi"
	"github.com/goki/gi/units"
	"github.com/goki/ki"
)

// DlgOpts are the basic dialog options accepted by all giv dialog methods --
// provides a named, optional way to specify these args
type DlgOpts struct {
	Title      string    `desc:"generally should be provided -- used for setting name of dialog and associated window"`
	Prompt     string    `desc:"optional more detailed description of what is being requested and how it will be used -- is word-wrapped and can contain full html formatting etc."`
	CSS        ki.Props  `desc:"optional style properties applied to dialog -- can be used to customize any aspect of existing dialogs"`
	TmpSave    ValueView `desc:"value view that needs to have SaveTmp called on it whenever a change is made to one of the underlying values -- pass this down to any sub-views created from a parent"`
	Ok         bool      `desc:"display the Ok button, in most View dialogs where it otherwise is not shown by default -- these views always apply edits immediately, and typically this obviates the need for Ok and Cancel, but sometimes you're giving users a temporary object to edit, and you want them to indicate if they want to proceed or not."`
	Cancel     bool      `desc:"display the Cancel button, in most View dialogs where it otherwise is not shown by default -- these views always apply edits immediately, and typically this obviates the need for Ok and Cancel, but sometimes you're giving users a temporary object to edit, and you want them to indicate if they want to proceed or not."`
	AddOnly    bool      `desc:"can the user delete elements of the slice"`
	DeleteOnly bool      `desc:"can the user add elements to the slice"`
	Inactive   bool      `desc:"if true all fields will be inactive"`
}

// ToGiOpts converts giv opts to gi opts
func (d *DlgOpts) ToGiOpts() gi.DlgOpts {
	return gi.DlgOpts{Title: d.Title, Prompt: d.Prompt, CSS: d.CSS}
}

// StructViewDialog is for editing fields of a structure using a StructView --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore)
func StructViewDialog(avp *gi.Viewport2D, stru interface{}, opts DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_StructView, prIdx+1, "struct-view").(*StructView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	if opts.Inactive {
		sv.SetInactive()
	}
	sv.SetStruct(stru, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		MainMenuView(stru, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}

// MapViewDialog is for editing elements of a map using a MapView -- optionally
// connects to given signal receiving object and function for dialog signals
// (nil to ignore)
func MapViewDialog(avp *gi.Viewport2D, mp interface{}, opts DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_MapView, prIdx+1, "map-view").(*MapView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetMap(mp, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		MainMenuView(mp, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}

// SliceViewDialog for editing elements of a slice using a SliceView --
// optionally connects to given signal receiving object and function for
// dialog signals (nil to ignore).    Also has an optional styling
// function for styling elements of the table.
func SliceViewDialog(avp *gi.Viewport2D, slice interface{}, opts DlgOpts, styleFunc SliceViewStyleFunc, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_SliceView, prIdx+1, "slice-view").(*SliceView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetInactiveState(false)
	sv.StyleFunc = styleFunc
	sv.DeleteOnly = opts.DeleteOnly
	sv.AddOnly = opts.AddOnly
	sv.SetSlice(slice, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(50, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		MainMenuView(slice, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}

// SliceViewSelectDialog for selecting one row from given slice -- connections
// functions available for both the widget signal reporting selection events,
// and the overall dialog signal.  Also has an optional styling function for
// styling elements of the table.
func SliceViewSelectDialog(avp *gi.Viewport2D, slice, curVal interface{}, opts DlgOpts, styleFunc SliceViewStyleFunc, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	if opts.CSS == nil {
		opts.CSS = ki.Props{
			"textfield": ki.Props{
				":inactive": ki.Props{
					"background-color": &gi.Prefs.Colors.Control,
				},
			},
		}
	}
	dlg := gi.NewStdDialog(opts.ToGiOpts(), true, true)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_SliceView, prIdx+1, "slice-view").(*SliceView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetInactiveState(true)
	sv.StyleFunc = styleFunc
	sv.SelVal = curVal
	sv.SetSlice(slice, nil)

	sv.SliceViewSig.Connect(dlg.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(SliceViewDoubleClicked) {
			ddlg := recv.Embed(gi.KiT_Dialog).(*gi.Dialog)
			ddlg.Accept()
		}
	})

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(50, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, nil)
	return dlg
}

// SliceViewSelectDialogValue gets the index of the selected item (-1 if nothing selected)
func SliceViewSelectDialogValue(dlg *gi.Dialog) int {
	frame := dlg.Frame()
	sv, ok := frame.ChildByName("slice-view", 0)
	if ok {
		svv := sv.(*SliceView)
		return svv.SelectedIdx
	}
	return -1
}

// TableViewDialog is for editing fields of a slice-of-struct using a
// TableView -- optionally connects to given signal receiving object and
// function for dialog signals (nil to ignore).  Also has an optional styling
// function for styling elements of the table.
func TableViewDialog(avp *gi.Viewport2D, slcOfStru interface{}, opts DlgOpts, styleFunc TableViewStyleFunc, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), opts.Ok, opts.Cancel)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_TableView, prIdx+1, "tableview").(*TableView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetInactiveState(false)
	sv.StyleFunc = styleFunc
	sv.SetSlice(slcOfStru, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(50, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, func() {
		MainMenuView(slcOfStru, dlg.Win, dlg.Win.MainMenu)
	})
	return dlg
}

// TableViewSelectDialog is for selecting a row from a slice-of-struct using a
// TableView -- optionally connects to given signal receiving object and
// functions for signals (nil to ignore): selFunc for the widget signal
// reporting selection events, and dlgFunc for the overall dialog signals.
// Also has an optional styling function for styling elements of the table.
func TableViewSelectDialog(avp *gi.Viewport2D, slcOfStru interface{}, opts DlgOpts, initRow int, styleFunc TableViewStyleFunc, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	if opts.CSS == nil {
		opts.CSS = ki.Props{
			"textfield": ki.Props{
				":inactive": ki.Props{
					"background-color": &gi.Prefs.Colors.Control,
				},
			},
		}
	}
	dlg := gi.NewStdDialog(opts.ToGiOpts(), true, true)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_TableView, prIdx+1, "tableview").(*TableView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetInactiveState(true)
	sv.StyleFunc = styleFunc
	sv.SelectedIdx = initRow
	sv.SetSlice(slcOfStru, nil)

	sv.TableViewSig.Connect(dlg.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(TableViewDoubleClicked) {
			ddlg := recv.Embed(gi.KiT_Dialog).(*gi.Dialog)
			ddlg.Accept()
		}
	})

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(50, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, nil)
	return dlg
}

// TableViewSelectDialogValue gets the index of the selected item (-1 if nothing selected)
func TableViewSelectDialogValue(dlg *gi.Dialog) int {
	frame := dlg.Frame()
	sv, ok := frame.ChildByName("tableview", 0)
	if ok {
		svv := sv.(*TableView)
		rval := svv.SelectedIdx
		return rval
	}
	return -1
}

// show fonts in a bigger size so you can actually see the differences
var FontChooserSize = 18
var FontChooserSizeDots = 18

// FontChooserDialog for choosing a font -- the recv and func signal receivers
// if non-nil are connected to the selection signal for the struct table view,
// so they are updated with that
func FontChooserDialog(avp *gi.Viewport2D, opts DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	FontChooserSizeDots = int(avp.Sty.UnContext.ToDots(float32(FontChooserSize), units.Pt))
	gi.FontLibrary.OpenAllFonts(FontChooserSizeDots)
	dlg := TableViewSelectDialog(avp, &gi.FontLibrary.FontInfo, opts, -1, FontInfoStyleFunc, recv, dlgFunc)
	return dlg
}

func FontInfoStyleFunc(tv *TableView, slice interface{}, widg gi.Node2D, row, col int, vv ValueView) {
	if col == 4 {
		finf, ok := slice.([]gi.FontInfo)
		if ok {
			widg.SetProp("font-family", (finf)[row].Name)
			widg.SetProp("font-stretch", (finf)[row].Stretch)
			widg.SetProp("font-weight", (finf)[row].Weight)
			widg.SetProp("font-style", (finf)[row].Style)
			widg.SetProp("font-size", units.NewValue(float32(FontChooserSize), units.Pt))
		}
	}
}

// IconChooserDialog for choosing an Icon -- the recv and fun signal receivers
// if non-nil are connected to the selection signal for the slice view, and
// the dialog signal.
func IconChooserDialog(avp *gi.Viewport2D, curIc gi.IconName, opts DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	if opts.CSS == nil {
		opts.CSS = ki.Props{
			"icon": ki.Props{
				"width":  units.NewValue(2, units.Em),
				"height": units.NewValue(2, units.Em),
			},
		}
	}
	dlg := SliceViewSelectDialog(avp, &gi.CurIconList, curIc, opts, IconChooserStyleFunc, recv, dlgFunc)
	return dlg
}

func IconChooserStyleFunc(sv *SliceView, slice interface{}, widg gi.Node2D, row int, vv ValueView) {
	ic, ok := slice.([]gi.IconName)
	if ok {
		widg.(*gi.Action).SetText(string(ic[row]))
		widg.SetProp("max-width", -1)
	}
}

// ColorViewDialog for editing a color using a ColorView -- optionally
// connects to given signal receiving object and function for dialog signals
// (nil to ignore)
func ColorViewDialog(avp *gi.Viewport2D, clr gi.Color, opts DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), true, true)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_ColorView, prIdx+1, "color-view").(*ColorView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetColor(clr, opts.TmpSave)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, nil)
	return dlg
}

// ColorViewDialogValue gets the color from the dialog
func ColorViewDialogValue(dlg *gi.Dialog) gi.Color {
	frame := dlg.Frame()
	cvvvk, ok := frame.Children().ElemByType(KiT_ColorView, true, 2)
	if ok {
		cvvv := cvvvk.(*ColorView)
		return cvvv.Color
	}
	return gi.Color{}
}

// FileViewDialog is for selecting / manipulating files -- ext is one or more
// (comma separated) extensions -- files with those will be highlighted
// (include the . at the start of the extension).  recv and dlgFunc connect to the
// dialog signal: if signal value is gi.DialogAccepted use FileViewDialogValue
// to get the resulting selected file.  The optional filterFunc can filter
// files shown in the view -- e.g., FileViewDirOnlyFilter (for only showing
// directories) and FileViewExtOnlyFilter (for only showing directories).
func FileViewDialog(avp *gi.Viewport2D, filename, ext string, opts DlgOpts, filterFunc FileViewFilterFunc, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), true, true)
	dlg.SetName("file-view") // use a consistent name for consistent sizing / placement

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	fv := frame.InsertNewChild(KiT_FileView, prIdx+1, "file-view").(*FileView)
	fv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	fv.FilterFunc = filterFunc
	fv.SetFilename(filename, ext)

	fv.FileSig.Connect(dlg.This(), func(recv, send ki.Ki, sig int64, data interface{}) {
		if sig == int64(FileViewDoubleClicked) {
			ddlg := recv.Embed(gi.KiT_Dialog).(*gi.Dialog)
			ddlg.Accept()
		}
	})

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(35, units.Em))
	dlg.DefSize = image.Point{600, 400} // avoids expensive computation
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, nil)
	return dlg
}

// FileViewDialogValue gets the full path of selected file
func FileViewDialogValue(dlg *gi.Dialog) string {
	frame := dlg.Frame()
	fvk, ok := frame.ChildByName("file-view", 0)
	if ok {
		fv := fvk.(*FileView)
		fn := fv.SelectedFile()
		return fn
	}
	return ""
}

// ArgViewDialog for editing args for a method call in the MethView system
func ArgViewDialog(avp *gi.Viewport2D, args []ArgData, opts DlgOpts, recv ki.Ki, dlgFunc ki.RecvFunc) *gi.Dialog {
	dlg := gi.NewStdDialog(opts.ToGiOpts(), true, true)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	sv := frame.InsertNewChild(KiT_ArgView, prIdx+1, "arg-view").(*ArgView)
	sv.Viewport = dlg.Embed(gi.KiT_Viewport2D).(*gi.Viewport2D)
	sv.SetInactiveState(false)
	sv.SetArgs(args)

	if recv != nil && dlgFunc != nil {
		dlg.DialogSig.Connect(recv, dlgFunc)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp, nil)
	return dlg
}
