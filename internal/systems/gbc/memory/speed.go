package memory

// speedSwitch implements KEY1 ($FF4D). Real hardware only performs the
// switch when a STOP instruction executes after arming it (bit0), which
// this project's shared CPU package (reused as-is from the DMG core)
// has no hook for; this simplifies to switching immediately on the
// arming write instead of waiting for the following STOP.
type speedSwitch struct {
	doubleSpeed bool
}

func (s *speedSwitch) writeKEY1(v byte) {
	if v&0x01 != 0 {
		s.doubleSpeed = !s.doubleSpeed
	}
}

func (s *speedSwitch) readKEY1() byte {
	v := byte(0x7E)
	if s.doubleSpeed {
		v |= 0x80
	}
	return v
}
