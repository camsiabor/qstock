package main

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/agenda"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qdaobundle/qelastic"
	"github.com/camsiabor/qdaobundle/qredis"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/httpv"
	"github.com/camsiabor/qstock/run/rscript"
	"github.com/camsiabor/qstock/sync"
	"github.com/camsiabor/qstock/sync/stock"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"
)

func master(g *global.G) {

	go initPerfAnalysis(g)

	var jsonstr, _ = json.Marshal(g.Config)
	qlog.Log(qlog.INFO, "config", string(jsonstr[:]))

	var master_listen = util.GetStr(g.Config, "127.0.0.1:65000", "master", "listen")
	if !strings.Contains(master_listen, ":") {
		master_listen = ":" + master_listen
	}
	var lerr error
	g.Listener, lerr = net.Listen("tcp", master_listen)
	if lerr != nil {
		qlog.Log(qlog.INFO, "establish master listen fail ", master_listen, lerr)
		panic(lerr)
	}
	qlog.Log(qlog.INFO, "master", "listener establish", master_listen)

	var timezone = util.GetStr(g.Config, "Asia/Shanghai", "global", "timezone")
	time.LoadLocation(timezone)

	// [mapper] ------------------------------------------------------------------------------------------------
	var mapperConfig = util.GetMap(g.Config, true, "mapping")
	qref.GetMapperManager().Init(mapperConfig)
	g.SetData("mapperm", qref.GetMapperManager())

	// [dao] ------------------------------------------------------------------------------------------------
	initDao(g)

	// [script] ------------------------------------------------------------------------------------------------
	rscript.InitScript(g)

	// [gin] ------------------------------------------------------------------------------------------------
	httpv.GetInstance().Run()

	// [agenda] ------------------------------------------------------------------------------------------------
	var agendaConfig = util.GetMap(g.Config, true, "agenda")
	agenda.GetAgendaManager().Init(agendaConfig)

	initCacher(g)

	initSyncer(g)

}

func initPerfAnalysis(g *global.G) {
	var config = util.GetMap(g.Config, false, "debug", "http")
	if config == nil {
		return
	}
	var active = util.GetBool(config, false, "active")
	if !active {
		return
	}
	var endpoint = util.GetStr(config, ":8080", "endpoint")
	http.ListenAndServe(endpoint, nil)
}

func initDao(g *global.G) {
	var schemaOpts = util.GetMap(g.Config, true, "dbschema")
	var databaseOpts = util.GetMap(g.Config, true, "database")
	qdao.GetManager().Init(func(manager *qdao.DaoManager, name string, opts map[string]interface{}) (qdao.D, error) {
		var daotype = util.GetStr(opts, "", "type")
		if len(daotype) == 0 {
			panic(fmt.Errorf("no dao type specific in dao init : %v", opts))
		}
		switch daotype {
		case "redis":
			var redis = &qredis.DaoRedis{}
			redis.Framework = manager.GetSchema()
			return redis, nil
		case "elasticsearch":
			var elastic = &qelastic.DaoElastic{}
			return elastic, nil
		default:
			panic("db not implement : " + daotype)
		}
		return nil, nil
	}, schemaOpts, databaseOpts)
	g.SetData("daom", qdao.GetManager())
}

