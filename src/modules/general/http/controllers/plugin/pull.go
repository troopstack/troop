package plugin

import (
	"github.com/gin-gonic/gin"
	"github.com/troopstack/troop/src/modules/general/utils"
	"net/http"
)

func PluginPull(c *gin.Context) {
	// 触发从git拉取插件
	h := gin.H{
		"result": "",
		"error":  "",
		"code":   0,
	}

	utils.PluginCh = make(chan int, 1)
	go utils.InitPlugins()
	initPluginsResult := <-utils.PluginCh
	if initPluginsResult == 0 {
		h["error"] = "error pulling plugin from git, please check general log"
		h["code"] = 1
		c.JSON(http.StatusInternalServerError, h)
	} else {
		h["result"] = "successfully"
		c.JSON(http.StatusOK, h)
	}
	return
}
