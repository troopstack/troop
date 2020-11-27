package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/troopstack/troop/src/model"
	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
)

// 随机生成TaskId
func RandomGenerateTaskId() string {
	TaskIdLen := 32
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	ByteKey := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < TaskIdLen; i++ {
		result = append(result, ByteKey[r.Intn(len(ByteKey))])
	}
	TaskId := string(result)
	return TaskId
}

// 处理RMQ不可达的任务
func TaskReturn(message model.ScoutMessage, routerKey string) {
	// AES解密
	data, err := AES_CBC_Decrypt(string(message.Data), AES)
	if err != nil {
		log.Print(err.Error())
		return
	}
	task := &model.ScoutTaskRequest{}
	err = json.Unmarshal(data, &task)
	if err != nil {
		log.Print(err.Error())
		return
	}

	scout := strings.Split(routerKey, ".")
	taskCache.Tasks.PutTaskScoutResult(task.TaskId, scout[len(scout)-1], "unreachable", true)
	taskCache.Tasks.PutTaskScoutStatus(task.TaskId, scout[len(scout)-1], "unreachable")
}

// 让FM清理缓存文件
func RemoveFMFile(taskScout taskCache.TaskScouts, taskId string) {
	fileRemoveUrl := Config().File.Address + "/file/remove"

	taskScout.Wg.Wait()
	data := make(map[string]interface{})
	data["task_id"] = taskId
	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", fileRemoveUrl, payload)
	req.Header.Add("Http-Token", Config().Http.Token)

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
}
