package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/http/controllers"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/gin-gonic/gin"
)

type fileUploadResponse struct {
	Url string
}

func PluginJob(c *gin.Context) {
	// 执行插件任务
	t := model.PluginRequest{}

	h := gin.H{
		"task_id": "",
		"result":  make(map[string]*model.TaskScoutInfo),
		"error":   "",
		"code":    0,
	}

	if !utils.Config().Plugin.Enabled {
		h["error"] = "plugin not enabled"
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	// 校验数据
	if err := c.ShouldBindJSON(&t); err != nil {
		h["error"] = err.Error()
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	// 目标校验
	scouts, err := controllers.TaskTargetVerify(t.Target, t.TargetType, t.OS, t.Tag)
	if err != nil {
		h["error"] = err.Error()
		h["code"] = 1
		c.JSON(http.StatusInternalServerError, h)
		return
	}

	// 生成TaskId
	taskId := controllers.TaskGeneration()

	Task := &model.ScoutPluginRequest{
		TaskId: taskId,
		Plugin: t.Plugin,
		Action: t.Action,
		Args:   t.Args,
	}

	pluginName := t.Plugin

	check := !t.NoCheck

	if t.Action == "update_plugins" {
		utils.PluginCh = make(chan int, 1)
		go utils.InitPlugins()
		initPluginsResult := <-utils.PluginCh
		if initPluginsResult == 0 {
			h["error"] = "error pulling plugin from git, please check general log"
			h["code"] = 1
			c.JSON(http.StatusInternalServerError, h)
			return
		}
		Task.Plugins = utils.Plugins
		pluginName = strings.Split(pluginName, ":")[0]
		if check && pluginName == "" {
			check = false
		}
	}

	if check {
		if _, pluginExists := utils.Plugins[pluginName]; !pluginExists {
			h["error"] = fmt.Sprintf("plugin %s not exists", pluginName)
			h["code"] = 1
			c.JSON(http.StatusInternalServerError, h)
			return
		}
	}

	if t.Action == "config.update" {
		// 将配置文件推送到文件系统
		fileServerUrl := utils.Config().File.Address + "/file/upload"
		fileDownloadBaseUrl := utils.Config().File.Address + "/file/download/"

		data := make(map[string]interface{})
		data["file"] = t.ConfigByte
		data["file_name"] = t.ConfigName
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
			h["error"] = "General connection file server failed. " + string(body)
			h["code"] = 1
			c.JSON(http.StatusBadRequest, h)
			return
		}

		if fileUploadResponse.Url == "" {
			h["error"] = "General sent file to file server failed. " + string(body)
			h["code"] = 1
			c.JSON(http.StatusBadRequest, h)
			return
		}

		Task.FileUrl = fileDownloadBaseUrl + fileUploadResponse.Url
		log.Print(Task.FileUrl)
	}

	data, err := json.Marshal(Task)

	if err != nil {
		log.Printf(err.Error())
		return
	}

	ScoutMessage := model.ScoutMessage{
		Type: "plugin",
		Data: []byte(utils.AES_CBC_Encrypt(data, utils.AES)),
	}

	// 任务存储
	TaskScouts := controllers.TaskSave(taskId, t.Detach, scouts)

	if t.Action == "config.update" {
		// 协程等待文件下发完毕之后让FM清理缓存文件
		go utils.RemoveFMFile(TaskScouts, taskId)
	}
	// 任务推送
	controllers.TaskPush(taskId, scouts, TaskScouts, ScoutMessage)

	// 获取结果
	h = controllers.TaskResult(taskId, t.Detach, t.Timeout, h)
	c.JSON(http.StatusOK, h)
	return
}
