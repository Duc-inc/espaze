package cpu

import "sync"

type executeFunc func(c *CPU, opcode uint16) int

// pattern is one instruction encoding: match against opcode&mask, the
// same "mask + match bit pattern" scheme every serious 68000 emulator
// builds its dispatch table from, since the encoding isn't regular
// enough to decode live the way this project's other CPU cores do.
type pattern struct {
	mask, match uint16
	execute     executeFunc
}

type dispatchEntry struct {
	execute executeFunc
}

var (
	dispatchTable [65536]dispatchEntry
	dispatchOnce  sync.Once
)

// buildDispatchTable fills dispatchTable from every opcodes_*.go file's
// pattern list, checked in registration order so more-specific patterns
// (registered first, by convention) win over more-general ones that
// might otherwise also match the same bits.
func buildDispatchTable() {
	// Narrower, more specific patterns are registered first, since a few
	// of them (ANDI/ORI/EORI to CCR/SR, MOVE to/from SR) land inside the
	// numeric range a more general pattern in a later group would also
	// match - the first match in this list wins.
	var patterns []pattern
	patterns = append(patterns, systemOpcodes()...)
	patterns = append(patterns, controlOpcodes()...)
	patterns = append(patterns, bitOpcodes()...)
	patterns = append(patterns, shiftOpcodes()...)
	// muldivOpcodes' patterns are narrow (exact opmode 011/111) slices
	// out of the same 1100xxx/1000xxx top nibbles AND/OR also use with a
	// much broader opmode match - it has to win first, or AND/OR would
	// swallow every MULU/MULS/DIVU/DIVS opcode as themselves instead.
	patterns = append(patterns, muldivOpcodes()...)
	patterns = append(patterns, logicOpcodes()...)
	patterns = append(patterns, compareOpcodes()...)
	patterns = append(patterns, arithmeticOpcodes()...)
	patterns = append(patterns, moveOpcodes()...)

	for opcode := 0; opcode < 65536; opcode++ {
		op := uint16(opcode)
		for _, p := range patterns {
			if op&p.mask == p.match {
				dispatchTable[opcode] = dispatchEntry{execute: p.execute}
				break
			}
		}
	}
}

func ensureDispatchTable() {
	dispatchOnce.Do(buildDispatchTable)
}
