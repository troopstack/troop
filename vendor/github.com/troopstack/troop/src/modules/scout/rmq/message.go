package rmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/scout/rpc"
	"github.com/troopstack/troop/src/modules/scout/run"
	"github.com/troopstack/troop/src/modules/scout/utils"

	"github.com/streadway/amqp"
)

func MessageProcess(d amqp.Delivery, autoAck bool) {
	body := d.Body
	ScoutMessage := model.ScoutMessage{}
	err := json.Unmarshal(body, &ScoutMessage)
	if !autoAck {
		defer func() {
			err = d.Ack(false)
			failOnError(err, "Rabbitmq ack failed")
		}()
	}
	if err != nil {
		log.Print(err.Error())
		return
	}
	if ScoutMessage.Type == "accept" {
		ScoutPriKey, is_have := utils.ReadScoutPriKey()
		if !is_have {
			log.Print("Scout Private read failed")
			return
		}
		err = utils.SaveGeneralInfo(ScoutMessage.Data, ScoutPriKey)
		if err != nil {
			log.Fatalf("Save general info failed：%s", err)
			return
		}
		log.Print("Scout have been accepted")
	} else if ScoutMessage.Type == "handshake" {
		utils.CallHandshake()
	} else if ScoutMessage.Type == "task" {
		if utils.GeneralAES == "" {
			return
		}
		// 使用General的AES解密
		data := utils.AES_CBC_Decrypt(string(ScoutMessage.Data), utils.GeneralAES)

		ScoutTaskRequest := model.ScoutTaskRequest{}
		err := json.Unmarshal(data, &ScoutTaskRequest)
		if err != nil {
			log.Print(err.Error())
			return
		}
		hostname, err := utils.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
			log.Print(err.Error())
			return
		}

		// 接收消息
		TaskAcceptRequest := model.TaskAcceptRequest{
			TaskId:    ScoutTaskRequest.TaskId,
			Scout:     hostname,
			ScoutType: "server",
		}
		acceptAck := rpc.SendTaskAccept(TaskAcceptRequest)
		if !acceptAck {
			log.Println("send task receive message error: ", TaskAcceptRequest)
			return
		}
		log.Printf("receive job. TaskId: %s , Task: %s", ScoutTaskRequest.TaskId, ScoutTaskRequest.Task)

		for i := range ScoutTaskRequest.Task {
			var result run.RunResult
			var runNormal bool
			result, runNormal = run.OrderRunStart(
				ScoutTaskRequest.Task[i].Name,
				ScoutTaskRequest.Task[i].Envs,
				ScoutTaskRequest.Task[i].Dir,
				ScoutTaskRequest.Task[i].Args)

			resCipher := utils.AES_CBC_Encrypt([]byte(result.Stdout), utils.AES)

			TaskResultRequest := model.TaskResultRequest{
				TaskId:    ScoutTaskRequest.TaskId,
				Scout:     hostname,
				ScoutType: "server",
				Result:    resCipher,
				Complete:  len(ScoutTaskRequest.Task)-1 == i,
			}

			if result.Error != "" {
				errorCipher := utils.AES_CBC_Encrypt([]byte(result.Error), utils.AES)
				TaskResultRequest.Error = errorCipher
			}

			if !runNormal {
				log.Printf("Exec order 「%s」 failed : %s",
					(ScoutTaskRequest.Task[i].Name + " " + ScoutTaskRequest.Task[i].Args), result.Error)
			}
			rpc.SendResult(TaskResultRequest)
		}
	} else if ScoutMessage.Type == "ping" {
		if utils.GeneralAES == "" {
			return
		}
		// 使用General的AES解密
		data := utils.AES_CBC_Decrypt(string(ScoutMessage.Data), utils.GeneralAES)

		ScoutTaskRequest := model.ScoutPingRequest{}
		err := json.Unmarshal(data, &ScoutTaskRequest)
		if err != nil {
			log.Print(err.Error())
			return
		}
		hostname, err := utils.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
			log.Print(err.Error())
			return
		}

		// 接收消息
		TaskAcceptRequest := model.TaskAcceptRequest{
			TaskId:    ScoutTaskRequest.TaskId,
			Scout:     hostname,
			ScoutType: "server",
		}
		acceptAck := rpc.SendPong(TaskAcceptRequest)
		if !acceptAck {
			log.Println("send task receive message error:", TaskAcceptRequest)
			return
		}
		log.Printf("receive ping job. TaskId: %s", ScoutTaskRequest.TaskId)
	} else if ScoutMessage.Type == "file_dist" {
		if utils.GeneralAES == "" {
			return
		}
		// 使用General的AES解密
		data := utils.AES_CBC_Decrypt(string(ScoutMessage.Data), utils.GeneralAES)

		ScoutTaskRequest := model.ScoutFileRequest{}
		err := json.Unmarshal(data, &ScoutTaskRequest)
		if err != nil {
			log.Print(err.Error())
			return
		}
		hostname, err := utils.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
			log.Print(err.Error())
			return
		}
		// 接收消息
		TaskAcceptRequest := model.TaskAcceptRequest{
			TaskId:    ScoutTaskRequest.TaskId,
			Scout:     hostname,
			ScoutType: "server",
		}
		acceptAck := rpc.SendTaskAccept(TaskAcceptRequest)
		if !acceptAck {
			log.Println("send file distribution job receive message error: ", TaskAcceptRequest)
			return
		}
		log.Printf("receive file distribution job. TaskId: %s , file: %s", ScoutTaskRequest.TaskId,
			ScoutTaskRequest.Dest+"/"+ScoutTaskRequest.FileName)

		run.DownloadFile(hostname, ScoutTaskRequest)
		return
	} else if ScoutMessage.Type == "plugin" {
		if utils.GeneralAES == "" {
			return
		}
		// 使用General的AES解密
		data := utils.AES_CBC_Decrypt(string(ScoutMessage.Data), utils.GeneralAES)

		ScoutPluginRequest := model.ScoutPluginRequest{}
		err := json.Unmarshal(data, &ScoutPluginRequest)
		if err != nil {
			log.Print(err.Error())
			return
		}
		if ScoutPluginRequest.TaskId != "" {
			hostname, err := utils.Hostname()
			if err != nil {
				hostname = fmt.Sprintf("error:%s", err.Error())
				log.Print(err.Error())
				return
			}
			// 接收消息
			TaskAcceptRequest := model.TaskAcceptRequest{
				TaskId:    ScoutPluginRequest.TaskId,
				Scout:     hostname,
				ScoutType: "server",
			}
			acceptAck := rpc.SendTaskAccept(TaskAcceptRequest)
			if !acceptAck {
				log.Println("send plugin job receive message error: ", TaskAcceptRequest)
				return
			}
			log.Printf("receive plugin job. TaskId: %s, Action: %s", ScoutPluginRequest.TaskId,
				ScoutPluginRequest.Action)
			var result string

			result, err = run.PluginAction(ScoutPluginRequest)

			resCipher := utils.AES_CBC_Encrypt([]byte(result), utils.AES)

			TaskResultRequest := model.TaskResultRequest{
				TaskId:    ScoutPluginRequest.TaskId,
				Scout:     hostname,
				ScoutType: "server",
				Result:    resCipher,
				Complete:  true,
			}

			if err != nil {
				errorCipher := utils.AES_CBC_Encrypt([]byte(err.Error()), utils.AES)
				TaskResultRequest.Error = errorCipher
			}

			rpc.SendResult(TaskResultRequest)

		} else {
			run.PluginAction(ScoutPluginRequest)
		}
	} else if ScoutMessage.Type == "fileManage" {
		if utils.GeneralAES == "" {
			return
		}
		// 使用General的AES解密
		data := utils.AES_CBC_Decrypt(string(ScoutMessage.Data), utils.GeneralAES)

		ScoutTaskRequest := model.FileManageTaskRequest{}
		err := json.Unmarshal(data, &ScoutTaskRequest)
		if err != nil {
			log.Print(err.Error())
			return
		}
		hostname, err := utils.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
			log.Print(err.Error())
			return
		}

		// 接收消息
		TaskAcceptRequest := model.TaskAcceptRequest{
			TaskId:    ScoutTaskRequest.TaskId,
			Scout:     hostname,
			ScoutType: "server",
		}
		acceptAck := rpc.SendTaskAccept(TaskAcceptRequest)
		if !acceptAck {
			log.Println("send task receive message error: ", TaskAcceptRequest)
			return
		}
		log.Printf("receive file manage job. TaskId: %s , action: %s, prefix: %s",
			ScoutTaskRequest.TaskId, ScoutTaskRequest.Action, ScoutTaskRequest.Prefix)
		if ScoutTaskRequest.Action == "ls" {
			fileList, err := run.FileList(ScoutTaskRequest.Prefix)

			TaskResultRequest := model.TaskResultRequest{
				TaskId:    ScoutTaskRequest.TaskId,
				Scout:     hostname,
				ScoutType: "server",
				Result:    "",
				Complete:  true,
			}

			if err != nil {
				errorCipher := utils.AES_CBC_Encrypt([]byte(err.Error()), utils.AES)
				TaskResultRequest.Error = errorCipher
			} else {
				resultData, err := json.Marshal(fileList)
				if err != nil {
					errorCipher := utils.AES_CBC_Encrypt([]byte(err.Error()), utils.AES)
					TaskResultRequest.Error = errorCipher
				}
				TaskResultRequest.Result = utils.AES_CBC_Encrypt(resultData, utils.AES)
			}
			rpc.SendResult(TaskResultRequest)
		}
	} else if ScoutMessage.Type == "bala_task" {
		if utils.GeneralAES == "" {
			return
		}
		// 使用General的AES解密
		data := utils.AES_CBC_Decrypt(string(ScoutMessage.Data), utils.GeneralAES)

		ScoutPluginRequest := model.ScoutPluginRequest{}
		err := json.Unmarshal(data, &ScoutPluginRequest)
		if err != nil {
			log.Print(err.Error())
			return
		}
		if ScoutPluginRequest.TaskId != "" {
			hostname, err := utils.Hostname()
			if err != nil {
				hostname = fmt.Sprintf("error:%s", err.Error())
				log.Print(err.Error())
				return
			}
			// 接收消息
			TaskAcceptRequest := model.TaskAcceptRequest{
				TaskId:    ScoutPluginRequest.TaskId,
				Scout:     hostname,
				ScoutType: "server",
				Tag:       ScoutMessage.Tag,
			}
			acceptAck := rpc.SendTaskAccept(TaskAcceptRequest)
			if !acceptAck {
				log.Println("send plugin job receive message error: ", TaskAcceptRequest)
				return
			}
			log.Printf("receive plugin job. TaskId: %s, Action: %s", ScoutPluginRequest.TaskId,
				ScoutPluginRequest.Action)
			var result string

			result, err = run.PluginAction(ScoutPluginRequest)

			resCipher := utils.AES_CBC_Encrypt([]byte(result), utils.AES)

			TaskResultRequest := model.TaskResultRequest{
				TaskId:    ScoutPluginRequest.TaskId,
				Scout:     hostname,
				ScoutType: "server",
				Result:    resCipher,
				Complete:  true,
			}

			if err != nil {
				errorCipher := utils.AES_CBC_Encrypt([]byte(err.Error()), utils.AES)
				TaskResultRequest.Error = errorCipher
			}

			rpc.SendResult(TaskResultRequest)

		} else {
			run.PluginAction(ScoutPluginRequest)
		}
	} else {
		log.Print(ScoutMessage)
	}
}
