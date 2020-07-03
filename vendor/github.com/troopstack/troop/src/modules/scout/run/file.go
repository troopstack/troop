package run

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/scout/rpc"
	"github.com/troopstack/troop/src/modules/scout/utils"
)

func DownloadFile(hostname string, ScoutTaskRequest model.ScoutFileRequest) {
	TaskResultRequest := model.TaskResultRequest{
		TaskId:    ScoutTaskRequest.TaskId,
		Scout:     hostname,
		ScoutType: "server",
		Complete:  true,
	}

	err := utils.CreateDir(ScoutTaskRequest.Dest)
	if err != nil {
		errorCipher := utils.AES_CBC_Encrypt([]byte("create save dir failed"), utils.AES)
		TaskResultRequest.Error = errorCipher
		rpc.SendResult(TaskResultRequest)
		return
	}

	req, _ := http.NewRequest("GET", ScoutTaskRequest.Url, nil)

	req.Header.Add("Http-Token", utils.Config().General.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Print(err.Error())
		errorCipher := utils.AES_CBC_Encrypt([]byte("file download failed: connection error."), utils.AES)
		TaskResultRequest.Error = errorCipher
		rpc.SendResult(TaskResultRequest)
		return
	}

	if res.StatusCode == 401 {
		errorCipher := utils.AES_CBC_Encrypt([]byte("file download failed: invalid token."), utils.AES)
		TaskResultRequest.Error = errorCipher
		rpc.SendResult(TaskResultRequest)
		return
	}

	fileP := ScoutTaskRequest.Dest + "/" + ScoutTaskRequest.FileName

	if !ScoutTaskRequest.Cover {
		if utils.IsFile(fileP) {
			errorCipher := utils.AES_CBC_Encrypt([]byte("file already exists"), utils.AES)
			TaskResultRequest.Error = errorCipher
			rpc.SendResult(TaskResultRequest)
			return
		}
	}

	f, err := os.Create(ScoutTaskRequest.Dest + "/" + ScoutTaskRequest.FileName)
	defer f.Close()
	if err != nil {
		log.Print(err.Error())
		errorCipher := utils.AES_CBC_Encrypt([]byte("file create failed"), utils.AES)
		TaskResultRequest.Error = errorCipher
		rpc.SendResult(TaskResultRequest)
		return
	}
	io.Copy(f, res.Body)
	resCipher := utils.AES_CBC_Encrypt([]byte("success"), utils.AES)
	TaskResultRequest.Result = resCipher
	rpc.SendResult(TaskResultRequest)
	return
}
