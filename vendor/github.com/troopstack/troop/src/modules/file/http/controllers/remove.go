package controllers

import (
	"net/http"
	"os"

	"github.com/troopstack/troop/src/modules/file/utils"

	"github.com/gin-gonic/gin"
)

type RemoveRequest struct {
	TaskId string `json:"task_id"`
}

func FileRemove(c *gin.Context) {
	t := RemoveRequest{}
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

	err := os.RemoveAll(utils.FileRoot + "/" + t.TaskId)
	utils.FailOnError(err, "remove task file failed")
	c.JSON(http.StatusOK, h)
	return
}
