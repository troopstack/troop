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

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var File = &cobra.Command{
	Use:   "file.send '<target>' <source> <dest dir>",
	Short: "Push Files To Scout",
	Args:  cobra.MinimumNArgs(3),
	RunE:  file,
	Example: `
troop file.send '*' /opt/test.txt /tmp/
`,
}

var cover bool

func init() {
	File.PersistentFlags().BoolVarP(&Detach, "detach", "d", false,
		"detach executing, default false")
	File.PersistentFlags().StringVarP(&Os, "os", "o", "",
		"operating system type, linux/windows")
	File.PersistentFlags().StringVarP(&Tag, "tag", "t", "",
		"target Tag")
	File.PersistentFlags().StringVarP(&TargetType, "device", "", "",
		"target device type, example: server")
	File.PersistentFlags().StringVarP(&TargetFile, "target_file", "f", "",
		"it contains the target ini file")
	File.PersistentFlags().StringVarP(&TargetFileGroup, "target_file_group", "g", "",
		"when using ini file, can choose one or more section from ini file")
	File.PersistentFlags().BoolVarP(&cover, "cover", "c", false,
		"when this file is present target path, covering it, default false")
}

func file(c *cobra.Command, args []string) error {
	if len(args) < 3 {
		return errors.New("error: missing parameters")
	}

	filePath := args[1]
	existed := utils.IsFile(filePath)
	if !existed {
		fmt.Println("error: file", filePath, "not exists")
		os.Exit(1)
	}
	file, _ := os.Open(filePath)
	defer file.Close()

	url := utils.Config().General.Addresses + "/file"

	target, err := TargetParse(args[0])
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	fileByte, err := ioutil.ReadFile(filePath)

	data["file_name"] = filepath.Base(filePath)
	data["file"] = fileByte
	data["dest"] = args[2]
	data["cover"] = cover
	data["target"] = target
	data["detach"] = true
	data["os"] = Os
	data["tag"] = Tag
	data["target_type"] = TargetType

	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", url, payload)

	req = utils.HttpHandler(req)
	req.Header.Add("Content-Type", "multipart/form-data")

	go sendTaskWait()
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		close(sendTaskChan)
		fmt.Println("Error: Can not connection General.")
		os.Exit(1)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	err = WaitResult(body)
	return err
}
