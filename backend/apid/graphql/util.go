package graphql

// clampInt returns int within given range.
func clampInt(num, min, max int) int {
	if num <= min {
		return min
	} else if num >= max {
		return max
	}
	return num
}

// maxUint32 returns larger of x or y.
func maxUint32(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

// maxUint32 returns larger of x or y.
func clampSlice(low, high, len int) (int, int) {
	low = clampInt(low, 0, len)
	high = clampInt(high, low, len)
	return low, high
}
