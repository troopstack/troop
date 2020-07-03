package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var Start = &cobra.Command{
	Use:           "start",
	Short:         "Start General",
	RunE:          start,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func execGeneral() error {
	cmd := exec.Command(utils.Bin())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
	//return cmd.Start()
}

func isStarted() bool {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if utils.IsRunning() {
				return true
			}
		case <-time.After(time.Second):
			return false
		}
	}
}

func start(c *cobra.Command, args []string) error {
	if utils.IsRunning() {
		fmt.Print("[", utils.PName, "] ", utils.Pid(), "\n")
		return nil
	}
	if err := execGeneral(); err != nil {
		return err
	}
	if isStarted() {
		fmt.Print("[", utils.PName, "] ", utils.Pid(), "\n")
		return nil
	}

	return fmt.Errorf("[%s] failed to start", utils.PName)
}
