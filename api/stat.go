package api

import (
	"time"

	"gihub.com/vlean/oneway/netx"
	"gihub.com/vlean/oneway/tool/stat"
	"github.com/gin-gonic/gin"
)

func Stat(ctx *gin.Context) (data any, err error) {
	gp := netx.GlobalGP()
	d := make(map[string]any)
	rt := stat.Runtime()
	d["client"] = gp.Stat()
	d["http"] = rt.Http
	d["start_at"] = rt.StartAt.Format("2006-01-02 15:04:05")
	d["run_time"] = time.Since(rt.StartAt).Seconds()
	data = d
	return
}
