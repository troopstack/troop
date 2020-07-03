package utils

import (
	"log"
	"path/filepath"
	"runtime"
)

var PName = "general"
var PidOf int

func Bin() string {
	p, _ := filepath.Abs(PName)
	return p
}

func init() {
	PidOf = 0
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
