package global

var MapWith = 40
var MapHeight = 40

func ToPosition(x, y int) int {
	return x+MapHeight*y
}
