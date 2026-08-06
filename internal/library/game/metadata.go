package game

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"
	"unicode"
)

// IDFromPath derives a stable identifier from a ROM's file path, so the
// same file always maps to the same game ID across scans and restarts.
func IDFromPath(path string) string {
	sum := sha1.Sum([]byte(filepath.Clean(path)))
	return hex.EncodeToString(sum[:8])
}

// TitleFromFilename turns "legend_of_something-v2.ch8" into a readable
// "Legend Of Something V2" guess, used until the user renames it.
func TitleFromFilename(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.Map(func(r rune) rune {
		if r == '_' || r == '-' {
			return ' '
		}
		return r
	}, base)
	return strings.TrimSpace(titleCase(base))
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
