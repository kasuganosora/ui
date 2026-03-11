//go:build linux && !android

package wayland

import "github.com/kasuganosora/ui/event"

// evdev key codes (from linux/input-event-codes.h)
const (
	evKeyEsc       = 1
	evKey1         = 2
	evKey2         = 3
	evKey3         = 4
	evKey4         = 5
	evKey5         = 6
	evKey6         = 7
	evKey7         = 8
	evKey8         = 9
	evKey9         = 10
	evKey0         = 11
	evKeyMinus     = 12
	evKeyEqual     = 13
	evKeyBackspace = 14
	evKeyTab       = 15
	evKeyQ         = 16
	evKeyW         = 17
	evKeyE         = 18
	evKeyR         = 19
	evKeyT         = 20
	evKeyY         = 21
	evKeyU         = 22
	evKeyI         = 23
	evKeyO         = 24
	evKeyP         = 25
	evKeyLeftbrace = 26
	evKeyRightbrace = 27
	evKeyEnter     = 28
	evKeyLeftctrl  = 29
	evKeyA         = 30
	evKeyS         = 31
	evKeyD         = 32
	evKeyF         = 33
	evKeyG         = 34
	evKeyH         = 35
	evKeyJ         = 36
	evKeyK         = 37
	evKeyL         = 38
	evKeySemicolon = 39
	evKeyApostrophe = 40
	evKeyGrave     = 41
	evKeyLeftshift = 42
	evKeyBackslash = 43
	evKeyZ         = 44
	evKeyX         = 45
	evKeyC         = 46
	evKeyV         = 47
	evKeyB         = 48
	evKeyN         = 49
	evKeyM         = 50
	evKeyComma     = 51
	evKeyDot       = 52
	evKeySlash     = 53
	evKeyRightshift = 54
	evKeyKpasterisk = 55
	evKeyLeftalt   = 56
	evKeySpace     = 57
	evKeyCapslock  = 58
	evKeyF1        = 59
	evKeyF2        = 60
	evKeyF3        = 61
	evKeyF4        = 62
	evKeyF5        = 63
	evKeyF6        = 64
	evKeyF7        = 65
	evKeyF8        = 66
	evKeyF9        = 67
	evKeyF10       = 68
	evKeyNumlock   = 69
	evKeyScrolllock = 70
	evKeyKp7       = 71
	evKeyKp8       = 72
	evKeyKp9       = 73
	evKeyKpminus   = 74
	evKeyKp4       = 75
	evKeyKp5       = 76
	evKeyKp6       = 77
	evKeyKpplus    = 78
	evKeyKp1       = 79
	evKeyKp2       = 80
	evKeyKp3       = 81
	evKeyKp0       = 82
	evKeyKpdot     = 83
	evKeyF11       = 87
	evKeyF12       = 88
	evKeyKpenter   = 96
	evKeyRightctrl = 97
	evKeyKpslash   = 98
	evKeySysrq     = 99
	evKeyRightalt  = 100
	evKeyHome      = 102
	evKeyUp        = 103
	evKeyPageup    = 104
	evKeyLeft      = 105
	evKeyRight     = 106
	evKeyEnd       = 107
	evKeyDown      = 108
	evKeyPagedown  = 109
	evKeyInsert    = 110
	evKeyDelete    = 111
	evKeyLeftmeta  = 125
	evKeyRightmeta = 126
	evKeyMenu      = 127
	evKeyPause     = 119
)

