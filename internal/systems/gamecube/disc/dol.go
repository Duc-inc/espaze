package disc

// DOL ("Dolphin executable") is the GameCube's native executable
// format - what a disc's header DOLOffset points to. It's a fixed
// 0x100-byte header of up to 7 code and 11 data section descriptors
// (each an offset/address/size triple), followed by the raw section
// bytes themselves.
const (
	dolNumTextSections = 7
	dolNumDataSections = 11
	dolTotalSections   = dolNumTextSections + dolNumDataSections

	dolOffsetsBase = 0x00
	dolAddrsBase   = 0x48
	dolSizesBase   = 0x90
	dolEntryOffset = 0xE0
)

// Section is one loadable chunk of a DOL: where its bytes live within
// the DOL image, where they belong in memory, and how many bytes.
type Section struct {
	FileOffset uint32
	LoadAddr   uint32
	Size       uint32
}

// DOL holds every section plus the program's entry point.
type DOL struct {
	Sections []Section
	Entry    uint32
}

// ParseDOL reads a DOL executable's header - real hardware's IPL
// (boot ROM) does this itself before jumping to Entry; this project
// has no IPL, so whatever loads a DOL is expected to call LoadInto
// directly instead.
func ParseDOL(image []byte) DOL {
	var sections []Section
	for i := 0; i < dolTotalSections; i++ {
		offset := be32(image[dolOffsetsBase+i*4:])
		addr := be32(image[dolAddrsBase+i*4:])
		size := be32(image[dolSizesBase+i*4:])
		if size == 0 {
			continue
		}
		sections = append(sections, Section{FileOffset: offset, LoadAddr: addr, Size: size})
	}
	return DOL{Sections: sections, Entry: be32(image[dolEntryOffset:])}
}

// Writer is the subset of a memory bus LoadInto needs.
type Writer interface {
	Write8(addr uint32, v byte)
}

// LoadInto copies every section's bytes from the DOL image into
// memory at its load address - the same job real hardware's IPL does
// via the disc drive, done here directly from an in-memory image.
func (d DOL) LoadInto(image []byte, mem Writer) {
	for _, s := range d.Sections {
		for i := uint32(0); i < s.Size; i++ {
			if int(s.FileOffset+i) >= len(image) {
				break
			}
			mem.Write8(s.LoadAddr+i, image[s.FileOffset+i])
		}
	}
}
