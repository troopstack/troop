package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/gin-gonic/gin"
)

type PingRequest struct {
	Target     string `json:"target" binding:"required"`
	TargetType string `json:"target_type" default:"server"`
	Tag        string `json:"tag"`
	OS         string `json:"os"`
	Detach     bool   `json:"detach" default:"false"`
	Timeout    int    `json:"timeout"`
}

func Ping(c *gin.Context) {
	t := PingRequest{}

	h := gin.H{
		"result": make(map[string]*model.TaskScoutInfo),
		"error":  "",
		"code":   0,
	}

	// 校验数据
	if err := c.ShouldBindJSON(&t); err != nil {
		h["error"] = "参数异常"
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
	Task := &model.ScoutPingRequest{
		TaskId: taskId,
	}
	data, err := json.Marshal(Task)

	if err != nil {
		log.Printf(err.Error())
		return
	}

	ScoutMessage := model.ScoutMessage{
		Type: "ping",
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
