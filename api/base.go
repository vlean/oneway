package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handle func(ctx *gin.Context) (data any, err error)

type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

func WrapH(h Handle) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp := &Response{}
		data, err := h(ctx)
		if err != nil {
			resp.Code = 1
			resp.Msg = err.Error()
		} else {
			resp.Data = data
		}
		ctx.JSON(http.StatusOK, resp)
	}
}
