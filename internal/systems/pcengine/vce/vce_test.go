package vce

import "testing"

func TestPaletteWriteAndResolveRoundTrip(t *testing.T) {
	v := New()
	v.WriteAddressLow(0x05)
	v.WriteAddressHigh(0x00)
	v.WriteDataLow(0x07) // R=7,G=0 low bits
	v.WriteDataHigh(0x00)

	r, _, _ := v.Resolve(5)
	if r == 0 {
		t.Fatal("expected a non-zero red channel after writing R=7 into the palette")
	}
}

func TestDataHighWriteAutoIncrementsAddress(t *testing.T) {
	v := New()
	v.WriteAddressLow(0x00)
	v.WriteAddressHigh(0x00)
	v.WriteDataLow(0xFF)
	v.WriteDataHigh(0x01) // completes color 0, should advance to index 1

	v.WriteDataLow(0x00)
	v.WriteDataHigh(0x00) // writes color 1 to black

	_, g, _ := v.Resolve(0)
	if g == 0 {
		t.Fatal("expected color 0's green channel to be non-zero (index should not have advanced before the first high-byte write)")
	}
}
