package api

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	// 静态资源
	// r.StaticFS("/ws", http.FS(FE))

	api := r.Group("/api", OAuthMiddleware)
	api.GET("/user", WrapH(Userinfo))
	// apiN := api.Group("/common")
	{
		// apiN.GET("/oauthurl", WrapH(OAuthUrl))
		// apiN.POST("/oauth", WrapH(OAuth))
	}
	apiFw := api.Group("/forward")
	{
		apiFw.GET("/", WrapH(ForwardList))
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
