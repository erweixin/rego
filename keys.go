package rego

import "github.com/gdamore/tcell/v2"

// Key 表示特殊按键
type Key int

// 按键常量
const (
	KeyNone Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEsc
	KeyBackspace
	KeyTab
	KeySpace
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyDelete
	KeyInsert
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH
	KeyCtrlI
	KeyCtrlJ
	KeyCtrlK
	KeyCtrlL
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
)

// Modifiers 表示修饰键
type Modifiers int

const (
	ModNone  Modifiers = 0
	ModShift Modifiers = 1 << iota
	ModCtrl
	ModAlt
)

// convertTcellKey 将 tcell 按键转换为 rego 按键
func convertTcellKey(e *tcell.EventKey) (Key, rune, Modifiers) {
	var mods Modifiers
	if e.Modifiers()&tcell.ModShift != 0 {
		mods |= ModShift
	}
	if e.Modifiers()&tcell.ModCtrl != 0 {
		mods |= ModCtrl
	}
	if e.Modifiers()&tcell.ModAlt != 0 {
		mods |= ModAlt
	}

	switch e.Key() {
	case tcell.KeyUp:
		return KeyUp, 0, mods
	case tcell.KeyDown:
		return KeyDown, 0, mods
	case tcell.KeyLeft:
		return KeyLeft, 0, mods
	case tcell.KeyRight:
		return KeyRight, 0, mods
	case tcell.KeyEnter:
		return KeyEnter, 0, mods
	case tcell.KeyEscape:
		return KeyEsc, 0, mods
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return KeyBackspace, 0, mods
	case tcell.KeyTab:
		return KeyTab, 0, mods
	case tcell.KeyHome:
		return KeyHome, 0, mods
	case tcell.KeyEnd:
		return KeyEnd, 0, mods
	case tcell.KeyPgUp:
		return KeyPageUp, 0, mods
	case tcell.KeyPgDn:
		return KeyPageDown, 0, mods
	case tcell.KeyDelete:
		return KeyDelete, 0, mods
	case tcell.KeyInsert:
		return KeyInsert, 0, mods
	case tcell.KeyF1:
		return KeyF1, 0, mods
	case tcell.KeyF2:
		return KeyF2, 0, mods
	case tcell.KeyF3:
		return KeyF3, 0, mods
	case tcell.KeyF4:
		return KeyF4, 0, mods
	case tcell.KeyF5:
		return KeyF5, 0, mods
	case tcell.KeyF6:
		return KeyF6, 0, mods
	case tcell.KeyF7:
		return KeyF7, 0, mods
	case tcell.KeyF8:
		return KeyF8, 0, mods
	case tcell.KeyF9:
		return KeyF9, 0, mods
	case tcell.KeyF10:
		return KeyF10, 0, mods
	case tcell.KeyF11:
		return KeyF11, 0, mods
	case tcell.KeyF12:
		return KeyF12, 0, mods
	case tcell.KeyCtrlA:
		return KeyCtrlA, 0, mods
	case tcell.KeyCtrlB:
		return KeyCtrlB, 0, mods
	case tcell.KeyCtrlC:
		return KeyCtrlC, 0, mods
	case tcell.KeyCtrlD:
		return KeyCtrlD, 0, mods
	case tcell.KeyCtrlE:
		return KeyCtrlE, 0, mods
	case tcell.KeyCtrlF:
		return KeyCtrlF, 0, mods
	case tcell.KeyCtrlG:
		return KeyCtrlG, 0, mods
	case tcell.KeyCtrlH:
		return KeyCtrlH, 0, mods
	case tcell.KeyCtrlI:
		return KeyCtrlI, 0, mods
	case tcell.KeyCtrlJ:
		return KeyCtrlJ, 0, mods
	case tcell.KeyCtrlK:
		return KeyCtrlK, 0, mods
	case tcell.KeyCtrlL:
		return KeyCtrlL, 0, mods
	case tcell.KeyCtrlN:
		return KeyCtrlN, 0, mods
	case tcell.KeyCtrlO:
		return KeyCtrlO, 0, mods
	case tcell.KeyCtrlP:
		return KeyCtrlP, 0, mods
	case tcell.KeyCtrlQ:
		return KeyCtrlQ, 0, mods
	case tcell.KeyCtrlR:
		return KeyCtrlR, 0, mods
	case tcell.KeyCtrlS:
		return KeyCtrlS, 0, mods
	case tcell.KeyCtrlT:
		return KeyCtrlT, 0, mods
	case tcell.KeyCtrlU:
		return KeyCtrlU, 0, mods
	case tcell.KeyCtrlV:
		return KeyCtrlV, 0, mods
	case tcell.KeyCtrlW:
		return KeyCtrlW, 0, mods
	case tcell.KeyCtrlX:
		return KeyCtrlX, 0, mods
	case tcell.KeyCtrlY:
		return KeyCtrlY, 0, mods
	case tcell.KeyCtrlZ:
		return KeyCtrlZ, 0, mods
	case tcell.KeyRune:
		ru := e.Rune()
		if ru == ' ' {
			return KeySpace, ' ', mods
		}
		return KeyNone, ru, mods
	default:
		return KeyNone, 0, mods
	}
}
