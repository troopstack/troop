package utils

import (
	"os"
)

func SetPid() {
	//output, _ := exec.Command("pgrep", "-f", PName).Output()
	//PidOf = strings.TrimSpace(string(output))
	PidOf = os.Getpid()
}

func Pid() int {
	if PidOf == 0 {
		SetPid()
	}
	return PidOf
}

func IsRunning() bool {
	SetPid()
	return Pid() != 0
}
