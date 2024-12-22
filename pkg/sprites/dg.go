package sprites

func encodeGraphicsByte(b byte) string {
	var res string
	var mask byte
	for mask = 0x80; mask != 0; mask >>= 1 {
		if b&mask != 0 {
			res += "#"
		} else {
			res += "."
		}
	}
	return res
}
