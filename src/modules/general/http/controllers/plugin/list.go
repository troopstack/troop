package plugin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/troopstack/troop/src/modules/general/utils"
)

func PluginList(c *gin.Context) {
	// 获取插件列表
	h := gin.H{
		"result": utils.Plugins,
		"error":  "",
		"code":   0,
	}

	c.JSON(http.StatusOK, h)
	return
}
