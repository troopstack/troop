package rpc

import (
	"errors"

	"github.com/troopstack/troop/src/model"
	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/utils"
)

func getMapKeys(m map[string]*model.TaskScoutInfo) []string {
	// 数组默认长度为map长度,后面append时,不需要重新申请内存和拷贝,效率较高
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (t *Scout) TaskAccept(args *model.TaskAcceptRequest, reply *model.SimpleRpcResponse) error {
	isTagTask := false
	taskKey := args.Tag
	if taskKey == "" {
		taskKey = args.Scout
	} else {
		isTagTask = true
	}
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
			return errors.New("Task has ended")
		}
	} else {
		// 任务未找到
		reply.Code = 1
		return errors.New("Task not exists")
	}
	//firstAccept := true
	scout, exists := taskCache.Tasks.GetTaskScout(args.TaskId, taskKey)
	//if !exists && isTagTask {
	//	firstAccept = false
	//	scout, exists = taskCache.Tasks.GetTaskScout(args.TaskId, args.Scout)
	//}
	if !exists && isTagTask {
		keys := getMapKeys(task.M)
		if len(keys) > 0 {
			scout, exists = taskCache.Tasks.GetTaskScout(args.TaskId, keys[0])
		}
	}
	if !exists {
		// 属于该Scout的任务未找到
		reply.Code = 1
		return errors.New("Task not exists or has been consumed")
	}
	//if firstAccept {
	if !isTagTask {
		if scout.Status != "wait" {
			// 任务已经被接收过了
			reply.Code = 1
			return errors.New("Task has been accepted")
		}
	} else {
		if scout.Status == "successful" || scout.Status == "failed" {
			// 任务已经被接收过了
			reply.Code = 1
			return errors.New("Task has been accepted")
		}
	}
	taskCache.Tasks.PutTaskScoutStatus(scout.TaskId, taskKey, "execution")

	if isTagTask {
		taskCache.Tasks.UpdateTaskScoutKey(scout.TaskId, args.Tag, args.Scout, args.ScoutType)
	}
	reply.Code = 0
	//}
	//reply.Code = 0
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
