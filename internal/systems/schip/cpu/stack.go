package cpu

import "fmt"

const stackDepth = 16

// stack implements the CALL/RET subroutine mechanism.
type stack struct {
	data [stackDepth]uint16
	sp   uint8
}

func (s *stack) push(addr uint16) error {
	if int(s.sp) >= stackDepth {
		return fmt.Errorf("cpu: stack overflow")
	}
	s.data[s.sp] = addr
	s.sp++
	return nil
}

func (s *stack) pop() (uint16, error) {
	if s.sp == 0 {
		return 0, fmt.Errorf("cpu: stack underflow")
	}
	s.sp--
	return s.data[s.sp], nil
}

func (s *stack) reset() {
	*s = stack{}
}
