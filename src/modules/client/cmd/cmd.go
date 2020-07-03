package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/fatih/color"
	tFile "github.com/toolkits/file"
)

var Detach bool
var Os string
var Tag string
var TargetType string
var TargetFile string
var TargetFileGroup string

func sendTaskWait() {
	sendTaskChan = make(chan os.Signal)
	signal.Notify(sendTaskChan)

	s := <-sendTaskChan // 阻塞直至有信号传入
	if s == os.Interrupt {
		fmt.Println("Exiting on CTRL+C")
		if taskId != "" {
			fmt.Println("Task id is", taskId)
			fmt.Println("The scouts may not have all finished running and any remaining scouts will return upon " +
				"completion. To look up the return data for this task later run")
			fmt.Println("troop result", taskId)
		}
		os.Exit(0)
	}
}

func WaitResult(body []byte) error {
	TaskResponse := model.TaskResponse{}
	err := json.Unmarshal(body, &TaskResponse)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(body))
	} else {
		if TaskResponse.Code == 0 {
			taskId = TaskResponse.TaskId
			if Detach {
				fmt.Println("Task has been pushed to the queue. Task id is", taskId)
				fmt.Println("To look up the return data for this task later run")
				fmt.Println("troop result", taskId)
			} else {
				resultData := make(map[string]interface{})
				resultData["task_id"] = taskId
				resultData["wait"] = true
				resultBytesData, _ := json.Marshal(resultData)
				var ResultUrl = utils.Config().General.Addresses + "/task"
				resultReq, _ := http.NewRequest("GET", ResultUrl, bytes.NewReader(resultBytesData))

				resultReq = utils.HttpHandler(resultReq)
				resultReq.Header.Add("Content-Type", "application/json")
				resultReq.Header.Add("cache-control", "no-cache")
				resultRes, resultErr := http.DefaultClient.Do(resultReq)
				if resultErr != nil {
					color.HiRed("Error: Can not connection general.")
					os.Exit(1)
				}
				defer resultRes.Body.Close()
				resultBody, _ := ioutil.ReadAll(resultRes.Body)
				resTaskResponse := model.TaskResponse{}
				resultErr = json.Unmarshal(resultBody, &resTaskResponse)
				if resultErr != nil {
					color.HiRed(resultErr.Error())
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
			}
		} else {
			fmt.Println("error: ", TaskResponse.Error)
		}
	}
	close(sendTaskChan)
	return nil
}

func CommonResult(body []byte) error {
	response := model.CommonResponse{}
	err := json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(body))
	} else {
		if response.Code != 0 {
			if response.Result != "" {
				color.HiRed(response.Result)
			}
			if response.Error != "" {
				color.HiRed("error: %s", response.Error)
			}
		} else {
			color.HiGreen(response.Result)
		}
	}
	return nil
}

func TargetParse(target string) (string, error) {
	if TargetFile != "" {
		if !tFile.IsExist(TargetFile) {
			return "", errors.New(fmt.Sprintln("file:", TargetFile, "is not exist."))
		}

		if err := utils.TargetIniLoad(TargetFile); err != nil {
			return "", errors.New(fmt.Sprintln("file:", TargetFile, ", failed:", err.Error()))
		}

		if TargetFileGroup != "" {
			return strings.Join(utils.TargetIniGetKeys(TargetFileGroup), ","), nil
		} else {
			return strings.Join(utils.TargetIniGetAllKeys(), ","), nil
		}
	}
	return target, nil
}
