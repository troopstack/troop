package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var Result = &cobra.Command{
	Use:   "result <taskId>",
	Short: "Get Task Execution Results",
	RunE:  result,
}

func result(c *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("error: missing taskId parameters")
	}
	resultUrl := utils.Config().General.Addresses + "/task"
	resultData := make(map[string]interface{})
	taskId = args[0]
	resultData["task_id"] = taskId
	resultBytesData, _ := json.Marshal(resultData)
	resultReq, _ := http.NewRequest("GET", resultUrl, bytes.NewReader(resultBytesData))
	resultReq = utils.HttpHandler(resultReq)

	resultReq.Header.Add("Content-Type", "application/json")
	resultReq.Header.Add("cache-control", "no-cache")
	resultRes, resultErr := http.DefaultClient.Do(resultReq)
	if resultErr != nil {
		color.HiRed("Error: Can not connection general.")
		os.Exit(0)
	}
	defer resultRes.Body.Close()
	resultBody, _ := ioutil.ReadAll(resultRes.Body)
	resTaskResponse := model.TaskResponse{}
	resultErr = json.Unmarshal(resultBody, &resTaskResponse)
	if resultErr != nil {
		fmt.Println(resultErr.Error())
		fmt.Println(string(resultBody))
	} else {
		if resTaskResponse.Code == 0 {
			for i := range resTaskResponse.Result {
				resScoutTaskResult := resTaskResponse.Result[i]
				color.HiBlue("[" + string(resScoutTaskResult.Scout) + "]")
				if resScoutTaskResult.Status == "successful" {
					color.HiGreen(resScoutTaskResult.Result)
				} else if resScoutTaskResult.Status == "failed" {
					if resScoutTaskResult.Result != "" {
						color.HiRed(resScoutTaskResult.Result)
					}
					if resScoutTaskResult.Error != "" {
						color.HiRed("error: %s", resScoutTaskResult.Error)
					}
				} else {
					color.HiRed(resScoutTaskResult.Status)
				}
			}
		} else {
			fmt.Println("error:", resTaskResponse.Error)
		}
	}
	return nil
}
