package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func FileManage(c *gin.Context) {
	t := model.FileManageRequest{}

	h := gin.H{
		"task_id": "",
		"result":  make(map[string]*model.TaskScoutInfo),
		"error":   "",
		"code":    0,
	}

	// 校验数据
	if err := c.ShouldBindWith(&t, binding.Query); err != nil {
		h["error"] = err.Error()
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

	// 生成TaskId
	taskId := TaskGeneration()

	action := "ls"
	if c.Request.Method == "POST" {
		action = t.Action
	}

	Task := &model.FileManageTaskRequest{
		TaskId: taskId,
		Action: action,
		Prefix: t.Prefix,
	}
	data, err := json.Marshal(Task)

	if err != nil {
		log.Printf(err.Error())
		return
	}

	ScoutMessage := model.ScoutMessage{
		Type: "fileManage",
		Data: []byte(utils.AES_CBC_Encrypt(data, utils.AES)),
	}

	// 任务存储
	TaskScouts := TaskSave(taskId, t.Detach, scouts)

	// 任务推送
	TaskPush(taskId, scouts, TaskScouts, ScoutMessage)

	// 获取结果
	h = TaskResult(taskId, t.Detach, t.Timeout, h)
	c.JSON(http.StatusOK, h)
	return
}
