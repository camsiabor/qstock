package httpv

import (
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/run/rscript"

	"github.com/gin-gonic/gin"
)

func (o *HttpServer) routeScript() {
	var group = o.Engine.Group("/script")

	var cache_script_by_name = scache.GetManager().Get(dict.CACHE_SCRIPT_BY_NAME)

	group.POST("/update", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var meta = &rscript.Meta{}
		var err = meta.FromMap(m)
		if err != nil {
			o.RespJsonEx(nil, err, c)
			return
		}
		err = cache_script_by_name.Set(meta, name)
		//var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		//var data, err = dao.Update(dict.DB_COMMON, "script", name, m, true, -1, nil)
		o.RespJsonEx("done", err, c)
	})

	group.POST("/list", func(c *gin.Context) {
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		var scripts, err = dao.Keys(dict.DB_COMMON, "script", "*", nil)
		o.RespJsonEx(scripts, err, c)
	})

	group.POST("/get", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var data, err = cache_script_by_name.Get(true, name)
		//var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		//var data, err = dao.Get(dict.DB_COMMON, "script", name, 1, nil)
		o.RespJsonEx(data, err, c)
	})

	group.POST("/delete", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var err = cache_script_by_name.Delete(name)
		//var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		//var data, err = dao.Delete(dict.DB_COMMON, "script", name, nil)
		o.RespJsonEx("done", err, c)
	})
}
