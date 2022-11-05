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

var PluginVersion = &cobra.Command{
	Use:   "plugin.version '<target>' '<plugin name>'",
	Short: "Get Plugin Version",
	Args:  cobra.MinimumNArgs(2),
	RunE:  pluginVersion,
	Example: `
troop plugin.version '*' plugin_name
`,
}

func init() {
	PluginVersion.PersistentFlags().BoolVarP(&NoCheck, "no_check", "n", false,
		"Skip the check for plugin existence, default false")
	PluginVersion.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	PluginVersion.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	PluginVersion.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	PluginVersion.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	PluginVersion.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	PluginVersion.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
}

func pluginVersion(c *cobra.Command, args []string) error {
	if len(args) < 2 {
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
	data["action"] = "version"
	data["plugin"] = args[1]
	if NoCheck {
		data["no_check"] = true
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
