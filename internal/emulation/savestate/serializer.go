package savestate

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
)

// Encode serializes an envelope to bytes suitable for storage or transport.
func Encode(env Envelope) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(env); err != nil {
		return nil, fmt.Errorf("savestate: encode: %w", err)
	}
	return buf.Bytes(), nil
}

// Decode reverses Encode, validating the envelope version along the way.
func Decode(data []byte) (Envelope, error) {
	var env Envelope
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&env); err != nil {
		return Envelope{}, fmt.Errorf("savestate: decode: %w", err)
	}
	if env.Version != CurrentVersion {
		return Envelope{}, fmt.Errorf("savestate: unsupported version %d", env.Version)
	}
	return env, nil
}

// SaveToFile encodes and writes an envelope to disk.
func SaveToFile(path string, env Envelope) error {
	data, err := Encode(env)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadFromFile reads and decodes an envelope previously written by SaveToFile.
func LoadFromFile(path string) (Envelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Envelope{}, fmt.Errorf("savestate: read %s: %w", path, err)
	}
	return Decode(data)
}
