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

var PluginPull = &cobra.Command{
	Use:   "plugin.pull",
	Short: "Update Plugin From Git",
	RunE:  pluginPull,
	Example: `
troop plugin.pull
`,
}

func pluginPull(c *cobra.Command, args []string) error {
	url := utils.Config().General.Addresses + "/plugin/pull"

	data := make(map[string]interface{})

	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", url, payload)

	req = utils.HttpHandler(req)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		close(sendTaskChan)
		fmt.Println("Error: Can not connection general.")
		os.Exit(1)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	err = CommonResult(body)
	return err
}
