package api

import "github.com/gin-gonic/gin"

func Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/user", WrapH(Userinfo))
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
	}

}
