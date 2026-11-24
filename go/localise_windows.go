// SPDX-License-Identifier: EUPL-1.2

//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

// systemLanguage is the Windows display language as a locale name
// ("tr-TR"), which detectLanguage falls back to: Windows sets no LANG.
//
//	systemLanguage() // "tr-TR" on a Turkish Windows
func systemLanguage() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	langID, _, _ := kernel32.NewProc("GetUserDefaultUILanguage").Call()
	if langID == 0 {
		return ""
	}
	buf := make([]uint16, 85) // LOCALE_NAME_MAX_LENGTH
	n, _, _ := kernel32.NewProc("LCIDToLocaleName").Call(langID, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}
