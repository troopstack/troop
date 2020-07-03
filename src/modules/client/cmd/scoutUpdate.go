package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var ScoutUpdate = &cobra.Command{
	Use:   "scout.update '<target>' <scout tar.gz http path>",
	Short: "Scout Update",
	Args:  cobra.MinimumNArgs(2),
	RunE:  scoutUpdate,
	Example: `
troop scout.update '*' http://example.com/troop-scout.tar.gz
`,
}

func init() {
	ScoutUpdate.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	ScoutUpdate.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	ScoutUpdate.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	ScoutUpdate.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	ScoutUpdate.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	ScoutUpdate.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
}

func scoutUpdate(c *cobra.Command, args []string) error {
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
	data["plugin"] = "scout_manager"
	data["action"] = "update"
	data["args"] = args[1]

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
