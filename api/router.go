package api

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	r.GET("/auth", WrapH(Auth))
	r.POST("/code", Code)
	r.POST("/token", Code)

	api := r.Group("/api", OAuthMiddleware)
	api.GET("/user", WrapH(Userinfo))

	apiStat := api.Group("/stat")
	{
		apiStat.GET("/info", WrapH(Stat))
	}

	apiFw := api.Group("/forward")
	{
		apiFw.GET("/list", WrapH(ForwardList))
		apiFw.POST("/save", WrapH(ForwardSave))
		apiFw.POST("/delete", WrapH(ForwardDelete))
	}

	apiSys := api.Group("/system")
	{
		apiSys.GET("/config", WrapH(SystemConfig))
		apiSys.POST("/config/update", WrapH(SystemConfigUpdate))
	}
}

func OAuthMiddleware(ctx *gin.Context) {
	// 全局登录校验
	ctx.Next()
}
