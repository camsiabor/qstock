package httpv

import (
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

func (o *HttpServer) routeStock() {
	var group = o.Engine.Group("/stock")

	group.POST("/sync", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var profileName = util.GetStr(m, "", "profile")
		var cmd = util.GetStr(m, "force,record", "cmd")
		var _, err = global.GetInstance().SendCmd(&global.Cmd{
			Service:  dict.SERVICE_SYNC,
			Function: profileName,
			SFlag:    cmd,
		}, 0)
		o.RespJsonEx("cmd sent", err, c)
	})

	group.POST("/clear", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var db = util.GetStr(m, "", "db")
		var group = util.GetStr(m, "", "group")
		var daoname = util.GetStr(m, dict.DAO_MAIN, "dao")
		dao, err := qdao.GetManager().Get(daoname)
		if err != nil {
			o.RespJsonEx(nil, err, c)
			return
		}
		keys, err := dao.Keys(db, group, "*", nil)
		if err != nil {
			o.RespJsonEx(nil, err, c)
			return
		}
		var ids = util.AsSlice(keys, 0)
		ret, err := dao.Deletes(db, group, ids, nil)
		o.RespJsonEx(ret, err, c)
	})

	group.POST("/keys", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var dbs = util.GetSlice(m, "dbs")
		var group = util.GetStr(m, "", "group")
		var keys = util.GetStr(m, "meta*", "keys")
		var ret = make(map[string]interface{})
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)
		for _, db := range dbs {
			var sdb = db.(string)
			keysret, err := dao.Keys(sdb, group, keys, nil)
			if err != nil {
				o.RespJsonEx(nil, err, c)
				return
			}
			for _, key := range keysret {
				var oneret, oneerr = dao.Get(sdb, group, key, 1, nil)
				if oneerr == nil {
					ret[key] = oneret
				} else {
					ret[key] = oneerr.Error()
				}
			}
		}
		o.RespJsonEx(ret, nil, c)
	})

	group.POST("/gets", func(c *gin.Context) {
		var m, _ = o.ReqParse(c)
		var ofetchs = util.Get(m, nil, "fetchs")
		var time_from_str = util.GetStr(m, "", "time_from")
		var time_to_str = util.GetStr(m, "", "time_to")

		var index = 0
		var fetchs = util.AsSlice(ofetchs, 0)
		var cacher_stock_snapshot = scache.GetManager().Get(dict.CACHE_STOCK_SNAPSHOT)
		var cacher_stock_khistory = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY)
		if fetchs == nil || len(fetchs) == 0 {
			o.RespJsonEx(nil, errors.New("fetchs is null"), c)
			return
		}

		var time_array_len = 0
		var time_array []string = nil
		var time_n_array []int = nil
		if len(time_from_str) > 0 {
			if len(time_to_str) <= 0 {
				var now = time.Now()
				time_to_str = qtime.YYYY_MM_dd(&now)
			}
			time_from, perr := time.Parse("20060102", time_from_str)
			if perr != nil {
				o.RespJsonEx(nil, perr, c)
				return
			}
			time_to, perr := time.Parse("20060102", time_to_str)
			if perr != nil {
				o.RespJsonEx(nil, perr, c)
				return
			}

			times, perr := qtime.GetTimeFormatIntervalArray(&time_from, &time_to, "20060102", time.Saturday, time.Sunday)
			if perr != nil {
				o.RespJsonEx(nil, perr, c)
				return
			}
			time_array = times
			time_array_len = len(time_array)
			time_n_array = make([]int, time_array_len)
			for n := 0; n < time_array_len; n++ {
				time_n_array[n], _ = strconv.Atoi(time_array[n])
			}
		}

		var data = make([]interface{}, len(fetchs))
		for _, fetch := range fetchs {
			if fetch == nil {
				continue
			}

			var scode, ok = fetch.(string)
			if !ok {
				var mfetch = util.AsMap(fetch, false)
				if mfetch == nil {
					continue
				}
				scode = util.GetStr(mfetch, "", "code")
			}

			snapshoto, err := cacher_stock_snapshot.Get(true, scode)
			snapshot := util.AsMap(snapshoto, false)
			if err != nil {
				o.RespJsonEx(0, err, c)
				return
			}

			if snapshot != nil {
				data[index] = snapshot
				index = index + 1
			}

			if time_array_len <= 0 {
				continue
			}

			var fetch_from_s = util.GetStr(fetch, "", "from")
			if fetch_from_s == "x" {
				continue
			}

			var khistory_subcache = cacher_stock_khistory.GetSub(scode)
			var fetch_from_index, fetch_to_index int = -1, -1
			var fetch_from = util.GetInt(fetch, 0, "from")
			var fetch_to = util.GetInt(fetch, 0, "to")
			if fetch_from > 0 {
				if fetch_to <= 0 {
					fetch_to = time_n_array[time_array_len-1]
				}
				for n := 0; n < time_array_len; n++ {
					var time_n = time_n_array[n]
					if fetch_from_index < 0 {
						if fetch_from == time_n && fetch_from >= time_n {
							fetch_from_index = n
						}
					}
					if fetch_from_index >= 0 && fetch_to >= time_n {
						fetch_to_index = n - 1
						break
					}
				}
			}
			if fetch_from_index <= 0 {
				fetch_from_index = 0
			}
			if fetch_to_index <= 0 {
				fetch_to_index = time_array_len - 1
			}
			//fetch_from = time_n_array[fetch_from_index];
			//fetch_to = time_n_array[fetch_to_index];
			var time_include = time_array[fetch_from_index : fetch_to_index+1]
			if len(time_include) > 0 {
				khistory, err := khistory_subcache.List(true, time_include...)
				if err != nil {
					o.RespJsonEx(0, err, c)
					return
				}
				snapshot["khistory"] = khistory
				snapshot["khistory_to"] = fetch_to
				snapshot["khistory_from"] = fetch_from
			}
		}

		o.RespJson(0, data[:index], c)
	})

}
