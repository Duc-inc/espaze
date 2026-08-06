package cpu

import (
	"fmt"

	"github.com/Duc-inc/espaze/internal/systems/chip8/memory"
)

// execute decodes and runs a single 16-bit opcode, following the standard
// CHIP-8 instruction set (Cowgod's reference is the canonical spec here).
func (c *CPU) execute(opcode uint16) error {
	x := byte((opcode & 0x0F00) >> 8)
	y := byte((opcode & 0x00F0) >> 4)
	n := byte(opcode & 0x000F)
	kk := byte(opcode & 0x00FF)
	nnn := opcode & 0x0FFF

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0:
			c.disp.Clear()
		case 0x00EE:
			addr, err := c.stack.pop()
			if err != nil {
				return err
			}
			c.regs.PC = addr
		}
	case 0x1000:
		c.regs.PC = nnn
	case 0x2000:
		if err := c.stack.push(c.regs.PC); err != nil {
			return err
		}
		c.regs.PC = nnn
	case 0x3000:
		if c.regs.V[x] == kk {
			c.regs.PC += 2
		}
	case 0x4000:
		if c.regs.V[x] != kk {
			c.regs.PC += 2
		}
	case 0x5000:
		if c.regs.V[x] == c.regs.V[y] {
			c.regs.PC += 2
		}
	case 0x6000:
		c.regs.V[x] = kk
	case 0x7000:
		c.regs.V[x] += kk
	case 0x8000:
		c.execute8(x, y, n)
	case 0x9000:
		if c.regs.V[x] != c.regs.V[y] {
			c.regs.PC += 2
		}
	case 0xA000:
		c.regs.I = nnn
	case 0xB000:
		c.regs.PC = nnn + uint16(c.regs.V[0])
	case 0xC000:
		c.regs.V[x] = byte(c.rand.Intn(256)) & kk
	case 0xD000:
		c.drawSprite(x, y, n)
	case 0xE000:
		c.executeE(x, kk)
	case 0xF000:
		return c.executeF(x, kk)
	default:
		return fmt.Errorf("cpu: unknown opcode 0x%04X", opcode)
	}
	return nil
}

func (c *CPU) execute8(x, y, n byte) {
	vx, vy := c.regs.V[x], c.regs.V[y]
	switch n {
	case 0x0:
		c.regs.V[x] = vy
	case 0x1:
		c.regs.V[x] = vx | vy
	case 0x2:
		c.regs.V[x] = vx & vy
	case 0x3:
		c.regs.V[x] = vx ^ vy
	case 0x4:
		sum := uint16(vx) + uint16(vy)
		c.regs.V[x] = byte(sum)
		c.regs.V[0xF] = boolToByte(sum > 0xFF)
	case 0x5:
		c.regs.V[x] = vx - vy
		c.regs.V[0xF] = boolToByte(vx >= vy)
	case 0x6:
		c.regs.V[0xF] = vx & 0x1
		c.regs.V[x] = vx >> 1
	case 0x7:
		c.regs.V[x] = vy - vx
		c.regs.V[0xF] = boolToByte(vy >= vx)
	case 0xE:
		c.regs.V[0xF] = (vx & 0x80) >> 7
		c.regs.V[x] = vx << 1
	}
}

func (c *CPU) executeE(x byte, kk byte) {
	switch kk {
	case 0x9E:
		if c.keys.IsDown(c.regs.V[x]) {
			c.regs.PC += 2
		}
	case 0xA1:
		if !c.keys.IsDown(c.regs.V[x]) {
			c.regs.PC += 2
		}
	}
}

func (c *CPU) executeF(x byte, kk byte) error {
	switch kk {
	case 0x07:
		c.regs.V[x] = c.delay.Get()
	case 0x0A:
		vx := x
		c.waitVX = &vx
	case 0x15:
		c.delay.Set(c.regs.V[x])
	case 0x18:
		c.sound.Set(c.regs.V[x])
	case 0x1E:
		c.regs.I += uint16(c.regs.V[x])
	case 0x29:
		c.regs.I = memory.FontStart + uint16(c.regs.V[x])*5
	case 0x33:
		v := c.regs.V[x]
		c.mem.Write(c.regs.I, v/100)
		c.mem.Write(c.regs.I+1, (v/10)%10)
		c.mem.Write(c.regs.I+2, v%10)
	case 0x55:
		for i := byte(0); i <= x; i++ {
			c.mem.Write(c.regs.I+uint16(i), c.regs.V[i])
		}
	case 0x65:
		for i := byte(0); i <= x; i++ {
			c.regs.V[i] = c.mem.Read(c.regs.I + uint16(i))
		}
	default:
		return fmt.Errorf("cpu: unknown Fx opcode 0xF%02X", kk)
	}
	return nil
}

func (c *CPU) drawSprite(x, y, n byte) {
	sprite := make([]byte, n)
	for i := byte(0); i < n; i++ {
		sprite[i] = c.mem.Read(c.regs.I + uint16(i))
	}
	collision := c.disp.DrawSprite(int(c.regs.V[x]), int(c.regs.V[y]), sprite)
	c.regs.V[0xF] = boolToByte(collision)
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
