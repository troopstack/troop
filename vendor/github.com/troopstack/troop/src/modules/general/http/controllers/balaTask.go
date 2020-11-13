package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/troopstack/troop/src/model"
	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/gin-gonic/gin"
)

func BalaTasks(c *gin.Context) {
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
	scouts, err := TaskTargetVerify(t.Target, t.TargetType, t.OS, t.Tag)
	if err != nil {
		h["error"] = err.Error()
		h["code"] = 1
		c.JSON(http.StatusInternalServerError, h)
		return
	}

	// 生成TaskId
	taskId := TaskGeneration()

	Task := &model.ScoutPluginRequest{
		TaskId: taskId,
		Plugin: t.Plugin,
		Action: t.Action,
		Args:   t.Args,
	}

	pluginName := t.Plugin

	check := !t.NoCheck

	if check {
		if _, pluginExists := utils.Plugins[pluginName]; !pluginExists {
			h["error"] = fmt.Sprintf("plugin %s not exists", pluginName)
			h["code"] = 1
			c.JSON(http.StatusInternalServerError, h)
			return
		}
	}

	data, err := json.Marshal(Task)

	if err != nil {
		log.Printf(err.Error())
		return
	}

	ScoutMessage := model.ScoutMessage{
		Type: "bala_task",
		Tag: t.Tag,
		Data: []byte(utils.AES_CBC_Encrypt(data, utils.AES)),
	}
	oScouts := []*model.Host{scouts[0]}
	// 任务存储
	TaskScouts := TaskSave(taskId, t.Detach, oScouts)

	priority := uint8(t.Priority)
	// 任务推送
	BalaTaskPush(taskId, t.Tag, TaskScouts, ScoutMessage, priority)

	// 获取结果
	h = TaskResult(taskId, t.Detach, t.Timeout, h)
	c.JSON(http.StatusOK, h)
	return
}

func BalaTaskInfo(c *gin.Context) {
	t := model.TaskInfoRequest{}

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

	taskId := t.TaskId
	h["task_id"] = taskId

	task, exists := taskCache.Tasks.GetTask(taskId)
	if !exists {
		h["error"] = "Task not exists"
		h["code"] = 1
		c.JSON(http.StatusOK, h)
	} else {
		if t.Wait {
			ch := make(chan int, 1)
			go func() {
				task.Wg.Wait()
				ch <- 1
			}()
			// 阻塞等待所有任务完成
		loop:
			for {
				select {
				case <-ch:
					h["result"] = task.M
					c.JSON(http.StatusOK, h)
					break loop
				}
			}
		} else {
			h["result"] = task.M
			c.JSON(http.StatusOK, h)
		}
	}

	return
}
