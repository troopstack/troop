package controllers

import (
	"github.com/troopstack/troop/src/modules/file/utils"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileRequest struct {
	FileName string `json:"file_name"`
	File     []byte `json:"file"`
	TaskId   string `json:"task_id"`
}

func FileUpload(c *gin.Context) {
	t := FileRequest{}
	h := gin.H{
		"error": "",
		"code":  0,
	}

	// 校验数据
	if err := c.ShouldBindJSON(&t); err != nil {
		h["error"] = err.Error()
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}
	saveDir := utils.FileRoot + "/" + t.TaskId
	err := utils.CreateDir(saveDir)
	if err != nil {
		h["error"] = "Error: Create Save dir failed."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	saveFile := saveDir + "/" + t.FileName

	err = ioutil.WriteFile(saveFile, t.File, 0444)

	if err != nil {
		utils.FailOnError(err, "File write failed")
		h["error"] = "Error: File write failed."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": t.TaskId + "/" + t.FileName})
	return
}
