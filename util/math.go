package util

func MinInt(x, y int) int{
	if x>y {
		return y
	}
	return x
}

func MaxInt(x, y int) int{
	if x<y {
		return y
	}
	return x
}
