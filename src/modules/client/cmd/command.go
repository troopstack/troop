package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var sendTaskChan chan os.Signal
var taskId string

var envs []string
var dir string

var Command = &cobra.Command{
	Use:   "command '<target>' <name> [arguments]",
	Short: "Run Scout Command",
	Args:  cobra.MinimumNArgs(2),
	RunE:  command,
	Example: `
Linux:
troop command '*' ifconfig -o linux
troop command '*' "/bin/bash -c ifconfig" -o linux

Windows:
troop command '*' ipconfig -o windows
troop command '*' "cmd /c ipconfig" -o windows
`,
}

func init() {
	Command.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	Command.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	Command.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	Command.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	Command.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	Command.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
	Command.PersistentFlags().StringArrayVarP(&envs, "env", "e", nil,
		"set additional environment variables as key=value")
	Command.PersistentFlags().StringVarP(&dir, "dir", "", "",
		"execute the command in a directory")
}

func command(c *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("error: missing parameters")
	}

	taskUrl := utils.Config().General.Addresses + "/tasks"

	target, err := TargetParse(args[0])
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["target"] = target
	data["detach"] = true
	data["os"] = Os
	data["tag"] = Tag
	data["target_type"] = TargetType

	task := model.Task{}
	task.Name = args[1]
	if len(args) > 2 {
		task.Args = strings.Join(args[2:(len(args))], " ")
	}
	// 环境变量
	if envs != nil {
		for i := range envs {
			env := strings.SplitN(envs[i], "=", 2)
			if len(env) > 1 {
				task.Envs = append(task.Envs, model.Env{
					Key:   env[0],
					Value: env[1],
				})
			} else {
				return errors.New("environment variable is malformed, must key=value")
			}
		}
	}

	// 执行目录
	if dir != "" {
		task.Dir = dir
	}

	data["task"] = []model.Task{task}

	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", taskUrl, payload)

	req = utils.HttpHandler(req)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	go sendTaskWait()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		close(sendTaskChan)
		fmt.Println("Error: Can not connection general.")
		os.Exit(1)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	err = WaitResult(body)
	return err
}
