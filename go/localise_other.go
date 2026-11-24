// SPDX-License-Identifier: EUPL-1.2

//go:build !windows

package i18n

// systemLanguage: elsewhere the environment (LANG, LC_ALL…) already says,
// and detectLanguage reads it first. See localise_windows.go.
func systemLanguage() string { return "" }
