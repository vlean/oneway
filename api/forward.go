package api

import (
	"fmt"

	"gihub.com/vlean/oneway/model"
	"github.com/gin-gonic/gin"
)

type ForwardListReq struct {
	Page     int    `form:"current"`
	PageSize int    `form:"pageSize"`
	Name     string `form:"keyword"`
}

type ForwardListResp struct {
	Data []model.Forward `json:"data"`
}

func Bind[T any](ctx *gin.Context) (req T, err error) {
	err = ctx.ShouldBind(&req)
	if err != nil {
		return
	}
	return
}

func ForwardList(ctx *gin.Context) (data any, err error) {
	req, err := Bind[ForwardListReq](ctx)
	if err != nil {
		return
	}
	fmt.Println(req)
	data = model.NewForwardDao().FindAll()

	fmt.Println(data)
	return
}

func ForwardSave(ctx *gin.Context) (data any, err error) {
	req, err := Bind[model.Forward](ctx)
	if err != nil {
		return
	}

	err = model.NewForwardDao().Save(&req)
	return req, err
}
