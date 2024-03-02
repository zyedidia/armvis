package main

func hilbertXY(s uint32, order int) (uint32, uint32) {
	var state, x, y uint32

	for i := 2*order - 2; i >= 0; i -= 2 {
		row := 4*state | ((s >> i) & 3)
		x = (x << 1) | ((0x936C >> row) & 1)
		y = (y << 1) | ((0x39C6 >> row) & 1)
		state = (0x3E6B94C1 >> (2 * row)) & 3
	}
	return x, y
}
