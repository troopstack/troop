package http

import (
	"net/http"
	"strings"

	"github.com/troopstack/troop/src/modules/file/http/controllers"
	"github.com/troopstack/troop/src/modules/file/utils"

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

	router.POST("/file/upload", controllers.FileUpload)
	router.POST("/file/remove", controllers.FileRemove)
	router.POST("/plugin/upload", controllers.PluginUpload)
	router.StaticFS("/file/download", http.Dir(utils.FileRoot))

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
