package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/gin-gonic/gin"
)

type fileUploadResponse struct {
	Url string
}

func FileSend(c *gin.Context) {
	t := model.FileRequest{}

	h := gin.H{
		"task_id": "",
		"result":  make(map[string]*model.TaskScoutInfo),
		"error":   "",
		"code":    0,
	}

	// 校验数据
	if err := c.ShouldBindJSON(&t); err != nil {
		h["error"] = err.Error()
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	err := ioutil.WriteFile(t.FileName, t.File, os.ModeAppend)
	if err != nil {
		h["error"] = "Error: File write failed."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	// 目标校验
	scouts, err := TaskTargetVerify(t.Target, t.TargetType, t.OS, t.Tag)
	if err != nil {
		h["error"] = err.Error()
		h["code"] = 1
		c.JSON(http.StatusInternalServerError, h)
		return
	}

	fileServerUrl := utils.Config().File.Address + "/file/upload"
	fileDownloadBaseUrl := utils.Config().File.Address + "/file/download/"

	// 生成TaskId
	taskId := TaskGeneration()

	data := make(map[string]interface{})
	data["file"] = t.File
	data["file_name"] = t.FileName
	data["task_id"] = taskId

	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", fileServerUrl, payload)

	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Add("Http-Token", utils.Config().Http.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		h["error"] = "Error: General can not connection file server."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode == 401 {
		h["error"] = "General connection file server failed. " + string(body)
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	fileUploadResponse := fileUploadResponse{}

	err = json.Unmarshal(body, &fileUploadResponse)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(body))
	} else {
		// 任务存储
		TaskScouts := TaskSave(taskId, t.Detach, scouts)

		// 协程等待文件下发完毕之后让FM清理缓存文件
		go utils.RemoveFMFile(TaskScouts, taskId)

		Task := &model.ScoutFileRequest{
			TaskId:   taskId,
			Url:      fileDownloadBaseUrl + fileUploadResponse.Url,
			FileName: t.FileName,
			Dest:     t.Dest,
			Cover:    t.Cover,
		}
		data, _ := json.Marshal(Task)

		ScoutMessage := model.ScoutMessage{
			Type: "file_dist",
			Data: []byte(utils.AES_CBC_Encrypt(data, utils.AES)),
		}

		// 任务推送
		TaskPush(taskId, scouts, TaskScouts, ScoutMessage)

		// 获取结果
		h = TaskResult(taskId, t.Detach, t.Timeout, h)
		c.JSON(http.StatusOK, h)
	}
	return
}
