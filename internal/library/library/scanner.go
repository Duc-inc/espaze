package library

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// ScannedFile is one ROM found on disk, already classified against a
// registered core by its file extension.
type ScannedFile struct {
	Path   string
	System string
}

// Scan walks root recursively and returns every file whose extension
// matches a registered core, per extToSystem (see core.ExtensionIndex).
func Scan(root string, extToSystem map[string]string) ([]ScannedFile, error) {
	var found []ScannedFile

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		system, ok := extToSystem[ext]
		if !ok {
			return nil
		}
		found = append(found, ScannedFile{Path: path, System: system})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}
