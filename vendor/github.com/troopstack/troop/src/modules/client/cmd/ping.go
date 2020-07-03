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

var Ping = &cobra.Command{
	Use:   "ping.pong '<target>'",
	Short: "Get Scout Survival Status",
	Args:  cobra.MinimumNArgs(1),
	RunE:  ping,
	Example: `
troop ping.pong '*'
`,
}

func init() {
	Ping.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	Ping.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	Ping.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	Ping.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	Ping.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	Ping.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
}

func ping(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("error: missing target parameters")
	}

	pingUrl := utils.Config().General.Addresses + "/ping"

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