func initCacher(g *global.G) {

	var cache_manager = scache.GetManager()
	g.SetData("cachem", cache_manager)

	var cache_trade_calendar = cache_manager.Get(dict.CACHE_CALENDAR)
	cache_trade_calendar.Dao = dict.DAO_MAIN
	cache_trade_calendar.Db = dict.DB_CALENDAR
	cache_trade_calendar.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ...interface{}) (interface{}, error) {
		var dao, _ = qdao.GetManager().Get(scache.Dao)
		var date = keys[0]
		var r, err = dao.Get(scache.Db, scache.Group, date, 1, nil)
		var is_open = util.Get(r, 0, "is_open")
		return is_open, err
	}

	var cache_timestamp = cache_manager.Get(dict.CACHE_TIMESTAMP)
	cache_timestamp.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ...interface{}) (v interface{}, err error) {
		var key = keys[0]
		var skey = util.AsStr(key, "")
		if strings.Contains(skey, "@") {
			return qtime.Time2Int64(nil), nil
		}
		return time.Now().Format("20060102150405"), nil
	}

	var scache_code = cache_manager.Get(dict.CACHE_STOCK_CODE)
	scache_code.Dao = dict.DAO_MAIN
	scache_code.Db = dict.DB_DEFAULT
	scache_code.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ...interface{}) (interface{}, error) {
		var dao, err = qdao.GetManager().Get(scache.Dao)
		if err != nil {
			qlog.Error(0, err)
			go func() {
				defer qerr.SimpleRecover(1)
				time.Sleep(time.Duration(10) * time.Second)
				scache.Loader(scache, factor, timeout, keys...)
			}()
			return nil, err
		}
		var codes, _ = dao.Keys(scache.Db, "", "*", nil)
		var sz, szn = make([]string, 5000), 0
		var sh, shn = make([]string, 5000), 0
		var su, sun = make([]string, 5000), 0
		var all, alln = make([]string, 15000), 0
		for _, code := range codes {
			var include = true
			switch code[0] {
			case '0':
				sz[szn] = code
				szn++
			case '3':
				su[szn] = code
				sun++
			case '6':
				sh[shn] = code
				shn++
			default:
				include = false
			}
			if include {
				all[alln] = code
				alln++
			}
		}

		var sz_sh = make([]string, 0)
		sz_sh = append(sz_sh, sz[:szn]...)
		sz_sh = append(sz_sh, sh[:shn]...)

		scache.Set(all[:alln], "all")
		scache.Set(sz[:szn], dict.SHENZHEN)
		scache.Set(sh[:shn], dict.SHANGHAI)
		scache.Set(su[:sun], dict.STARTUP)
		scache.Set(sz_sh, dict.SHENZHEN+"."+dict.SHANGHAI)
		qlog.Log(qlog.INFO, "cache", "code", "shenzhen", szn, "shanghai", shn, "startup", sun, "all", alln)
		return scache, nil
	}
	scache_code.Loader(scache_code, 1, 0)

	var scache_snapshot = scache.GetManager().Get(dict.CACHE_STOCK_SNAPSHOT)
	scache_snapshot.Dao = dict.DAO_MAIN
	scache_snapshot.Db = dict.DB_DEFAULT
	scache_snapshot.Loader = func(scache *scache.SCache, factor int, timeout time.Duration, keys ...interface{}) (interface{}, error) {
		conn, err := qdao.GetManager().Get(scache.Dao)
		if err != nil {
			return nil, err
		}
		var code = keys[0]
		return conn.Get(scache.Db, "", code, 1, nil)
	}

	var cache_khistory_loader_generator = func(prefix string, profilesuffix string) scache.Loader {
		return func(scache *scache.SCache, factor int, timeout time.Duration, keys ...interface{}) (interface{}, error) {
			if len(keys) < 1 {
				return nil, fmt.Errorf("keys len invalid for this cache Loader %s", scache.Name)
			}
			conn, err := qdao.GetManager().Get(scache.Dao)
			if err != nil {
				return nil, err
			}

			var code string
			if len(prefix) == 0 {
				code = util.AsStr(keys[0], "")
			} else {
				var skey = util.AsStr(keys[0], "")
				code = prefix + skey
			}
			var datestr = keys[1]
			data, err := conn.Get(scache.Db, code, datestr, 1, nil)
			if data != nil {
				return data, err
			}
			if factor <= 0 {
				return data, err
			}
			var g = global.GetInstance()
			var cmd = &global.Cmd{
				Service:  dict.SERVICE_SYNC,
				Function: "k.history" + profilesuffix,
				SFlag:    "go",
			}
			cmd.SetData("from", datestr)
			cmd.SetData("codes", []string{code})
			var reply, _ = g.SendCmd(cmd, timeout)
			if reply != nil {
				err = reply.RetErr
				data = reply.RetVal
			}
			return data, err
		}
	}

	var scache_khistory = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY)
	scache_khistory.Dao = dict.DAO_MAIN
	scache_khistory.Db = dict.DB_HISTORY
	scache_khistory.Timeout = -1 //time.Second * time.Duration(20);
	scache_khistory.Loader = cache_khistory_loader_generator("", "")

	var cache_stock_khistory_week = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY_WEEK)
	cache_stock_khistory_week.Dao = dict.DAO_MAIN
	cache_stock_khistory_week.Db = dict.DB_HISTORY
	cache_stock_khistory_week.Timeout = -1 //time.Second * time.Duration(20);
	cache_stock_khistory_week.Loader = cache_khistory_loader_generator("w.", "week")

	var cache_stock_khistory_month = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY_MONTH)
	cache_stock_khistory_month.Dao = dict.DAO_MAIN
	cache_stock_khistory_month.Db = dict.DB_HISTORY
	cache_stock_khistory_month.Timeout = -1 //time.Second * time.Duration(20);
	cache_stock_khistory_month.Loader = cache_khistory_loader_generator("m.", "month")

	g.SetData("calendar", stock.GetStockCalendar())

}

func initSyncer(g *global.G) {
	var api_config = util.GetMap(g.Config, true, "api")
	for apiname, api := range api_config {
		var active = util.GetBool(api, true, "active")
		if !active {
			continue
		}
		var syncer = new(sync.Syncer)
		g.CmdHandlerRegister(dict.SERVICE_SYNC, syncer)
		syncer.Run(apiname)
	}
}
