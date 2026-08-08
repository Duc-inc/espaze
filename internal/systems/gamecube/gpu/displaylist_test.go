package gpu

import (
	"testing"
	"time"
)

// callDisplayListBytes builds one CALL_DISPLAY_LIST command: opcode,
// then a 4-byte address and a 4-byte byte-length.
func callDisplayListBytes(addr, size uint32) []byte {
	b := []byte{cmdCallDisplayList}
	b = append(b, byte(addr>>24), byte(addr>>16), byte(addr>>8), byte(addr))
	b = append(b, byte(size>>24), byte(size>>16), byte(size>>8), byte(size))
	return b
}

func TestCallDisplayListExecutesCommandsFromMemory(t *testing.T) {
	cp := New()
	mem := &fakeMemory{}

	sublist := []byte{cmdDrawTriangles, 0x00, 0x03}
	v := Vertex{X: 1, Y: 2, Z: 3, R: 255, G: 255, B: 255, A: 255}
	sublist = append(sublist, vertexBytesOf(v)...)
	sublist = append(sublist, vertexBytesOf(v)...)
	sublist = append(sublist, vertexBytesOf(v)...)
	mem.data = make([]byte, 0x100+len(sublist))
	copy(mem.data[0x100:], sublist)
	cp.SetMemoryReader(mem)

	stream := callDisplayListBytes(0x100, uint32(len(sublist)))
	cp.Execute(stream)

	tris := cp.DrainTriangles()
	if len(tris) != 1 {
		t.Fatalf("got %d triangles, want 1 (from the called display list)", len(tris))
	}
	if tris[0].V0.X != 1 {
		t.Fatalf("V0.X = %d, want 1", tris[0].V0.X)
	}
}

func TestCallDisplayListWithoutMemoryReaderIsNoOp(t *testing.T) {
	cp := New() // no SetMemoryReader
	stream := callDisplayListBytes(0x100, 100)
	cp.Execute(stream) // must not panic

	if len(cp.DrainTriangles()) != 0 {
		t.Fatal("expected no triangles when the display list can't be fetched")
	}
}

func TestCallDisplayListDoesNotRecurseForever(t *testing.T) {
	cp := New()
	// A display list that calls itself - must terminate via the depth
	// cap rather than recursing until a stack overflow.
	selfCall := callDisplayListBytes(0x100, uint32(len(callDisplayListBytes(0, 0))))
	cp.SetMemoryReader(&selfCallingMemory{payload: selfCall})

	done := make(chan struct{})
	go func() {
		cp.Execute(selfCall)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Execute did not terminate - display list recursion guard failed")
	}
}

type selfCallingMemory struct{ payload []byte }

func (m *selfCallingMemory) ReadBytes(addr uint32, length int) []byte { return m.payload }
