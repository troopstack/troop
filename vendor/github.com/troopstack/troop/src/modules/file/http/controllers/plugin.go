package controllers

import (
	"io/ioutil"
	"net/http"
	"path"

	"github.com/troopstack/troop/src/modules/file/utils"

	"github.com/gin-gonic/gin"
)

type pluginRequest struct {
	FileName        string `json:"file_name"`
	File            []byte `json:"file"`
	PluginsPathName string `json:"plugins_pathname"`
}

func PluginUpload(c *gin.Context) {
	t := pluginRequest{}
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
	pluginDir := path.Join(utils.FileRoot, "plugins")
	err := utils.CreateDir(pluginDir)
	if err != nil {
		h["error"] = "Error: Create plugins save dir failed."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	pluginTarFile := path.Join(utils.FileRoot, t.FileName)
	err = ioutil.WriteFile(pluginTarFile, t.File, 0600)

	if err != nil {
		utils.FailOnError(err, "File write failed")
		h["error"] = "Error: file write failed."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	err = utils.DeCompress(pluginTarFile, pluginDir)

	if err != nil {
		utils.FailOnError(err, "File decompress failed")
		h["error"] = "Error: file decompress failed."
		h["code"] = 1
		c.JSON(http.StatusBadRequest, h)
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "plugins/" + t.PluginsPathName})
	return
}
