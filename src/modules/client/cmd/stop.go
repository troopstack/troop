package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var Stop = &cobra.Command{
	Use:   "stop",
	Short: "Stop General",
	RunE:  stop,
}

func stop(c *cobra.Command, args []string) error {
	if !utils.IsRunning() {
		fmt.Print("[", utils.PName, "] down\n")
		return nil
	}
	sysType := runtime.GOOS
	var cmd *exec.Cmd
	var commArg string
	if sysType == "windows" {
		commArg = fmt.Sprintf("/c taskkill /F /pid %s", utils.Pid())
		cmd = exec.Command("cmd", commArg)
	} else {
		commArg = fmt.Sprintf("-TERM %s", utils.Pid())
		cmd = exec.Command("kill", commArg)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		fmt.Print("[", utils.PName, "] down\n")
		return nil
	}
	return err
}