// evdevToKey maps a Linux evdev keycode (as sent by Wayland wl_keyboard) to event.Key.
func evdevToKey(keycode uint32) event.Key {
	switch keycode {
	case evKeyEsc:
		return event.KeyEscape
	case evKey1:
		return event.Key1
	case evKey2:
		return event.Key2
	case evKey3:
		return event.Key3
	case evKey4:
		return event.Key4
	case evKey5:
		return event.Key5
	case evKey6:
		return event.Key6
	case evKey7:
		return event.Key7
	case evKey8:
		return event.Key8
	case evKey9:
		return event.Key9
	case evKey0:
		return event.Key0
	case evKeyMinus:
		return event.KeyMinus
	case evKeyEqual:
		return event.KeyEqual
	case evKeyBackspace:
		return event.KeyBackspace
	case evKeyTab:
		return event.KeyTab
	case evKeyQ:
		return event.KeyQ
	case evKeyW:
		return event.KeyW
	case evKeyE:
		return event.KeyE
	case evKeyR:
		return event.KeyR
	case evKeyT:
		return event.KeyT
	case evKeyY:
		return event.KeyY
	case evKeyU:
		return event.KeyU
	case evKeyI:
		return event.KeyI
	case evKeyO:
		return event.KeyO
	case evKeyP:
		return event.KeyP
	case evKeyLeftbrace:
		return event.KeyLeftBracket
	case evKeyRightbrace:
		return event.KeyRightBracket
	case evKeyEnter, evKeyKpenter:
		return event.KeyEnter
	case evKeyLeftctrl:
		return event.KeyLeftCtrl
	case evKeyA:
		return event.KeyA
	case evKeyS:
		return event.KeyS
	case evKeyD:
		return event.KeyD
	case evKeyF:
		return event.KeyF
	case evKeyG:
		return event.KeyG
	case evKeyH:
		return event.KeyH
	case evKeyJ:
		return event.KeyJ
	case evKeyK:
		return event.KeyK
	case evKeyL:
		return event.KeyL
	case evKeySemicolon:
		return event.KeySemicolon
	case evKeyApostrophe:
		return event.KeyApostrophe
	case evKeyGrave:
		return event.KeyGraveAccent
	case evKeyLeftshift:
		return event.KeyLeftShift
	case evKeyBackslash:
		return event.KeyBackslash
	case evKeyZ:
		return event.KeyZ
	case evKeyX:
		return event.KeyX
	case evKeyC:
		return event.KeyC
	case evKeyV:
		return event.KeyV
	case evKeyB:
		return event.KeyB
	case evKeyN:
		return event.KeyN
	case evKeyM:
		return event.KeyM
	case evKeyComma:
		return event.KeyComma
	case evKeyDot:
		return event.KeyPeriod
	case evKeySlash:
		return event.KeySlash
	case evKeyRightshift:
		return event.KeyRightShift
	case evKeyKpasterisk:
		return event.KeyNumpadMultiply
	case evKeyLeftalt:
		return event.KeyLeftAlt
	case evKeySpace:
		return event.KeySpace
	case evKeyCapslock:
		return event.KeyCapsLock
	case evKeyF1:
		return event.KeyF1
	case evKeyF2:
		return event.KeyF2
	case evKeyF3:
		return event.KeyF3
	case evKeyF4:
		return event.KeyF4
	case evKeyF5:
		return event.KeyF5
	case evKeyF6:
		return event.KeyF6
	case evKeyF7:
		return event.KeyF7
	case evKeyF8:
		return event.KeyF8
	case evKeyF9:
		return event.KeyF9
	case evKeyF10:
		return event.KeyF10
	case evKeyNumlock:
		return event.KeyNumLock
	case evKeyScrolllock:
		return event.KeyScrollLock
	case evKeyKp7:
		return event.KeyNumpad7
	case evKeyKp8:
		return event.KeyNumpad8
	case evKeyKp9:
		return event.KeyNumpad9
	case evKeyKpminus:
		return event.KeyNumpadSubtract
	case evKeyKp4:
		return event.KeyNumpad4
	case evKeyKp5:
		return event.KeyNumpad5
	case evKeyKp6:
		return event.KeyNumpad6
	case evKeyKpplus:
		return event.KeyNumpadAdd
	case evKeyKp1:
		return event.KeyNumpad1
	case evKeyKp2:
		return event.KeyNumpad2
	case evKeyKp3:
		return event.KeyNumpad3
	case evKeyKp0:
		return event.KeyNumpad0
	case evKeyKpdot:
		return event.KeyNumpadDecimal
	case evKeyF11:
		return event.KeyF11
	case evKeyF12:
		return event.KeyF12
	case evKeyRightctrl:
		return event.KeyRightCtrl
	case evKeyKpslash:
		return event.KeyNumpadDivide
	case evKeySysrq:
		return event.KeyPrintScreen
	case evKeyRightalt:
		return event.KeyRightAlt
	case evKeyHome:
		return event.KeyHome
	case evKeyUp:
		return event.KeyArrowUp
	case evKeyPageup:
		return event.KeyPageUp
	case evKeyLeft:
		return event.KeyArrowLeft
	case evKeyRight:
		return event.KeyArrowRight
	case evKeyEnd:
		return event.KeyEnd
	case evKeyDown:
		return event.KeyArrowDown
	case evKeyPagedown:
		return event.KeyPageDown
	case evKeyInsert:
		return event.KeyInsert
	case evKeyDelete:
		return event.KeyDelete
	case evKeyLeftmeta:
		return event.KeyLeftSuper
	case evKeyRightmeta:
		return event.KeyRightSuper
	case evKeyMenu:
		return event.KeyMenu
	case evKeyPause:
		return event.KeyPause
	}
	return event.KeyUnknown
}

