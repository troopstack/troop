package http

import (
	"net/http"
	"strings"

	"github.com/troopstack/troop/src/modules/general/http/controllers"
	"github.com/troopstack/troop/src/modules/general/http/controllers/plugin"
	"github.com/troopstack/troop/src/modules/general/utils"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
)

func InitRouter() http.Handler {
	router := gin.New()
	router.Use(gin.Logger())

	if !utils.Config().Debug.Enabled {
		gin.SetMode(gin.ReleaseMode)
	} else {
		ginpprof.Wrap(router)
	}

	// 校验Token中间件
	router.Use(tokenVerify())

	router.POST("/bala_tasks", controllers.BalaTasks)
	router.POST("/ping", controllers.Ping)
	router.POST("/file", controllers.FileSend)
	router.POST("/tasks", controllers.Tasks)

	router.GET("/plugin", plugin.PluginList)
	router.POST("/plugin/job", plugin.PluginJob)
	router.POST("/plugin/pull", plugin.PluginPull)

	router.GET("/task", controllers.TaskInfo)
	router.GET("/hosts", controllers.HostList)
	router.GET("/hosts/all", controllers.AllHostList)
	router.GET("/host/detail/:hostname", controllers.HostDetail)
	router.GET("/host/keys", controllers.HostKeyList)

	router.POST("/host/accept", controllers.AcceptHost)
	router.POST("/host/accept/all", controllers.AcceptAllHost)
	router.POST("/host/reject", controllers.RejectHost)
	router.POST("/host/reject/all", controllers.RejectAllHost)
	router.POST("/host/delete", controllers.DeleteHost)
	router.POST("/host/delete/all", controllers.DeleteAllHost)

	router.GET("/file_manage", controllers.FileManage)

	return router
}

func tokenVerify() gin.HandlerFunc {
	return func(c *gin.Context) {

		token := utils.Config().Http.Token

		reqToken := c.Request.Header["Http-Token"]

		if token == strings.Join(reqToken, " ") {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, "Invalid API token")
			c.Abort()
			return
		}
	}
}
