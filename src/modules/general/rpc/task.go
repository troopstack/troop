package rpc

import (
	"errors"

	"github.com/troopstack/troop/src/model"
	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/utils"
)

func (t *Scout) TaskAccept(args *model.TaskAcceptRequest, reply *model.SimpleRpcResponse) error {
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
	taskCache.Tasks.PutTaskScoutStatus(scout.TaskId, scout.Scout, "execution")
	reply.Code = 0
	return nil
}

func (t *Scout) TaskResult(args *model.TaskResultRequest, reply *model.SimpleRpcResponse) error {
	if args.Scout == "" {
		// 参数有误
		reply.Code = 1
		return errors.New("Parameter error")
	}
	host, exists := database.IsExistTypeScout(args.Scout, args.ScoutType)

	if !exists {
		// Scout未找到
		reply.Code = 1
		return errors.New("Scout not exists")
	}
	ScoutAES := utils.AES_CBC_Decrypt(host.AES, utils.AESKey)
	if args.Result != "" {
		result := utils.AES_CBC_Decrypt(args.Result, string(ScoutAES))
		taskCache.Tasks.PutTaskScoutResult(args.TaskId, args.Scout, string(result), false)
	}
	if args.Error != "" {
		errorMsg := utils.AES_CBC_Decrypt(args.Error, string(ScoutAES))
		taskCache.Tasks.PutTaskScoutResult(args.TaskId, args.Scout, string(errorMsg), true)
	}

	if args.Complete {
		if args.Error != "" {
			taskCache.Tasks.PutTaskScoutStatus(args.TaskId, args.Scout, "failed")
		} else {
			taskCache.Tasks.PutTaskScoutStatus(args.TaskId, args.Scout, "successful")
		}
	}
	reply.Code = 0
	return nil
}
