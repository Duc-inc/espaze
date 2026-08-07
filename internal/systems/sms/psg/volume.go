package psg

// volumeTable converts a 4-bit attenuation register (0 = loudest, 15 =
// silent) into a linear amplitude - real hardware attenuates in ~2dB
// steps, which this table approximates.
var volumeTable = [16]int16{
	8191, 6506, 5168, 4105, 3261, 2590, 2057, 1634,
	1298, 1031, 819, 650, 516, 410, 325, 0,
}

func attenToVolume(atten byte) int16 { return volumeTable[atten&0x0F] }
