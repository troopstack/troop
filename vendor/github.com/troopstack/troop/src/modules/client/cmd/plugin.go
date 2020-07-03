package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var Plugin = &cobra.Command{
	Use:   "plugin '<target>' <plugin name> <action> [arguments]",
	Short: "Call Plugin Execution",
	Args:  cobra.MinimumNArgs(3),
	RunE:  plugin,
	Example: `
troop plugin '*' plugin_name deploy
troop plugin '*' plugin_name start
`,
}

var NoCheck bool

func init() {
	Plugin.PersistentFlags().BoolVarP(&NoCheck, "no_check", "n", false,
		"Skip the check for plugin existence, default false")
	Plugin.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	Plugin.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	Plugin.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	Plugin.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	Plugin.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	Plugin.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
}

func plugin(c *cobra.Command, args []string) error {

	pingUrl := utils.Config().General.Addresses + "/plugin"

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
	data["plugin"] = args[1]
	data["action"] = args[2]
	data["args"] = strings.Join(args[3:(len(args))], " ")
	if NoCheck {
		data["no_check"] = true
	}

	if args[2] == "config.update" {
		if len(args) < 4 {
			return errors.New("error: missing config file parameters")
		}
		filePath := args[3]
		existed := utils.IsFile(filePath)
		if !existed {
			fmt.Println("error: config file", filePath, "not exists")
			os.Exit(1)
		}
		file, _ := os.Open(filePath)
		defer file.Close()

		fileByte, err := ioutil.ReadFile(filePath)

		if err != nil {
			fmt.Println("error: config file:", filePath, "read failed:", err.Error())
			os.Exit(1)
		}

		data["config_name"] = filepath.Base(filePath)
		data["config_byte"] = fileByte
	}

	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", pingUrl, payload)

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
