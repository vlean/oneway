package api

import (
	"gihub.com/vlean/oneway/model"
	"github.com/gin-gonic/gin"
)

type ForwardListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Name     string `form:"name"`
}

type ForwardListResp struct {
	Data []model.Forward `json:"data"`
}

func ForwardList(ctx *gin.Context) (data any, err error) {
	return nil, nil
}
