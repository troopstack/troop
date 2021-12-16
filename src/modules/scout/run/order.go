package run

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/scout/utils"

	"github.com/axgle/mahonia"
)

type RunResult struct {
	Stdout string
	Error  string
}

// 执行命令 如interpreter='/bin/bash', arg="-c df -lh"
func OrderRunStart(name string, envs []model.Env, dir string, arg ...string) (RunResult, bool) {
	RunResult := RunResult{
		Stdout: "",
		Error:  "",
	}
	if name == "" {
		res := fmt.Sprintf("Error: parameter missing 'name' \n")
		RunResult.Error = res
		return RunResult, false
	}

	r := regexp.MustCompile(`[^\s'"]+|'([^']*)'|"([^"]*)`)
	nameSplit := r.FindAllString(name, -1)

	for i := range nameSplit {
		if strings.HasPrefix(nameSplit[i], "'") && strings.HasSuffix(nameSplit[i], "'") {
			nameSplit[i] = strings.Trim(nameSplit[i], "'")
		}
	}

	if len(nameSplit) > 1 {
		name = nameSplit[0]
		switchArgs := nameSplit[1:]
		if len(arg) == 1 && arg[0] == "" {
			arg = switchArgs
		} else {
			arg = append(switchArgs, arg...)
		}
	}

	_, err := exec.LookPath(name)
	if err != nil {
		res := fmt.Sprintf("didn't find '%s' executable\n", name)
		RunResult.Error = res
		return RunResult, false
	}
	var cmd *exec.Cmd
	if len(arg) == 1 && arg[0] == "" {
		cmd = exec.Command(name)
	} else {
		cmd = exec.Command(name, arg...)
	}

	if err := OrderSafeVerify(strings.Join(cmd.Args, " ")); err != nil {
		RunResult.Error = err.Error()
		return RunResult, false
	}

	// 环境变量
	osEnv := os.Environ()
	for e := range envs {
		if envs[e].Key != "" {
			osEnv = append(osEnv, fmt.Sprintf("%s=%s", envs[e].Key, envs[e].Value))
		}
	}
	cmd.Env = osEnv

	// 执行目录
	if dir != "" {
		cmd.Dir = dir
	}

	// 合并标准输出与错误输出
	stdout, err := cmd.CombinedOutput()

	if runtime.GOOS == "windows" {
		// 转码
		var dec mahonia.Decoder
		dec = mahonia.NewDecoder("gbk")
		res := dec.ConvertString(string(stdout))
		RunResult.Stdout = res
	} else {
		RunResult.Stdout = string(stdout)
	}
	if utils.Config().Debug.Enabled {
		log.Printf(fmt.Sprintf("Stdin: %s, Stdout: %s\n", cmd.Args, RunResult.Stdout))
	}

	if err != nil {
		res := fmt.Sprintf("Stderr: %s\n", err)
		RunResult.Error = res
		return RunResult, false
	}

	return RunResult, true
}

func OrderSafeVerify(order string) error {
	ignoreCommands := utils.GeneralIgnoreCommands
	ignoreCommands = append(ignoreCommands, utils.Config().IgnoreCommand.Commands...)
	for c := range ignoreCommands {
		cOr := strings.TrimSpace(ignoreCommands[c])
		if cOr != "" && strings.Contains(order, cOr) {
			return errors.New(fmt.Sprintf("refused to execute the [%s] command in the [%s].", cOr, order))
		}
	}
	return nil
}
