package httpv

import (
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/run/rscript"
	"strings"

	"github.com/gin-gonic/gin"
)

var cacheScriptByName *scache.SCache
var cacheScriptByHash *scache.SCache

func (o *HttpServer) routeScript() {
	var group = o.engine.Group("/script")
	cacheScriptByName = scache.GetManager().Get(dict.CACHE_SCRIPT_BY_NAME)
	cacheScriptByHash = scache.GetManager().Get(dict.CACHE_SCRIPT_BY_HASH)

	group.POST("/update", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var meta = &rscript.Meta{}
		var err = meta.FromMap(m)
		if err != nil {
			o.RespJsonEx(nil, err, c)
			return
		}
		err = cacheScriptByName.Set(meta, name)
		//var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		//var data, err = dao.Update(dict.DB_COMMON, "script", name, m, true, -1, nil)
		o.RespJsonEx("done", err, c)
	})

	group.POST("/list", func(c *gin.Context) {
		var err error
		var data interface{}
		var m, _ = o.ReqParse(c)
		var stype = util.GetStr(m, "script", "type")
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)

		var datamap map[string]interface{}
		var multiple = strings.Index(stype, ",") >= 0
		if multiple {
			datamap = make(map[string]interface{})
		}

		if strings.Index(stype, "script") >= 0 {
			data, err = dao.Keys(dict.DB_COMMON, "script", "*", nil)
			if multiple {
				datamap["script_names"] = data
			}
		}

		if err == nil && strings.Index(stype, "group") >= 0 {
			data, err = qdao.ListAll(dao, dict.DB_COMMON, "script_group", 0, 256, 1, nil)
			if multiple {
				datamap["script_group"] = data
			}
		}

		if multiple {
			data = datamap
		}

		o.RespJsonEx(data, err, c)
	})

	group.POST("/get", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var data, err = cacheScriptByName.Get(true, name)
		if data != nil {
			var meta = data.(*rscript.Meta)
			data = meta.ToMap()
		}
		//var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		//var data, err = dao.Get(dict.DB_COMMON, "script", name, 1, nil)
		o.RespJsonEx(data, err, c)
	})

	group.POST("/delete", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var name = util.GetStr(m, "", "name")
		var err = cacheScriptByName.Delete(name, true)
		//var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		//var data, err = dao.Delete(dict.DB_COMMON, "script", name, nil)
		o.RespJsonEx("done", err, c)
	})

}
