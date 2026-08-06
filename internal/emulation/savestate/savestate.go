package savestate

import "time"

// CurrentVersion is bumped whenever the Envelope layout changes shape.
const CurrentVersion = 1

// Envelope wraps a core-specific save blob with just enough context to
// know which system and core version can safely load it back.
type Envelope struct {
	System    string
	Version   int
	CreatedAt time.Time
	CoreData  []byte
}

// Wrap packages raw core state (produced by Core.SaveState) into an
// envelope ready to be written to disk or handed to the frontend.
func Wrap(system string, coreData []byte) Envelope {
	return Envelope{
		System:    system,
		Version:   CurrentVersion,
		CreatedAt: time.Now(),
		CoreData:  coreData,
	}
}
