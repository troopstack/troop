package rpc

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/scout/utils"
)

func HandleCacheTask() {
	dirs, err := ioutil.ReadDir(utils.TaskCacheDir)
	if err != nil {
		return
	}
	for _, taskCacheFile := range dirs {
		task := strings.Split(taskCacheFile.Name(), "-")
		taskId := task[0]
		if len(task) > 0 {
			taskType := task[1]
			if taskType == "update" {
				go func() {
					TaskResult(taskId, "successfully", nil)
					os.Remove(path.Join(utils.TaskCacheDir, taskCacheFile.Name()))
				}()
			}
		}
	}
}

func TaskResult(taskId, result string, err error) {
	hostname, err := utils.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("error:%s", err.Error())
		log.Print(err.Error())
		return
	}

	resCipher := utils.AES_CBC_Encrypt([]byte(result), utils.AES)

	TaskResultRequest := model.TaskResultRequest{
		TaskId:    taskId,
		Scout:     hostname,
		ScoutType: "server",
		Result:    resCipher,
		Complete:  true,
	}

	if err != nil {
		errorCipher := utils.AES_CBC_Encrypt([]byte(err.Error()), utils.AES)
		TaskResultRequest.Error = errorCipher
	}

	SendResult(TaskResultRequest)
}

func SendResult(req model.TaskResultRequest) bool {
	var resp model.SimpleRpcResponse

	return utils.CallGeneral("Scout.TaskResult", req, &resp)
}

func SendTaskAccept(req model.TaskAcceptRequest) bool {
	var resp model.SimpleRpcResponse

	return utils.CallGeneral("Scout.TaskAccept", req, &resp)
}

func SendPong(req model.TaskAcceptRequest) bool {
	var resp model.SimpleRpcResponse

	return utils.CallGeneral("Scout.Pong", req, &resp)
}

func SendScoutHavePlugins(req model.UpdateScoutHavePluginRequest) bool {
	var resp model.SimpleRpcResponse

	return utils.CallGeneral("Scout.UpdateScoutHavePlugins", req, &resp)
}
