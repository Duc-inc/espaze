// fst.go parses a GameCube disc's File System Table (FST): the index
// of every file and directory on the disc, beyond just the boot DOL
// disc.go already handles. This is what a real game's own DVD access
// calls (which this project doesn't emulate - see ipl's own doc
// comment) would use to find any file other than the main executable.
// The format itself - a flat array of fixed 12-byte entries followed
// by a string table, with the root entry's own length field giving
// the total entry count - is documented in public community
// references, independent of any specific emulator's source.
package disc

import "fmt"

const entryBytes = 12

// EntryType distinguishes a file entry from a directory entry.
type EntryType byte

const (
	EntryFile EntryType = 0
	EntryDir  EntryType = 1
)

// Entry is one file or directory in the disc's file system table.
// FileOffset/FileLength are meaningful for EntryFile entries (where a
// file's bytes start on the disc image, and how many); for EntryDir
// entries, FileLength instead holds the index of the entry just past
// this directory's last descendant - real hardware's own "next"
// field, which is how the FST encodes nesting without this project
// needing to build a separate tree structure.
type Entry struct {
	Type       EntryType
	Name       string
	FileOffset uint32
	FileLength uint32
}

// FST holds every entry parsed from a disc's File System Table, in
// their on-disc order (the root directory is always index 0).
type FST struct {
	Entries []Entry
}

// ParseFST reads the File System Table a Header points to.
func ParseFST(image []byte, header Header) (FST, error) {
	base := header.FSTOffset
	if int(base)+entryBytes > len(image) {
		return FST{}, fmt.Errorf("disc: FST offset %#x out of range", base)
	}

	// The root entry (index 0) is always a directory whose own
	// "length" field holds the total number of FST entries.
	entryCount := be32(image[base+8:])
	stringTable := base + entryCount*entryBytes

	entries := make([]Entry, 0, entryCount)
	for i := uint32(0); i < entryCount; i++ {
		off := base + i*entryBytes
		if int(off)+entryBytes > len(image) {
			break
		}
		nameOffset := be32(image[off:]) & 0x00FFFFFF // low 24 bits; top byte is Type
		entries = append(entries, Entry{
			Type:       EntryType(image[off]),
			Name:       readCString(image, stringTable+nameOffset),
			FileOffset: be32(image[off+4:]),
			FileLength: be32(image[off+8:]),
		})
	}
	return FST{Entries: entries}, nil
}

// Lookup returns the first entry whose Name matches, by simple linear
// scan across every entry - this doesn't resolve full nested paths
// through the directory hierarchy Type/FileLength otherwise encode,
// just a flat name match.
func (f FST) Lookup(name string) (Entry, bool) {
	for _, e := range f.Entries {
		if e.Name == name {
			return e, true
		}
	}
	return Entry{}, false
}

func readCString(b []byte, offset uint32) string {
	if int(offset) >= len(b) {
		return ""
	}
	end := offset
	for int(end) < len(b) && b[end] != 0 {
		end++
	}
	return string(b[offset:end])
}