// evdevToRune converts an evdev keycode to the corresponding Unicode rune.
// This is a simplified mapping for US QWERTY layout.
// A full implementation would use xkbcommon to handle arbitrary layouts.
func evdevToRune(keycode uint32, shift bool) rune {
	type mapping struct {
		normal rune
		shifted rune
	}
	mappings := map[uint32]mapping{
		evKey1:          {'1', '!'},
		evKey2:          {'2', '@'},
		evKey3:          {'3', '#'},
		evKey4:          {'4', '$'},
		evKey5:          {'5', '%'},
		evKey6:          {'6', '^'},
		evKey7:          {'7', '&'},
		evKey8:          {'8', '*'},
		evKey9:          {'9', '('},
		evKey0:          {'0', ')'},
		evKeyMinus:      {'-', '_'},
		evKeyEqual:      {'=', '+'},
		evKeyQ:          {'q', 'Q'},
		evKeyW:          {'w', 'W'},
		evKeyE:          {'e', 'E'},
		evKeyR:          {'r', 'R'},
		evKeyT:          {'t', 'T'},
		evKeyY:          {'y', 'Y'},
		evKeyU:          {'u', 'U'},
		evKeyI:          {'i', 'I'},
		evKeyO:          {'o', 'O'},
		evKeyP:          {'p', 'P'},
		evKeyLeftbrace:  {'[', '{'},
		evKeyRightbrace: {']', '}'},
		evKeyA:          {'a', 'A'},
		evKeyS:          {'s', 'S'},
		evKeyD:          {'d', 'D'},
		evKeyF:          {'f', 'F'},
		evKeyG:          {'g', 'G'},
		evKeyH:          {'h', 'H'},
		evKeyJ:          {'j', 'J'},
		evKeyK:          {'k', 'K'},
		evKeyL:          {'l', 'L'},
		evKeySemicolon:  {';', ':'},
		evKeyApostrophe: {'\'', '"'},
		evKeyGrave:      {'`', '~'},
		evKeyBackslash:  {'\\', '|'},
		evKeyZ:          {'z', 'Z'},
		evKeyX:          {'x', 'X'},
		evKeyC:          {'c', 'C'},
		evKeyV:          {'v', 'V'},
		evKeyB:          {'b', 'B'},
		evKeyN:          {'n', 'N'},
		evKeyM:          {'m', 'M'},
		evKeyComma:      {',', '<'},
		evKeyDot:        {'.', '>'},
		evKeySlash:      {'/', '?'},
		evKeySpace:      {' ', ' '},
	}
	if m, ok := mappings[keycode]; ok {
		if shift {
			return m.shifted
		}
		return m.normal
	}
	return 0
}
