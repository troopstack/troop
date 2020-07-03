package rpc

import (
	"errors"

	"github.com/troopstack/troop/src/model"
	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
)

func (t *Scout) Pong(args *model.TaskAcceptRequest, reply *model.SimpleRpcResponse) error {
	if args.Scout == "" || args.TaskId == "" {
		// 参数有误
		reply.Code = 1
		return errors.New("Parameter error")
	}
	task, exists := taskCache.Tasks.GetTask(args.TaskId)
	if exists {
		if task.Lock {
			// 任务已经被锁定
			reply.Code = 1
			return errors.New("task has ended")
		}
	} else {
		// 任务未找到
		reply.Code = 1
		return errors.New("Task not exists")
	}

	scout, exists := taskCache.Tasks.GetTaskScout(args.TaskId, args.Scout)
	if !exists {
		// Scout未找到
		reply.Code = 1
		return errors.New("Scout not exists")
	}
	if scout.Status != "wait" {
		// 任务已经被接收过了
		reply.Code = 1
		return errors.New("task has been accepted")
	}
	taskCache.Tasks.PutTaskScoutResult(scout.TaskId, scout.Scout, "True", false)
	taskCache.Tasks.PutTaskScoutStatus(scout.TaskId, scout.Scout, "successful")
	reply.Code = 0
	return nil
}
