package global

import "github.com/llr104/slgserver/config"

var MapWith = 200
var MapHeight = 200

func ToPosition(x, y int) int {
	return x + MapHeight*y
}

func IsDev() bool {
	return config.File.MustBool("slgserver", "is_dev", false)
}
