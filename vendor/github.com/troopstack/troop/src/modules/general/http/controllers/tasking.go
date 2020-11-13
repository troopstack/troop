package controllers

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/troopstack/troop/src/model"
	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/rmq"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/chenhg5/collection"
	"github.com/gin-gonic/gin"
)

var TaskGenLock sync.Mutex

func TaskTargetVerify(target, targetType, os, tag string) ([]*model.Host, error) {

	matchedArg := model.MatchedScout{
		Target:     target,
		TargetType: targetType,
		OS:         os,
		Tag:        tag,
	}
	scouts, err := database.MatchedScout(&matchedArg)
	if err != nil {
		return scouts, err
	}

	if len(scouts) == 0 {
		return scouts, errors.New("no scouts matched the target")
	}

	if matchedArg.Target != "*" {
		scoutMap := strings.Replace(matchedArg.Target, " ", "", -1)
		targetSplit := strings.Split(scoutMap, ",")

		scoutsSplit := []string{}
		for i := range scouts {
			scoutsSplit = append(scoutsSplit, scouts[i].Hostname)
		}

		for i := range targetSplit {
			if !collection.Collect(scoutsSplit).Contains(targetSplit[i]) {
				return scouts, errors.New(fmt.Sprintf("scout '%s' not exist", targetSplit[i]))
			}
		}
	}

	if rmq.AvailableConnNum() <= 0 {
		return scouts, errors.New("troop server abnormal. General failed to connect with RabbitMQ")
	}

	return scouts, nil
}

func TaskGeneration() (taskId string) {
	TaskGenLock.Lock()
	defer TaskGenLock.Unlock()
	for {
		taskId = utils.RandomGenerateTaskId()
		if _, exists := taskCache.Tasks.GetTask(taskId); !exists {
			break
		}
	}
	return
}

func TaskSave(taskId string, detach bool, scouts []*model.Host) (TaskScouts taskCache.TaskScouts) {
	TaskScouts = taskCache.TaskScouts{
		TaskId:        taskId,
		AcceptCount:   0,
		CompleteCount: 0,
		Lock:          false,
		Detach:        detach,
		M:             make(map[string]*model.TaskScoutInfo),
		Wg:            sync.WaitGroup{},
		CreateAt:      time.Now(),
	}
	taskCache.Tasks.PutTask(&TaskScouts)
	TaskScouts.Wg.Add(len(scouts))
	if !detach {
		TaskScouts.Ch = make(chan int, 1)
	}
	return
}

func TaskPush(taskId string, scouts []*model.Host, TaskScouts taskCache.TaskScouts, ScoutMessage model.ScoutMessage) {

	for i := range scouts {
		TaskScoutInfo := model.TaskScoutInfo{
			TaskId:    taskId,
			Scout:     scouts[i].Hostname,
			ScoutType: scouts[i].Type,
			Result:    "",
			Error:     "",
			Status:    "wait",
		}
		taskCache.Tasks.CreateTaskScout(&TaskScoutInfo)

		target := fmt.Sprintf("scout.%s.%s", scouts[i].Type, scouts[i].Hostname)

		// 推送任务到目标scout
		_, err := rmq.AmqpServer.PutIntoQueue("scout", target, ScoutMessage, 0)

		if err != nil {
			TaskScoutInfo.Error = "Task reception failed. Please check the connection status between general and rabbitMQ."
			TaskScoutInfo.Status = "failed"
			TaskScouts.TaskDone()
		}
	}
}

func BalaTaskPush(taskId string, tag string, TaskScouts taskCache.TaskScouts, ScoutMessage model.ScoutMessage, Priority uint8) {

	TaskScoutInfo := model.TaskScoutInfo{
		TaskId:    taskId,
		Tag:       tag,
		Result:    "",
		Error:     "",
		Status:    "wait",
	}

	taskCache.Tasks.CreateBalaTaskScout(&TaskScoutInfo)

	target := fmt.Sprintf("scout.tag.%s", tag)

	// 推送任务到目标scout
	_, err := rmq.AmqpServer.PutIntoQueue("scout", target, ScoutMessage, Priority)

	if err != nil {
		TaskScoutInfo.Error = "Task reception failed. Please check the connection status between general and rabbitMQ."
		TaskScoutInfo.Status = "failed"
		TaskScouts.TaskDone()
	}
}

func TaskResult(taskId string, detach bool, timeout int, h gin.H) gin.H {
	if !detach {
		task, exists := taskCache.Tasks.GetTask(taskId)
		if !exists {
			h["error"] = "Task not exists"
			h["code"] = 1
			return h
		}

		go task.TaskWait()
		if timeout <= 0 {
			// 阻塞等待所有任务完成
		loop1:
			for {
				select {
				case <-task.Ch:
					h["task_id"] = taskId
					h["result"] = task.M
					break loop1
				}
			}
		} else {

		loop:
			for {
				select {
				case <-task.Ch:
					h["task_id"] = taskId
					h["result"] = task.M
					break loop

				case <-time.After(time.Second * time.Duration(timeout)):
					// 锁定任务
					taskCache.Tasks.PutTaskLock(taskId, true)
					// 设置未接收的任务状态为超时
					taskCache.Tasks.PutTaskTimeout(taskId)
					if task.CompleteCount-len(task.M) < 0 {
						// 结束剩余的等待组
						go task.Wg.Add(task.CompleteCount - len(task.M))
					}
					// 返回结果
					h["task_id"] = taskId
					h["result"] = task.M
					break loop
				}
			}
		}

	} else {
		h["task_id"] = taskId
	}
	return h
}
