package api

import "github.com/gin-gonic/gin"

func Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/forward", WrapH(ForwardList))
		api.POST("/forward/save", WrapH(ForwardSave))
	}
}
