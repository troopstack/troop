package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var PluginUpdate = &cobra.Command{
	Use:   "plugin.update '<target>'",
	Short: "Tell The Target Scout To Update The Plugin",
	Args:  cobra.MinimumNArgs(1),
	RunE:  pluginUpdate,
	Example: `
troop plugin.update '*'
`,
}

var Cover bool
var pluginName string

func init() {
	PluginUpdate.PersistentFlags().StringVarP(&pluginName, "plugin", "p", "",
		"specify plugin")
	PluginUpdate.PersistentFlags().BoolVarP(&Cover, "cover", "c", false,
		"force updates even if plugin versions are the same, default false")
	PluginUpdate.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	PluginUpdate.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	PluginUpdate.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	PluginUpdate.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	PluginUpdate.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	PluginUpdate.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
}

func pluginUpdate(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("error: missing target parameters")
	}

	pingUrl := utils.Config().General.Addresses + "/plugin/job"

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
	data["action"] = "update_plugins"
	if Cover {
		data["args"] = "cover"
	}
	if pluginName != "" {
		data["plugin"] = pluginName
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
